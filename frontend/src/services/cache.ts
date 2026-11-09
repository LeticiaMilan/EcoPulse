// Cópia local (SQLite) do que veio do servidor, para o app abrir com dados mesmo sem rede
// e não perder nada ao ser fechado. A fonte da verdade continua sendo o backend.
import Storage from 'expo-sqlite/kv-store';

export function load<T>(key: string, fallback: T): T {
  try {
    const raw = Storage.getItemSync(key);
    return raw ? (JSON.parse(raw) as T) : fallback;
  } catch {
    return fallback;
  }
}

export function save(key: string, value: unknown) {
  try {
    Storage.setItemSync(key, JSON.stringify(value));
  } catch {
    // cache é só conveniência: se falhar, o app segue com o que tem em memória
  }
}
