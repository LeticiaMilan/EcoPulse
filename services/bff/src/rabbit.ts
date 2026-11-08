import amqp from 'amqplib';
import { randomUUID } from 'crypto';

const URL = process.env.RABBITMQ_URL ?? 'amqp://ecopulse:ecopulse@127.0.0.1:5672/';
const EXCHANGE = 'exchange.ecopulse';

let channel: amqp.ConfirmChannel | null = null;

export async function connectRabbit() {
    for (let attempt = 1; ; attempt++) {
        try {
            const conn = await amqp.connect(URL);
            channel = await conn.createConfirmChannel();
            await channel.assertExchange(EXCHANGE, 'topic', { durable: true });

            // Fila só para testes: recebe todos os eventos, para ver no painel
            if (process.env.DEBUG_QUEUE === '1') {
                await channel.assertQueue('queue.bff.debug', { durable: true });
                await channel.bindQueue('queue.bff.debug', EXCHANGE, '#');
            }

            conn.on('close', () => { channel = null; console.log('RabbitMQ caiu, reconectando...'); setTimeout(connectRabbit, 2000); });
            console.log('RabbitMQ conectado');
            return;
        } catch (err: any) {
            const wait = Math.min(30, 2 ** attempt);
            console.log(`RabbitMQ indisponível (${err.message}), tentando de novo em ${wait}s`);
            await new Promise(r => setTimeout(r, wait * 1000));
        }
    }
}

// Publica no formato combinado no RabbitMq_documentation.md
export async function publish(routingKey: string, data: unknown) {
    if (!channel) throw new Error('RABBIT_DOWN');
    const envelope = { id: randomUUID(), source: 'SERVICE.FRONTEND', contentType: 'application/json', data };
    channel.publish(EXCHANGE, routingKey, Buffer.from(JSON.stringify(envelope)), {
        persistent: true,
        contentType: 'application/json',
        messageId: envelope.id,
    });
    await channel.waitForConfirms(); // garante que o broker recebeu
}