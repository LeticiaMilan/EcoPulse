import React, { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { useNetInfo } from '@react-native-community/netinfo';
import { HazardType, Status, hazard } from '../theme';
import { initialModQueue, NEED_VOTES, Report, ReportArea, Severity } from '../data/mock';
import { enqueue, flushOutbox, pendingCount as outboxCount, pendingIds, OutboxRow } from '../services/outbox';
import { fetchReports, sendEvent, ServerReport, ReportCreatedPayload, VotePayload, AreaVotePayload } from '../services/api';
import * as auth from '../services/auth';
import * as cache from '../services/cache';
import * as Crypto from 'expo-crypto';

export type Filters = { types: Record<HazardType, boolean>; statuses: Record<Status, boolean>; radiusKm: number; period: '24h' | '7d' | '30d' };
const allTypes = { alag: true, temp: true, inc: true, desl: true };
const defaultFilters: Filters = { types: allTypes, statuses: { pend: true, val: true, fin: false }, radiusKm: 2, period: '24h' };
const timeNow = () => new Date().toLocaleTimeString('pt-BR', { hour: '2-digit', minute: '2-digit' });
const REFRESH_MS = 15_000;

// ---------- conversão app <-> backend ----------
const TYPE_FROM_SERVER: Record<string, HazardType> = { Flooding: 'alag', Thunderstorm: 'temp', 'Fire Disaster': 'inc', Landslide: 'desl' };
const TYPE_TO_SERVER: Record<HazardType, string> = { alag: 'Flooding', temp: 'Thunderstorm', inc: 'Fire Disaster', desl: 'Landslide' };
const SEV_FROM_SERVER: Record<string, Severity> = { LOW: 'Baixa', MEDIUM: 'Média', HIGH: 'Alta' };
const SEV_TO_SERVER: Record<Severity, string> = { Baixa: 'LOW', 'Média': 'MEDIUM', Alta: 'HIGH' };

function toArea(a: ServerReport['area']): ReportArea | null {
  if (!a) return null;
  const ring = a.polygon?.coordinates?.[0] ?? [];
  return {
    id: a.id, status: a.status, areaM2: a.area_m2,
    polygon: ring.map(([lng, lat]) => ({ latitude: lat, longitude: lng })),
    votes: { stillHappening: a.votes?.still_happening ?? 0, resolved: a.votes?.resolved ?? 0, needed: a.votes?.needed || 3 },
  };
}

function fromServer(s: ServerReport): Report {
  const type = TYPE_FROM_SERVER[s.type] ?? 'alag';
  const status: Status = s.status === 'pending' ? 'pend' : s.status === 'validated' ? 'val' : 'fin';
  return {
    id: s.id, type, status,
    title: s.description || `${hazard[type].name} relatado`,
    addr: s.address || `${s.latitude.toFixed(4)}, ${s.longitude.toFixed(4)}`,
    dist: '', ago: '', // calculados na hora (dependem da posição e do relógio)
    by: s.username || 'Usuário',
    userId: s.user_id,
    votes: s.votes.positive, negative: s.votes.negative,
    lat: s.latitude, lng: s.longitude, radius: 110,
    mine: false, sync: 'ok',
    severity: SEV_FROM_SERVER[s.severity] ?? 'Média',
    datetime: s.datetime,
    rejected: s.status === 'rejected',
    area: toArea(s.area),
  };
}

function agoText(iso: string | undefined, now: number) {
  if (!iso) return '';
  const min = Math.floor((now - new Date(iso).getTime()) / 60000);
  if (!Number.isFinite(min)) return '';
  if (min < 1) return 'agora';
  if (min < 60) return `há ${min} min`;
  const h = Math.floor(min / 60);
  return h < 24 ? `há ${h} h` : `há ${Math.floor(h / 24)} d`;
}

// Distância em metros entre dois pontos (haversine)
export function distanceM(here: { lat: number; lng: number }, lat: number, lng: number) {
  const R = 6371000, rad = Math.PI / 180;
  const dLat = (lat - here.lat) * rad, dLng = (lng - here.lng) * rad;
  const a = Math.sin(dLat / 2) ** 2 + Math.cos(here.lat * rad) * Math.cos(lat * rad) * Math.sin(dLng / 2) ** 2;
  return 2 * R * Math.asin(Math.sqrt(a));
}

function distText(here: { lat: number; lng: number } | null, lat: number, lng: number) {
  if (!here) return '';
  const m = distanceM(here, lat, lng);
  return m < 1000 ? `${Math.round(m / 10) * 10} m` : `${(m / 1000).toFixed(1).replace('.', ',')} km`;
}

// Legenda curta usada no mapa e na lista
export function voteCaption(r: Report) {
  const conf = (n: number) => `${n} ${n === 1 ? 'confirmação' : 'confirmações'}`;
  if (r.status === 'pend') return `${r.votes}/${NEED_VOTES} confirmam`;
  if (r.status === 'val') return r.area?.status === 'active' ? 'área afetada ativa' : 'confirmado pela comunidade';
  if (r.rejected) return r.votes ? `${conf(r.votes)} · ${r.negative ?? 0} contra` : 'ninguém confirmou';
  return 'situação resolvida';
}

// Chave do voto: a confirmação é por relato; o "já foi resolvido?" é por área
export const voteKey = (r: Report) => (r.status === 'val' && r.area ? `area:${r.area.id}` : `report:${r.id}`);

type VoteMap = Record<string, 'yes' | 'no'>;

function useAppState() {
  const net = useNetInfo();
  const [simOffline, setSimOffline] = useState(false);
  const offline = simOffline || net.isConnected === false;

  // Usuário real vindo do login-service
  const [user, setUser] = useState<auth.User | null>(null);
  const [sessionChecked, setSessionChecked] = useState(false);
  const authed = !!user;
  useEffect(() => { auth.restoreSession().then(u => { setUser(u); setSessionChecked(true); }); }, []);

  // Relatos vindos do servidor (cópia guardada no SQLite) + relatos criados aqui que o servidor ainda não listou
  const [serverReports, setServerReports] = useState<Report[]>(() => cache.load<Report[]>('reports', []));
  const [localReports, setLocalReports] = useState<Report[]>(() => cache.load<Report[]>('localReports', []));
  useEffect(() => cache.save('reports', serverReports), [serverReports]);
  useEffect(() => cache.save('localReports', localReports), [localReports]);

  // Votos dados neste aparelho, por conta: { [userId]: { 'report:<id>' | 'area:<id>': 'yes' | 'no' } }
  const [votesByUser, setVotesByUser] = useState<Record<string, VoteMap>>(() => cache.load('votesByUser', {}));
  useEffect(() => cache.save('votesByUser', votesByUser), [votesByUser]);

  const [queueIds, setQueueIds] = useState<Set<string>>(() => pendingIds());
  const [queueSize, setQueueSize] = useState(() => outboxCount());
  const refreshQueue = useCallback(() => { setQueueIds(pendingIds()); setQueueSize(outboxCount()); }, []);

  const [here, setHere] = useState<{ lat: number; lng: number } | null>(null);
  const [now, setNow] = useState(Date.now());

  const [selectedId, setSelectedId] = useState('');
  const [focusTick, setFocusTick] = useState(0);
  const focusOn = (id: string) => { setSelectedId(id); setFocusTick(n => n + 1); };
  const [layers, setLayers] = useState({ reports: true, aqi: false, heat: false });
  const [filters, setFilters] = useState<Filters>(defaultFilters);
  // Preferências de alerta ficam salvas no aparelho
  const [alertRadius, setAlertRadiusRaw] = useState(() => cache.load<number>('alertRadius', 2));
  const [alertTypes, setAlertTypes] = useState<Record<HazardType, boolean>>(() => cache.load('alertTypes', allTypes));
  useEffect(() => cache.save('alertRadius', alertRadius), [alertRadius]);
  useEffect(() => cache.save('alertTypes', alertTypes), [alertTypes]);
  // O slider pode entregar valores como 1.5000000002: arredonda para o passo de 0,5 km
  const setAlertRadius = useCallback((v: number) => setAlertRadiusRaw(Math.min(5, Math.max(0.5, Math.round(v * 2) / 2))), []);
  // Alertas já vistos: "<id>:<status>" (um relato que muda de status vira alerta novo)
  const [seenAlerts, setSeenAlerts] = useState<string[]>(() => cache.load<string[]>('seenAlerts', []));
  useEffect(() => cache.save('seenAlerts', seenAlerts), [seenAlerts]);
  // O conflito de edição ainda não existe no backend: começa sem conflito
  const [conflict, setConflict] = useState<'none' | 'open' | 'mine' | 'server'>('none');
  const [syncing, setSyncing] = useState(false);
  const [lastSync, setLastSync] = useState(() => cache.load<string>('lastSync', '—'));
  const [modQueue, setModQueue] = useState(initialModQueue);
  const [toast, setToast] = useState('');
  const toastTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const flash = useCallback((m: string) => {
    if (toastTimer.current) clearTimeout(toastTimer.current);
    setToast(m);
    toastTimer.current = setTimeout(() => setToast(''), 2400);
  }, []);

  // Busca a lista no servidor
  const refreshReports = useCallback(async () => {
    if (offline) return false;
    try {
      const list = (await fetchReports()).map(fromServer);
      setServerReports(list);
      const listed = new Set(list.map(r => r.id));
      setLocalReports(ls => ls.filter(r => !listed.has(r.id)));
      setNow(Date.now());
      return true;
    } catch {
      return false;
    }
  }, [offline]);

  // Envia a fila do aparelho e depois atualiza a lista
  const syncRunning = useRef(false);
  const runSync = useCallback(async (opts: { quiet?: boolean } = {}) => {
    if (offline || syncRunning.current) return;
    syncRunning.current = true;
    setSyncing(true);
    const dropped: OutboxRow[] = [];
    let sessionExpired = false;
    try {
      await flushOutbox(async row => {
        const result = await sendEvent(row, user?.id);
        if (result === 'drop') dropped.push(row);
        if (result === 'auth') sessionExpired = true;
        return result;
      });
      refreshQueue();
      if (sessionExpired) {
        // O que estava na fila continua lá e é enviado depois de entrar de novo
        auth.logout();
        setUser(null);
        flash('Sua sessão expirou. Entre de novo para enviar seus relatos e votos.');
      }
      // Relato recusado pelo servidor não vai aparecer na lista: tira da tela
      const droppedReports = new Set(dropped.filter(r => r.kind === 'ReportCreated').map(r => r.id));
      if (droppedReports.size) setLocalReports(ls => ls.filter(r => !droppedReports.has(r.id)));
      const ok = await refreshReports();
      if (ok) { const t = timeNow(); setLastSync(t); cache.save('lastSync', t); }
      if (!opts.quiet) {
        const left = outboxCount();
        flash(!ok ? 'Servidor indisponível. Tentaremos de novo.' : left ? 'Alguns itens não foram enviados. Tentaremos de novo.' : 'Tudo sincronizado');
      } else if (droppedReports.size) {
        flash('Um relato foi recusado pelo servidor');
      }
    } finally {
      syncRunning.current = false;
      setSyncing(false);
    }
  }, [offline, user, refreshQueue, refreshReports, flash]);

  // Sempre a versão mais nova de runSync (o timer não pode usar uma cópia antiga com a conta anterior)
  const runSyncRef = useRef(runSync);
  runSyncRef.current = runSync;
  const followUps = useRef<ReturnType<typeof setTimeout>[]>([]);
  useEffect(() => () => followUps.current.forEach(clearTimeout), []);

  // Sincroniza ao abrir o app, ao trocar de conta e ao voltar a internet; depois, a cada 15 s
  useEffect(() => {
    if (!sessionChecked || offline) return;
    runSyncRef.current({ quiet: true });
    const timer = setInterval(() => runSyncRef.current({ quiet: true }), REFRESH_MS);
    return () => clearInterval(timer);
  }, [sessionChecked, offline, user?.id]);

  // Lista final: "é meu?", distância, tempo e se ainda está na fila do aparelho
  const reports = useMemo(() => {
    const listed = new Set(serverReports.map(r => r.id));
    const all = [...localReports.filter(r => !listed.has(r.id)), ...serverReports];
    return all.map(r => ({
      ...r,
      mine: !!user && r.userId === user.id,
      sync: queueIds.has(r.id) ? 'pending' as const : 'ok' as const,
      ago: agoText(r.datetime, now),
      dist: distText(here, r.lat, r.lng),
    }));
  }, [serverReports, localReports, user, queueIds, now, here]);

  const myVotes: VoteMap = useMemo(() => (user ? votesByUser[user.id] ?? {} : {}), [votesByUser, user]);
  const myVoteFor = (r: Report): 'yes' | 'no' | undefined => myVotes[voteKey(r)];

  const selected: Report | undefined = reports.find(r => r.id === selectedId) ?? reports[0];
  // Filtros da tela de Filtros: tipo, status, raio (precisa da localização) e período
  const visibleReports = useMemo(() => {
    const periodMs = { '24h': 24, '7d': 24 * 7, '30d': 24 * 30 }[filters.period] * 60 * 60 * 1000;
    return reports.filter(r => {
      if (!filters.types[r.type] || !filters.statuses[r.status]) return false;
      if (here && distanceM(here, r.lat, r.lng) > filters.radiusKm * 1000) return false;
      if (r.datetime && now - new Date(r.datetime).getTime() > periodMs) return false;
      return true;
    });
  }, [reports, filters, here, now]);

  // Alertas: relatos dentro do raio de alerta e dos tipos ligados, do mais perto para o mais longe
  const nearbyAlerts = useMemo(() => {
    if (!here) return [] as { r: Report; m: number }[];
    return reports
      .filter(r => alertTypes[r.type])
      .map(r => ({ r, m: distanceM(here, r.lat, r.lng) }))
      .filter(x => x.m <= alertRadius * 1000)
      .sort((a, b) => a.m - b.m);
  }, [reports, here, alertRadius, alertTypes]);
  const alertKey = (r: Report) => `${r.id}:${r.status}`;
  // Número no sininho: alertas perto que a pessoa ainda não viu
  const alertBadge = useMemo(() => {
    const seen = new Set(seenAlerts);
    return nearbyAlerts.filter(x => !x.r.mine && !seen.has(alertKey(x.r))).length; // os próprios relatos não contam
  }, [nearbyAlerts, seenAlerts]);
  // Chamado quando a aba Alertas está aberta: zera o número
  const markAlertsSeen = useCallback(() => {
    setSeenAlerts(prev => {
      const seen = new Set(prev);
      const fresh = nearbyAlerts.map(x => alertKey(x.r)).filter(k => !seen.has(k));
      return fresh.length ? [...prev, ...fresh].slice(-1000) : prev;
    });
  }, [nearbyAlerts]);
  const pendingCount = queueSize + (conflict === 'open' ? 1 : 0);
  const filtersActive = Object.values(filters.types).some(v => !v) || !filters.statuses.pend || !filters.statuses.val || filters.statuses.fin
    || filters.radiusKm !== defaultFilters.radiusKm || filters.period !== defaultFilters.period;

  // Voto da etapa atual do relato:
  //   aguardando validação -> confirma o relato (validacao-relato)
  //   validado com área    -> "já foi resolvido?" (yes = já resolvido; vai para o serviço de área)
  // Sincroniza agora e de novo em 4 s e 10 s: o resultado passa por vários serviços e chega em alguns segundos
  const syncSoon = () => {
    runSync({ quiet: true });
    followUps.current.forEach(clearTimeout);
    followUps.current = [4000, 10000].map(ms => setTimeout(() => runSyncRef.current({ quiet: true }), ms));
  };

  const vote = (id: string, yes: boolean) => {
    if (!user) return;
    const target = reports.find(r => r.id === id);
    if (!target) return;
    const key = voteKey(target);
    if (myVotes[key]) return;

    if (target.status === 'pend') {
      if (target.mine) return; // ninguém confirma o próprio relato
      const payload: VotePayload = { userId: user.id, reportId: id, positive: yes };
      enqueue(`vote:${user.id}:${id}`, 'VoteCast', payload);
    } else if (target.status === 'val' && target.area?.status === 'active') {
      const payload: AreaVotePayload = { userId: user.id, areaId: target.area.id, stillHappening: !yes };
      enqueue(`areavote:${user.id}:${target.area.id}`, 'AreaVoteCast', payload);
    } else {
      return;
    }

    setVotesByUser(all => ({ ...all, [user.id]: { ...all[user.id], [key]: yes ? 'yes' : 'no' } }));
    refreshQueue();
    flash(offline ? 'Voto salvo no aparelho, envia quando a rede voltar' : 'Voto registrado. Obrigado!');
    if (!offline) syncSoon();
  };

  const submitReport = (data: { type: HazardType; severity: Severity; description: string; lat: number; lng: number; addr: string }) => {
    const id = Crypto.randomUUID();
    const datetime = new Date().toISOString();
    const description = (data.description.trim() || `${hazard[data.type].name} relatado`).slice(0, 280);
    const rep: Report = {
      id, type: data.type, title: description, addr: data.addr, dist: '0 m', ago: 'agora',
      by: user?.username ?? 'Visitante', userId: user?.id,
      status: 'pend', votes: 0, negative: 0, lat: data.lat, lng: data.lng, radius: 110, mine: !!user,
      sync: offline ? 'pending' : 'ok', severity: data.severity, datetime, area: null,
    };
    if (!user) return rep; // a tela já manda para o login antes de chegar aqui

    // A fila do aparelho é o único caminho de envio: online ela é enviada na hora
    const payload: ReportCreatedPayload = {
      userId: user.id,
      body: {
        id, type: TYPE_TO_SERVER[data.type], severity: SEV_TO_SERVER[data.severity], description,
        latitude: data.lat, longitude: data.lng, address: data.addr, datetime,
      },
    };
    enqueue(id, 'ReportCreated', payload);
    setLocalReports(ls => [rep, ...ls]);
    setSelectedId(id);
    refreshQueue();
    if (!offline) syncSoon();
    return rep;
  };

  return {
    offline, simOffline, setSimOffline, syncing, lastSync, runSync, pendingCount,
    authed, user,
    login: async (u: string, p: string) => setUser(await auth.login(u, p)),
    register: async (u: string, p: string) => setUser(await auth.register(u, p)),
    logout: () => { auth.logout(); setUser(null); },
    reports, visibleReports, selected, selectedId, setSelectedId, focusTick, focusOn, myVotes, myVoteFor, vote, submitReport,
    here, setHere,
    layers, toggleLayer: (k: keyof typeof layers) => setLayers(l => ({ ...l, [k]: !l[k] })),
    filters, setFilters, filtersActive, clearFilters: () => setFilters({ ...defaultFilters, statuses: { pend: true, val: true, fin: true } }),
    alertRadius, setAlertRadius, alertTypes, toggleAlertType: (k: HazardType) => setAlertTypes(a => ({ ...a, [k]: !a[k] })),
    nearbyAlerts, alertBadge, markAlertsSeen,
    conflict, simulateConflict: () => setConflict('open'),
    resolveConflict: (v: 'mine' | 'server') => { setConflict(v); flash(v === 'mine' ? 'Sua versão será enviada' : 'Versão do servidor aplicada'); },
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
