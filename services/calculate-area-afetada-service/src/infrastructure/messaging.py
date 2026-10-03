import asyncio
import json
from datetime import datetime, timezone
import aio_pika
from aio_pika.abc import AbstractChannel, AbstractExchange, AbstractQueue
from psycopg import AsyncConnection
from application.use_cases import HANDLERS
from core.rules import AreaRules
from infrastructure import database

SOURCE = "calculate-area-service"
OUTBOX_BATCH_SIZE = 50

# Cria (se não existir) a exchange, a fila deste serviço e liga a fila nos eventos que ele trata
async def setup(
    channel: AbstractChannel, exchange_name: str, queue_name: str, routing_keys: list[str]
) -> tuple[AbstractExchange, AbstractQueue]:
    exchange = await channel.declare_exchange(exchange_name, aio_pika.ExchangeType.TOPIC, durable=True)
    queue = await channel.declare_queue(queue_name, durable=True)
    for routing_key in routing_keys:
        await queue.bind(exchange, routing_key=routing_key)
    return exchange, queue

# Só recebe a mensagem e guarda no Inbox. Quem processa é o Inbox Consumer.
async def run_event_consumer(conn: AsyncConnection, queue: AbstractQueue) -> None:
    print(f"Event Consumer escutando a fila {queue.name}")
    async with queue.iterator() as messages:
        async for message in messages:
            try:
                body = json.loads(message.body)
                event_id = str(body["id"])
                event_type = body["type"]
                data = body["data"]
            except (ValueError, KeyError, TypeError):
                print(f"Mensagem fora do formato, descartada: {message.body[:200]}")
                await message.reject()
                continue

            async with conn.transaction():
                is_new = await database.inbox_add(conn, event_id, event_type, data)
            if not is_new:
                print(f"Evento {event_id} repetido, ignorado")

            # Só confirma para o RabbitMQ depois de salvo no banco
            await message.ack()


async def run_inbox_consumer(conn: AsyncConnection, rules: AreaRules, poll_interval_s: float) -> None:
    print("Inbox Consumer iniciado")
    while True:
        async with conn.transaction():
            event = await database.inbox_fetch_next_pending(conn)
            if event is not None:
                await _process(conn, event, rules)

        if event is None:
            await asyncio.sleep(poll_interval_s)

async def _process(conn: AsyncConnection, event: dict, rules: AreaRules) -> None:
    event_id, event_type = event["event_id"], event["event_type"]
    handler = HANDLERS.get(event_type)
    if handler is None:
        await database.inbox_mark_rejected(conn, event_id, f"Tipo de evento desconhecido: {event_type}")
        return

    try:
        # Transação interna: se der erro, desfaz só o processamento e mantém o registro do erro
        async with conn.transaction():
            await handler(conn, event["payload"], rules, datetime.now(timezone.utc))
        await database.inbox_mark_processed(conn, event_id)
        print(f"Evento {event_id}, {event_type} processado")
    except (KeyError, ValueError, TypeError) as error:
        # Dado inválido: tentar de novo não resolve
        print(f"Evento {event_id} rejeitado: {error}")
        await database.inbox_mark_rejected(conn, event_id, str(error))
    except Exception as error:
        print(f"Erro ao processar o evento {event_id}")
        await database.inbox_mark_failed(conn, event_id, str(error))


async def run_outbox_dispatcher(
    conn: AsyncConnection, exchange: AbstractExchange, poll_interval_s: float
) -> None:
    print("Outbox Dispatcher iniciado")
    while True:
        # Se a publicação falhar no meio, a transação volta e o lote é reenviado.
        # Pode gerar mensagem duplicada, por isso quem consome usa o id no Inbox.
        async with conn.transaction():
            events = await database.outbox_fetch_unpublished(conn, OUTBOX_BATCH_SIZE)
            for event in events:
                await exchange.publish(_to_message(event), routing_key=event["event_type"])
                await database.outbox_mark_published(conn, event["id"])
                print(f"Evento {event["id"]} publicado {event["event_type"]}")

        if not events:
            await asyncio.sleep(poll_interval_s)


# Envelope no formato CloudEvents
def _to_message(event: dict) -> aio_pika.Message:
    body = {
        "specversion": "1.0",
        "id": str(event["id"]),
        "source": SOURCE,
        "type": event["event_type"],
        "time": event["created_at"].isoformat(),
        "datacontenttype": "application/json",
        "data": event["payload"],
    }
    return aio_pika.Message(
        body=json.dumps(body).encode(),
        content_type="application/json",
        message_id=str(event["id"]),
        type=event["event_type"],
        delivery_mode=aio_pika.DeliveryMode.PERSISTENT,
    )
