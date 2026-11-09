// Integração com o backend pelo BFF (services/bff)
import type { OutboxRow, SendResult } from './outbox';
import { URLS } from './config';
import { getToken } from './auth';

export const API_URL = URLS.bff;
export const WS_URL = '';

// Formato devolvido por GET /reports no BFF
export type ServerArea = {
  id: string;
  status: 'candidate' | 'active' | 'resolved' | 'expired';
  area_m2: number;
  center: { latitude: number; longitude: number };
  polygon: { type: string; coordinates: number[][][] } | null; // GeoJSON: [longitude, latitude]
  // Votos "a situação ainda está acontecendo?"; needed = mínimo de "já resolvido" para encerrar
  votes: { still_happening: number; resolved: number; needed: number };
};
export type ServerReport = {
  id: string;
  type: string; // Flooding | Thunderstorm | Fire Disaster | Landslide
  status: 'pending' | 'validated' | 'rejected' | 'resolved';
  severity: string; // LOW | MEDIUM | HIGH
  description: string;
  address: string;
  latitude: number;
  longitude: number;
  user_id: string;
  username: string;
  datetime: string;
  updated_at: string;
  votes: { positive: number; negative: number; my_vote: boolean | null };
  area: ServerArea | null;
};

// O que vai para a fila do aparelho. userId = conta que fez a ação:
// o item só é enviado quando essa conta está logada (o token é dela).
export type ReportCreatedPayload = {
  userId: string;
  body: { id: string; type: string; severity: string; description: string; latitude: number; longitude: number; address: string; datetime: string };
};
export type VotePayload = { userId: string; reportId: string; positive: boolean };
export type AreaVotePayload = { userId: string; areaId: string; stillHappening: boolean };

async function call(path: string, init: RequestInit = {}, timeoutMs = 8000): Promise<Response> {
  const token = await getToken();
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    return await fetch(API_URL + path, {
      ...init,
      signal: controller.signal,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(init.headers || {}),
      },
    });
  } finally {
    clearTimeout(timer);
  }
}

export async function fetchReports(): Promise<ServerReport[]> {
  const res = await call('/reports?limit=500');
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const body = await res.json();
  return (body.data ?? []) as ServerReport[];
}

// Envia um item da fila do aparelho para o BFF
export async function sendEvent(row: OutboxRow, currentUserId: string | undefined): Promise<SendResult> {
  const payload = JSON.parse(row.payload);
  if (!currentUserId || payload.userId !== currentUserId) return 'skip';

  let res: Response;
  switch (row.kind) {
    case 'ReportCreated': {
      const p = payload as ReportCreatedPayload;
      res = await call('/reports', { method: 'POST', headers: { 'Idempotency-Key': row.id }, body: JSON.stringify(p.body) });
      break;
    }
    case 'VoteCast': {
      const p = payload as VotePayload;
      res = await call(`/reports/${p.reportId}/vote`, { method: 'POST', body: JSON.stringify({ positive: p.positive }) });
      break;
    }
    case 'AreaVoteCast': {
      const p = payload as AreaVotePayload;
      res = await call(`/areas/${p.areaId}/vote`, { method: 'POST', body: JSON.stringify({ still_happening: p.stillHappening }) });
      break;
    }
    default:
      return 'drop';
  }

  if (res.ok) return 'sent';
  if (res.status === 401) return 'auth'; // token expirado (dura 1 h): precisa entrar de novo
  // Recusa definitiva: dado inválido (400) ou voto repetido (409). Tentar de novo não resolve.
  if (res.status === 400 || res.status === 409) return 'drop';
  // 404/502 (relato ainda não chegou ao validacao), 503 (serviço fora): tenta depois
  throw new Error(`HTTP ${res.status}`);
}

// Tempo real: o servidor publica ocorrências por GeoHash/H3, o app assina a célula atual
export function connectRealtime(cell: string, onEvent: (e: any) => void) {
  if (!WS_URL) return () => {};
  let ws: WebSocket | null = null;
  let retry = 0;
  let closed = false;
  const open = () => {
    ws = new WebSocket(`${WS_URL}?cell=${cell}`);
    ws.onopen = () => { retry = 0; };
    ws.onmessage = msg => { try { onEvent(JSON.parse(String(msg.data))); } catch {} };
    ws.onclose = () => { if (!closed) setTimeout(open, Math.min(30000, 1000 * 2 ** retry++)); };
  };
  open();
  return () => { closed = true; ws?.close(); };
}
