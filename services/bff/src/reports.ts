import { Router } from 'express';
import { requireAuth, AuthUser } from './auth';
import { publish } from './rabbit';

export const reports = Router();

const TYPES = ['Flooding', 'Thunderstorm', 'Fire Disaster', 'Landslide'];
const SEVERITIES = ['LOW', 'MEDIUM', 'HIGH'];
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

// guarda as respostas já dadas em memória
const seen = new Map<string, unknown>();

const fail = (res: any, status: number, code: string, message: string) =>
    res.status(status).json({ error: { code, message } });

reports.post('/reports', requireAuth, async (req, res) => {
    const user = res.locals.user as AuthUser;
    const b = req.body ?? {};
    const key = String(req.headers['idempotency-key'] ?? b.id ?? '');
    if (seen.has(key)) return res.status(202).json(seen.get(key));

    if (!UUID.test(b.id ?? '')) return fail(res, 400, 'INVALID_ID', 'id deve ser UUID');
    if (!TYPES.includes(b.type)) return fail(res, 400, 'INVALID_TYPE', 'Tipo de relato inválido');
    if (!SEVERITIES.includes(b.severity)) return fail(res, 400, 'INVALID_SEVERITY', 'Gravidade inválida');
    if (typeof b.description !== 'string' || !b.description.trim() || b.description.length > 280) return fail(res, 400, 'INVALID_DESCRIPTION', 'Descrição deve ter de 1 a 280 caracteres');
    if (typeof b.latitude !== 'number' || Math.abs(b.latitude) > 90) return fail(res, 400, 'INVALID_COORDS', 'Latitude inválida');
    if (typeof b.longitude !== 'number' || Math.abs(b.longitude) > 180) return fail(res, 400, 'INVALID_COORDS', 'Longitude inválida');
    const datetime = b.datetime ?? new Date().toISOString();

    try {
        await publish('EVENTS.REPORT.GENERATED', {
            id: b.id, latitude: b.latitude, longitude: b.longitude, type: b.type,
            user_id: user.id, area: b.address ?? '', datetime,
            description: b.description, severity: b.severity,
        });
        await publish('EVENTS.VALIDATION.CREATE', {
            id: b.id, usuario_id: user.id, tipo: b.type, descricao: b.description,
            latitude: b.latitude, longitude: b.longitude,
        });
    } catch {
        return fail(res, 503, 'BROKER_UNAVAILABLE', 'Serviço temporariamente indisponível');
    }

    const out = { id: b.id, status: 'POSSIVEL_RELATO', accepted_at: new Date().toISOString() };
    seen.set(key, out);
    res.status(202).json(out);
});

reports.post('/reports/:id/vote', requireAuth, async (req, res) => {
    const user = res.locals.user as AuthUser;
    const id = String(req.params.id);
    if (!UUID.test(id)) return fail(res, 400, 'INVALID_ID', 'id deve ser UUID');
    if (typeof req.body?.positive !== 'boolean') return fail(res, 400, 'INVALID_VOTE', 'positive deve ser true ou false');

    const key = `vote:${id}:${user.id}`;
    if (seen.has(key)) return fail(res, 409, 'ALREADY_VOTED', 'Você já votou neste relato');

    try {
        await publish('EVENTS.VALIDATION.VOTE', { id, tipo: req.body.positive ? 'positivo' : 'negativo', usuario_id: user.id });
    } catch {
        return fail(res, 503, 'BROKER_UNAVAILABLE', 'Serviço temporariamente indisponível');
    }

    const out = { report_id: id, positive: req.body.positive, accepted_at: new Date().toISOString() };
    seen.set(key, out);
    res.status(202).json(out);
});