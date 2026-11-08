# Link para acessar o site do rabbtmq: http://localhost:15672
# Usuário e senha para acesso da página do rabbtmq: ecopulse

'''Exemplos para a comunicação:

Mensagem de entrada:
{
  "id": "3a8d1e05-6f2c-4b91-8840-d9e2358b1c47",
  "source": "SERVICE.REPORT",
  "contentType": "application/json",
  "data": {
    "id": "8f62b19e-5c40-42ae-bb73-936d5e1a7b80",
    "user_id": "c19b43d7-e218-450f-a396-745812fe69b2",
    "type": "Flooding",
    "latitude": -21.1775,
    "longitude": -47.8103,
    "datetime": "2026-10-31T15:30:00Z"
  }
}

Mensagem de Saída:
{
  "id": "2f76aaed-109a-4542-9734-2eb4a20c8cd3",
  "source": "SERVICE.AFFECTED_AREA",
  "contentType": "application/json",
  "data": {
    "area_id": "45f229ff-1127-4dbc-9b5b-318f8eccc9eb",
    "type": "Flooding",
    "status": "active",
    "area_m2": 58598.47,
    "center": {
      "latitude": -21.177542928654507,
      "longitude": -47.81039045186131
    },
    "polygon": {
      "type": "Polygon",
      "coordinates": [
        [
          [-47.80915977255845, -21.1775609594799],
          [-47.809118275766075, -21.17763871413161],
          [-47.80908501840699, -21.177719844912232],
          [-47.80915977255845, -21.1775609594799]
        ]
      ]
    },
    "reports": 3,
    "report_ids": [
      "8b1f6c2e-3a4d-4e5f-9a7b-1c2d3e4f5a6b",
      "c3d4e5f6-7a8b-4c9d-8e0f-1a2b3c4d5e6f",
      "f1e2d3c4-b5a6-4978-8a9b-0c1d2e3f4a5b"
    ],
    "users": 3
  }
}

'''

import asyncio
import os
import aio_pika
from application.use_cases import HANDLERS, run_scheduler
from core.rules import AreaRules
from infrastructure import database, messaging

# Os valores padrão servem para rodar local com o docker-compose deste serviço
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://ecopulse:ecopulse@localhost:5432/calculate_area")
RABBITMQ_URL = os.getenv("RABBITMQ_URL", "amqp://ecopulse:ecopulse@localhost:5672/")
RABBITMQ_EXCHANGE = os.getenv("RABBITMQ_EXCHANGE", "exchange.ecopulse")
RABBITMQ_QUEUE = os.getenv("RABBITMQ_QUEUE", "queue.service.affected-area")
POLL_INTERVAL_S = float(os.getenv("POLL_INTERVAL_SECONDS", "1"))

async def main() -> None:
    rules = AreaRules()
    setup_conn = await database.connect(DATABASE_URL)
    await database.create_tables(setup_conn)
    await setup_conn.close()
    rabbit = await aio_pika.connect_robust(RABBITMQ_URL)
    async with rabbit:
        channel = await rabbit.channel()
        await channel.set_qos(prefetch_count=10)
        exchange, queue = await messaging.setup(channel, RABBITMQ_EXCHANGE, RABBITMQ_QUEUE, list(HANDLERS))

        # Cada parte do diagrama roda ao mesmo tempo, com sua própria conexão no banco
        await asyncio.gather(
            messaging.run_event_consumer(await database.connect(DATABASE_URL), queue),
            messaging.run_inbox_consumer(await database.connect(DATABASE_URL), rules, POLL_INTERVAL_S),
            messaging.run_outbox_dispatcher(await database.connect(DATABASE_URL), exchange, POLL_INTERVAL_S),
            run_scheduler(await database.connect(DATABASE_URL), rules, POLL_INTERVAL_S),
        )


print("Iniciando calculate-area-service")
asyncio.run(main(), loop_factory=asyncio.SelectorEventLoop)
