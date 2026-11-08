import { Router } from 'express';
import { requireAuth, optionalAuth, AuthUser } from './auth';
import { publish } from './rabbit';

export const reports = Router();

const TYPES = ['Flooding', 'Thunderstorm', 'Fire Disaster', 'Landslide'];
const SEVERITIES = ['LOW', 'MEDIUM', 'HIGH'];
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const NEEDED = 5;
const REPORT_URL = process.env.REPORT_SERVICE_URL ?? '';
const VALIDATION_URL = process.env.VALIDATION_SERVICE_URL ?? '';  

// respostas já dadas
const seen = new Map<string, unknown>();

// cópia local dos relatos
type StoredReport = {
    id: string; type: string; status: string; severity: string; description: string;
    address: string; latitude: number; longitude: number; user_id: string; username: string;
    datetime: string; updated_at: string; votes: { positive: number; negative: number };
};
const store = new Map<string, StoredReport>();
const myVotes = new Map<string, boolean>(); // chave: "reportId:userId"

const fail = (res: any, status: number, code: string, message: string) =>
    res.status(status).json({ error: { code, message } });

function view(r: StoredReport, userId?: string) {
    const mine = userId ? myVotes.get(`${r.id}:${userId}`) : undefined;
    return { ...r, votes: { ...r.votes, needed: NEEDED, my_vote: mine ?? null }, area: null };
}

// ---------- POST /reports ----------
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
    } catch {
        return fail(res, 503, 'BROKER_UNAVAILABLE', 'Serviço temporariamente indisponível');
    }

    const now = new Date().toISOString();
    store.set(b.id, {
        id: b.id, type: b.type, status: 'POSSIVEL_RELATO', severity: b.severity, description: b.description,
        address: b.address ?? '', latitude: b.latitude, longitude: b.longitude,
        user_id: user.id, username: user.username, datetime, updated_at: now,
        votes: { positive: 0, negative: 0 },
    });

    const out = { id: b.id, status: 'POSSIVEL_RELATO', accepted_at: now };
    seen.set(key, out);
    res.status(202).json(out);
});

// ---------- POST /reports/:id/vote ----------
reports.post('/reports/:id/vote', requireAuth, async (req, res) => {
    const user = res.locals.user as AuthUser;
    const id = String(req.params.id);
    if (!UUID.test(id)) return fail(res, 400, 'INVALID_ID', 'id deve ser UUID');
    if (typeof req.body?.positive !== 'boolean') return fail(res, 400, 'INVALID_VOTE', 'positive deve ser true ou false');
    const positive = req.body.positive as boolean;

    if (VALIDATION_URL) {
        try {
            const check = await fetch(`${VALIDATION_URL}/validacaorelatos/votou/${user.id}/${id}`, { signal: AbortSignal.timeout(3000) });
            if (check.ok && (await check.json()).votou) return fail(res, 409, 'ALREADY_VOTED', 'Você já votou neste relato');

            const r = await fetch(`${VALIDATION_URL}/validacaorelatos/votar/${id}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ positivo: positive, usuario_id: user.id }),
                signal: AbortSignal.timeout(3000),
            });
            if (r.status === 404) return fail(res, 404, 'NOT_FOUND', 'Relato não encontrado');
            if (!r.ok) {
                const msg = await r.text();
                if (/vot/i.test(msg)) return fail(res, 409, 'ALREADY_VOTED', 'Você já votou neste relato');
                return fail(res, 502, 'VALIDATION_ERROR', msg || 'Erro no serviço de validação');
            }
        } catch {
            return fail(res, 503, 'VALIDATION_UNAVAILABLE', 'Serviço de validação indisponível');
        }
    } else {
        if (myVotes.has(`${id}:${user.id}`)) return fail(res, 409, 'ALREADY_VOTED', 'Você já votou neste relato');
    }

    myVotes.set(`${id}:${user.id}`, positive);
    const local = store.get(id);
    if (local) {
        if (positive) local.votes.positive++; else local.votes.negative++;
        if (local.status === 'POSSIVEL_RELATO' && local.votes.positive >= NEEDED) local.status = 'RELATO_VALIDADO';
        local.updated_at = new Date().toISOString();
    }

    res.status(202).json({ report_id: id, positive, accepted_at: new Date().toISOString() });
});

// ---------- GET /reports ----------
async function fromReportService(since: string, limit: number) {
    if (!REPORT_URL) return null;
    try {
        const q = new URLSearchParams({ limit: String(limit), ...(since ? { since } : {}) });
        const res = await fetch(`${REPORT_URL}/report/api/sync?${q}`, { signal: AbortSignal.timeout(3000) });
        if (!res.ok) return null;
        const body = await res.json();
        return (body.data ?? []) as any[];
    } catch {
        return null; // serviço fora: usa a cópia local
    }
}

reports.get('/reports', optionalAuth, async (req, res) => {
    const user = res.locals.user as AuthUser | null;
    const since = String(req.query.since ?? '');
    const limit = Math.min(100, Math.max(1, Number(req.query.limit ?? 100)));

    const remote = await fromReportService(since, limit);
    if (remote) {
        const data = remote.map(d => {
            const local = store.get(d.id);
            return {
                id: d.id, type: d.type, status: d.status,
                severity: local?.severity ?? 'MEDIUM', description: local?.description ?? '',
                address: local?.address ?? '', latitude: d.latitude, longitude: d.longitude,
                user_id: d.user_id, username: local?.username ?? '', datetime: d.datetime,
                updated_at: d.area?.updated_at ?? d.datetime,
                votes: { ...(local?.votes ?? { positive: 0, negative: 0 }), needed: NEEDED, my_vote: user ? myVotes.get(`${d.id}:${user.id}`) ?? null : null },
                area: d.area?.area_m2 ? { area_m2: d.area.area_m2, center_lat: d.area.center_lat, center_lng: d.area.center_lng, polygon: d.area.polygon } : null,
            };
        });
        return res.json({ data, next_since: new Date().toISOString(), source: 'report-service' });
    }

    const data = [...store.values()]
        .filter(r => !since || r.updated_at > since)
        .sort((a, b) => b.updated_at.localeCompare(a.updated_at))
        .slice(0, limit)
        .map(r => view(r, user?.id));
    res.json({ data, next_since: new Date().toISOString(), source: 'bff-cache' });
});

// ---------- GET /reports/:id ----------
reports.get('/reports/:id', optionalAuth, (req, res) => {
    const r = store.get(String(req.params.id));
    if (!r) return fail(res, 404, 'NOT_FOUND', 'Relato não encontrado');
    const user = res.locals.user as AuthUser | null;
    res.json(view(r, user?.id));
});