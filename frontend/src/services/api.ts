// Integração com o backend (Go + RabbitMQ) 
import type { OutboxRow } from './outbox';

export const API_URL = ''; 
export const WS_URL = '';  

const wait = (ms: number) => new Promise(res => setTimeout(res, ms));

export async function sendEvent(row: OutboxRow) {
  if (!API_URL) { await wait(400); return; }
  const res = await fetch(`${API_URL}/events`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Idempotency-Key': row.id },
    body: JSON.stringify({ id: row.id, kind: row.kind, payload: JSON.parse(row.payload), createdAt: row.created_at }),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
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
