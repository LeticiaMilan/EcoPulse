import { Router } from 'express';
import { requireAuth, optionalAuth, AuthUser } from './auth';
import { publish } from './rabbit';

export const reports = Router();

const TYPES = ['Flooding', 'Thunderstorm', 'Fire Disaster', 'Landslide'];
const SEVERITIES = ['LOW', 'MEDIUM', 'HIGH'];
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const REPORT_URL = process.env.REPORT_SERVICE_URL ?? '';
const VALIDATION_URL = process.env.VALIDATION_SERVICE_URL ?? '';
const LOGIN_URL = process.env.LOGIN_SERVICE_URL ?? '';

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
        local.updated_at = new Date().toISOString();
    }

    res.status(202).json({ report_id: id, positive, accepted_at: new Date().toISOString() });
});

// ---------- GET /reports ----------
// Monta a lista juntando os serviços (cada um é dono de uma parte do relato):
//   validacao-relato -> relato, descrição, votos e estado da votação
//   report-service   -> área afetada à qual o relato foi ligado
//   login-service    -> nome de quem criou
//   cópia local      -> gravidade e endereço, e relatos recém-criados que ainda estão a caminho
// Status devolvido: pending | validated | rejected | resolved

type Relato = {
    id: string; usuario_id: string; tipo: string; descricao: string; latitude: number; longitude: number;
    votos_positivos: number; votos_negativos: number; estado: string; criado_em: string; atualizado_em: string;
};
type SyncArea = {
    id: string; status: string; area_m2: number; center_lat: number; center_lng: number; polygon: unknown; updated_at: string;
    votes?: { still_happening: number; resolved: number; needed: number };
};
type SyncRow = { id: string; area?: SyncArea };

// ok=false: o serviço não respondeu. Um corpo `null` com ok=true é lista vazia (Go serializa slice vazio como null).
async function getJson<T>(url: string): Promise<{ ok: boolean; body: T | null }> {
    try {
        const res = await fetch(url, { signal: AbortSignal.timeout(3000) });
        if (!res.ok) return { ok: false, body: null };
        return { ok: true, body: (await res.json()) as T | null };
    } catch {
        return { ok: false, body: null };
    }
}

const usernames = new Map<string, string>();
let usernamesAt = 0;
async function refreshUsernames() {
    if (!LOGIN_URL || Date.now() - usernamesAt < 10_000) return;
    const { ok, body } = await getJson<{ id: string; username: string }[]>(`${LOGIN_URL}/users`);
    if (!ok) return;
    for (const u of body ?? []) usernames.set(u.id, u.username);
    usernamesAt = Date.now();
}

function statusOf(estado: string | undefined, areaStatus: string | undefined) {
    if (areaStatus === 'resolved' || areaStatus === 'expired') return 'resolved';
    if (areaStatus === 'active') return 'validated';
    if (estado === 'RELATO_VALIDADO') return 'validated';
    if (estado === 'REJEITADO') return 'rejected';
    return 'pending';
}

async function buildReports(userId: string | undefined) {
    const since = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString();
    const [relatos, sync] = await Promise.all([
        VALIDATION_URL ? getJson<Relato[]>(`${VALIDATION_URL}/validacaorelatos?limit=500`) : Promise.resolve({ ok: false, body: null }),
        REPORT_URL ? getJson<{ data: SyncRow[] | null }>(`${REPORT_URL}/report/api/sync?limit=100&since=${encodeURIComponent(since)}`) : Promise.resolve({ ok: false, body: null }),
        refreshUsernames(),
    ]);

    const relatoById = new Map((relatos.body ?? []).map(r => [r.id, r]));
    const areaByReport = new Map<string, SyncArea>();
    for (const row of sync.body?.data ?? []) if (row.area?.id) areaByReport.set(row.id, row.area);

    const ids = new Set<string>([...relatoById.keys(), ...store.keys()]);
    const data = [...ids].map(id => {
        const v = relatoById.get(id);
        const s = store.get(id);
        const a = areaByReport.get(id);
        const user_id = v?.usuario_id ?? s?.user_id ?? '';
        return {
            id,
            type: v?.tipo ?? s?.type ?? '',
            status: statusOf(v?.estado ?? 'POSSIVEL_RELATO', a?.status),
            severity: s?.severity ?? 'MEDIUM',
            description: v?.descricao || s?.description || '',
            address: s?.address ?? '',
            latitude: v?.latitude ?? s?.latitude ?? 0,
            longitude: v?.longitude ?? s?.longitude ?? 0,
            user_id,
            username: usernames.get(user_id) ?? s?.username ?? '',
            datetime: v?.criado_em ?? s?.datetime ?? '',
            updated_at: a?.updated_at ?? v?.atualizado_em ?? s?.updated_at ?? '',
            votes: {
                positive: v?.votos_positivos ?? s?.votes.positive ?? 0,
                negative: v?.votos_negativos ?? s?.votes.negative ?? 0,
                my_vote: userId ? myVotes.get(`${id}:${userId}`) ?? null : null,
            },
            area: a ? {
                id: a.id, status: a.status, area_m2: a.area_m2,
                center: { latitude: a.center_lat, longitude: a.center_lng }, polygon: a.polygon,
                votes: a.votes ?? { still_happening: 0, resolved: 0, needed: 0 },
            } : null,
        };
    });
    data.sort((x, y) => y.datetime.localeCompare(x.datetime));
    return { data, partial: !relatos.ok || !sync.ok };
}

reports.get('/reports', optionalAuth, async (req, res) => {
    const user = res.locals.user as AuthUser | null;
    const limit = Math.min(500, Math.max(1, Number(req.query.limit ?? 200) || 200));
    const { data, partial } = await buildReports(user?.id);
    // partial = algum serviço não respondeu (a lista pode estar incompleta)
    res.json({ data: data.slice(0, limit), partial, generated_at: new Date().toISOString() });
});

// ---------- GET /reports/:id ----------
reports.get('/reports/:id', optionalAuth, async (req, res) => {
    const user = res.locals.user as AuthUser | null;
    const { data } = await buildReports(user?.id);
    const r = data.find(d => d.id === String(req.params.id));
    if (!r) return fail(res, 404, 'NOT_FOUND', 'Relato não encontrado');
    res.json(r);
});

// ---------- POST /areas/:id/vote ----------
// Voto "a situação ainda está acontecendo?" numa área afetada ativa.
// O calculate-area-service encerra a área (RESOLVED) quando os "não" atingem o mínimo e superam os "sim".
reports.post('/areas/:id/vote', requireAuth, async (req, res) => {
    const user = res.locals.user as AuthUser;
    const id = String(req.params.id);
    if (!UUID.test(id)) return fail(res, 400, 'INVALID_ID', 'id deve ser UUID');
    if (typeof req.body?.still_happening !== 'boolean') return fail(res, 400, 'INVALID_VOTE', 'still_happening deve ser true ou false');

    try {
        await publish('EVENTS.AFFECTED_AREA.VOTE', { area_id: id, user_id: user.id, still_happening: req.body.still_happening });
    } catch {
        return fail(res, 503, 'BROKER_UNAVAILABLE', 'Serviço temporariamente indisponível');
    }
    res.status(202).json({ area_id: id, still_happening: req.body.still_happening, accepted_at: new Date().toISOString() });
});
