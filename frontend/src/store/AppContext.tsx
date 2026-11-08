import React, { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { useNetInfo } from '@react-native-community/netinfo';
import { HazardType, Status } from '../theme';
import { initialModQueue, initialReports, NEED_VOTES, Report, Severity } from '../data/mock';
import { enqueue, flushOutbox, pendingIds } from '../services/outbox';
import { sendEvent } from '../services/api';
import * as auth from '../services/auth';
import * as Crypto from 'expo-crypto';

export type Filters = { types: Record<HazardType, boolean>; statuses: Record<Status, boolean>; radiusKm: number; period: '24h' | '7d' | '30d' };
const allTypes = { alag: true, temp: true, inc: true, desl: true };
const defaultFilters: Filters = { types: allTypes, statuses: { pend: true, val: true, fin: false }, radiusKm: 2, period: '24h' };
const timeNow = () => new Date().toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });

function useAppState() {
  const net = useNetInfo();
  const [simOffline, setSimOffline] = useState(false);
  const offline = simOffline || net.isConnected === false;

  // Usuário real vindo do login-service
  const [user, setUser] = useState<auth.User | null>(null);
  const authed = !!user;
  useEffect(() => { auth.restoreSession().then(setUser); }, []);

  const [reports, setReports] = useState<Report[]>(initialReports);
  const [selectedId, setSelectedId] = useState('r1');
  const [myVotes, setMyVotes] = useState<Record<string, 'yes' | 'no'>>({});
  const [layers, setLayers] = useState({ reports: true, aqi: false, heat: false });
  const [filters, setFilters] = useState<Filters>(defaultFilters);
  const [alertRadius, setAlertRadius] = useState(2);
  const [alertTypes, setAlertTypes] = useState<Record<HazardType, boolean>>(allTypes);
  const [conflict, setConflict] = useState<'open' | 'mine' | 'server'>('open');
  const [syncing, setSyncing] = useState(false);
  const [lastSync, setLastSync] = useState('09:12');
  const [modQueue, setModQueue] = useState(initialModQueue);
  const [toast, setToast] = useState('');
  const toastTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const flash = useCallback((m: string) => {
    if (toastTimer.current) clearTimeout(toastTimer.current);
    setToast(m);
    toastTimer.current = setTimeout(() => setToast(''), 2400);
  }, []);

  const runSync = useCallback(async () => {
    setSyncing(true);
    await flushOutbox(sendEvent);
    const left = pendingIds();
    setReports(rs => rs.map(r => (left.has(r.id) || left.has('vote:' + r.id) ? r : { ...r, sync: 'ok' })));
    setSyncing(false);
    setLastSync(timeNow());
    flash(left.size ? 'Alguns itens não foram enviados. Tentaremos de novo.' : 'Relatos pendentes sincronizados');
  }, [flash]);

  // Ao reconectar, envia a fila local automaticamente
  const wasOffline = useRef(offline);
  useEffect(() => {
    if (wasOffline.current && !offline && reports.some(r => r.sync === 'pending')) runSync();
    wasOffline.current = offline;
  }, [offline]);

  const selected = reports.find(r => r.id === selectedId) ?? reports[0];
  const visibleReports = useMemo(() => reports.filter(r => filters.types[r.type] && filters.statuses[r.status]), [reports, filters]);
  const pendingCount = reports.filter(r => r.sync === 'pending').length + (conflict === 'open' ? 1 : 0);
  const filtersActive = Object.values(filters.types).some(v => !v) || !filters.statuses.pend || !filters.statuses.val || filters.statuses.fin;

  const vote = (id: string, yes: boolean) => {
    setMyVotes(v => ({ ...v, [id]: yes ? 'yes' : 'no' }));
    setReports(rs => rs.map(r => {
      if (r.id !== id) return r;
      let votes = r.votes + (yes ? 1 : 0);
      let st = r.status;
      if (votes >= NEED_VOTES) { st = r.status === 'pend' ? 'val' : 'fin'; votes = r.status === 'pend' ? 0 : NEED_VOTES; }
      return { ...r, votes, status: st, sync: offline ? 'pending' : r.sync };
    }));
    const row = { id: 'vote:' + id, kind: 'VoteCast' as const, payload: JSON.stringify({ reportId: id, yes }), attempts: 0, created_at: Date.now() };
    if (offline) enqueue(row.id, row.kind, { reportId: id, yes });
    else sendEvent(row).catch(() => enqueue(row.id, row.kind, { reportId: id, yes }));
    flash(offline ? 'Voto salvo no aparelho, envia quando a rede voltar' : 'Voto registrado. Obrigado!');
  };

  const submitReport = (data: { type: HazardType; severity: Severity; description: string; lat: number; lng: number; addr: string }) => {
    const id = Crypto.randomUUID();
    const rep: Report = {
      id, type: data.type, title: data.description.trim() || 'Ocorrência relatada', addr: data.addr, dist: '0 m', ago: 'agora',
      by: user?.username ?? 'Visitante',
      status: 'pend', votes: 0, lat: data.lat, lng: data.lng, radius: 110, mine: true, sync: offline ? 'pending' : 'ok', severity: data.severity,
    };
    setReports(rs => [rep, ...rs]);
    setSelectedId(id);
    const payload = { ...data, id };
    if (offline) enqueue(id, 'ReportCreated', payload);
    else sendEvent({ id, kind: 'ReportCreated', payload: JSON.stringify(payload), attempts: 0, created_at: Date.now() }).catch(() => enqueue(id, 'ReportCreated', payload));
    return rep;
  };

  return {
    offline, simOffline, setSimOffline, syncing, lastSync, runSync, pendingCount,
    authed, user,
    login: async (u: string, p: string) => setUser(await auth.login(u, p)),
    register: async (u: string, p: string) => setUser(await auth.register(u, p)),
    logout: () => { auth.logout(); setUser(null); },
    reports, visibleReports, selected, selectedId, setSelectedId, myVotes, vote, submitReport,
    layers, toggleLayer: (k: keyof typeof layers) => setLayers(l => ({ ...l, [k]: !l[k] })),
    filters, setFilters, filtersActive, clearFilters: () => setFilters({ ...defaultFilters, statuses: { pend: true, val: true, fin: true } }),
    alertRadius, setAlertRadius, alertTypes, toggleAlertType: (k: HazardType) => setAlertTypes(a => ({ ...a, [k]: !a[k] })),
    conflict, resolveConflict: (v: 'mine' | 'server') => { setConflict(v); flash(v === 'mine' ? 'Sua versão será enviada' : 'Versão do servidor aplicada'); },
    modQueue, moderate: (id: string, ok: boolean) => { setModQueue(q => q.filter(m => m.id !== id)); flash(ok ? 'Relato aprovado' : 'Relato reprovado e removido'); },
    toast, flash,
  };
}

type Ctx = ReturnType<typeof useAppState>;
const AppCtx = createContext<Ctx | null>(null);

export function AppProvider({ children }: { children: React.ReactNode }) {
  const value = useAppState();
  return <AppCtx.Provider value={value}>{children}</AppCtx.Provider>;
}

export function useApp() {
  const v = useContext(AppCtx);
  if (!v) throw new Error('useApp precisa estar dentro de <AppProvider>');
  return v;
}