import * as SecureStore from 'expo-secure-store';
import { URLS } from './config';

const TOKEN_KEY = 'ecopulse.token';
export type User = { id: string; username: string };

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
    let res: Response;
    try {
        res = await fetch(URLS.login + path, { ...init, headers: { 'Content-Type': 'application/json', ...(init.headers || {}) } });
    } catch {
        throw new Error('Sem conexão com o servidor. Confira o IP e o Wi-Fi.');
    }
    if (res.status === 401) throw new Error('Usuário ou senha incorretos.');
    if (res.status === 409) throw new Error('Esse usuário já existe.');
    if (!res.ok) throw new Error('Erro ' + res.status);
    return res.json();
}

export const me = (token: string) =>
    request<User>('/auth/me', { headers: { Authorization: `Bearer ${token}` } });

export async function login(username: string, password: string) {
    const r = await request<{ token: string }>('/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) });
    await SecureStore.setItemAsync(TOKEN_KEY, r.token);
    return me(r.token);
}

export async function register(username: string, password: string) {
    await request('/users', { method: 'POST', body: JSON.stringify({ username, password }) });
    return login(username, password);
}

export async function restoreSession() {
    const token = await SecureStore.getItemAsync(TOKEN_KEY);
    if (!token) return null;
    try { return await me(token); } catch { await logout(); return null; }
}

export const logout = () => SecureStore.deleteItemAsync(TOKEN_KEY);

// Token da sessão atual, usado nas chamadas ao BFF
export const getToken = () => SecureStore.getItemAsync(TOKEN_KEY);