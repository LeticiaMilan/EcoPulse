# Link para acessar o site do rabbtmq: http://localhost:15672
# Usuário e senha para acesso da página do rabbtmq: ecopulse
import asyncio
import os
import aio_pika
from application.use_cases import HANDLERS, run_scheduler
from core.rules import AreaRules
from infrastructure import database, messaging

# Os valores padrão servem para rodar local com o docker-compose deste serviço
DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://ecopulse:ecopulse@localhost:5432/calculate_area")
RABBITMQ_URL = os.getenv("RABBITMQ_URL", "amqp://ecopulse:ecopulse@localhost:5672/")
RABBITMQ_EXCHANGE = os.getenv("RABBITMQ_EXCHANGE", "ecopulse.events")
RABBITMQ_QUEUE = os.getenv("RABBITMQ_QUEUE", "calculate-area-service")
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
