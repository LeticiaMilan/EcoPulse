// fila local (offline-first) em SQLite, cada evento tem id único = chave de idempotência no backend 
import * as SQLite from 'expo-sqlite';

export type OutboxRow = { id: string; kind: 'ReportCreated' | 'VoteCast' | 'AreaVoteCast' | 'ReportEdited'; payload: string; attempts: number; created_at: number };

// Resultado de um envio:
//   'sent'  -> o servidor aceitou (sai da fila)
//   'drop'  -> o servidor recusou de vez, ex.: dado inválido ou voto repetido (sai da fila)
//   'skip'  -> não dá para enviar agora, ex.: item de outra conta (fica na fila, segue para o próximo)
//   'auth'  -> sessão expirada (401): fica na fila e é enviado depois de entrar de novo
// Erro lançado -> falha temporária (rede/servidor fora): conta uma tentativa e para.
export type SendResult = 'sent' | 'drop' | 'skip' | 'auth';

const db = SQLite.openDatabaseSync('ecopulse.db');
db.execSync(`
  PRAGMA journal_mode = WAL;
  CREATE TABLE IF NOT EXISTS outbox (
    id TEXT PRIMARY KEY NOT NULL,
    kind TEXT NOT NULL,
    payload TEXT NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL
  );
`);

export const MAX_ATTEMPTS = 5;

export function enqueue(id: string, kind: OutboxRow['kind'], payload: unknown) {
  db.runSync('INSERT OR REPLACE INTO outbox (id, kind, payload, attempts, created_at) VALUES (?, ?, ?, 0, ?)', [id, kind, JSON.stringify(payload), Date.now()]);
}

export function listPending(): OutboxRow[] {
  return db.getAllSync<OutboxRow>('SELECT * FROM outbox ORDER BY created_at');
}

export function pendingIds(): Set<string> {
  return new Set(listPending().map(r => r.id));
}

export function pendingCount(): number {
  return db.getFirstSync<{ n: number }>('SELECT COUNT(*) AS n FROM outbox')?.n ?? 0;
}

// envia em ordem, em falha incrementa tentativas e para (backoff fica a cargo de quem chama)
export async function flushOutbox(send: (row: OutboxRow) => Promise<SendResult>) {
  for (const row of listPending()) {
    try {
      const result = await send(row);
      if (result === 'skip' || result === 'auth') continue;
      db.runSync('DELETE FROM outbox WHERE id = ?', [row.id]);
    } catch {
      db.runSync('UPDATE outbox SET attempts = attempts + 1 WHERE id = ?', [row.id]);
      if (row.attempts + 1 >= MAX_ATTEMPTS) continue; // fica para revisão manual na tela de sincronização
      break;
    }
  }
}
