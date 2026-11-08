import express from 'express';
import { requireAuth } from './auth';
import { connectRabbit } from './rabbit';
import { reports } from './reports';

const app = express();
app.use(express.json());

app.get('/health', (_req, res) => res.json({ status: 'ok', service: 'bff' }));
app.get('/me', requireAuth, (_req, res) => res.json(res.locals.user));
app.use(reports);

const PORT = Number(process.env.PORT ?? 8090);
app.listen(PORT, '0.0.0.0', () => console.log(`BFF rodando na porta ${PORT}`));
connectRabbit();