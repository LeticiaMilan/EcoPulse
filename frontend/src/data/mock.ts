import type { HazardType, Status } from '../theme';

export type Severity = 'Baixa' | 'Média' | 'Alta';
export type Report = {
  id: string;
  type: HazardType;
  title: string;
  addr: string;
  dist: string;
  ago: string;
  by: string;
  status: Status;
  votes: number;
  lat: number;
  lng: number;
  radius: number; // metros na zona de influência
  mine: boolean;
  sync: 'ok' | 'pending';
  severity: Severity;
};

export const USER = { lat: -21.1775, lng: -47.8103 };
export const CURRENT_USER = { name: 'Usuário 1', initials: 'U1', role: 'Moderador(a)', city: 'Ribeirão Preto, SP' };
export const NEED_VOTES = 5;

export const initialReports: Report[] = [
  { id: 'r1', type: 'alag', title: 'Rua alagada, carros parados', addr: 'Rua A × Av. B', dist: '380 m', ago: 'há 12 min', by: 'Usuário 2', status: 'pend', votes: 4, lat: -21.1752, lng: -47.8071, radius: 160, mine: false, sync: 'ok', severity: 'Alta' },
  { id: 'r2', type: 'desl', title: 'Deslizamento de terra', addr: 'Rua C, 820', dist: '1,2 km', ago: 'há 40 min', by: CURRENT_USER.name, status: 'val', votes: 2, lat: -21.1690, lng: -47.8168, radius: 120, mine: true, sync: 'ok', severity: 'Alta' },
  { id: 'r3', type: 'inc', title: 'Foco de incêndio em terreno', addr: 'Parque A', dist: '1,8 km', ago: 'há 1 h', by: 'Usuário 3', status: 'val', votes: 1, lat: -21.1880, lng: -47.8043, radius: 130, mine: false, sync: 'ok', severity: 'Média' },
  { id: 'r4', type: 'temp', title: 'Árvore caída após vendaval', addr: 'Av. D, 400', dist: '650 m', ago: 'há 3 h', by: CURRENT_USER.name, status: 'fin', votes: 5, lat: -21.1825, lng: -47.8143, radius: 100, mine: true, sync: 'ok', severity: 'Média' },
  { id: 'r5', type: 'alag', title: 'Bueiro transbordando', addr: 'Rua E, 55', dist: '520 m', ago: 'há 5 min', by: CURRENT_USER.name, status: 'pend', votes: 0, lat: -21.1720, lng: -47.8118, radius: 100, mine: true, sync: 'pending', severity: 'Média' },
];

export const AQI_POINTS = [
  { lat: -21.1760, lng: -47.8198, v: 42 },
  { lat: -21.1700, lng: -47.8058, v: 88 },
  { lat: -21.1775, lng: -47.8018, v: 132 },
  { lat: -21.1860, lng: -47.8148, v: 71 },
  { lat: -21.1845, lng: -47.8088, v: 104 },
];

export const HEAT_SPOTS = [
  { lat: -21.1720, lng: -47.8053, radius: 420, color: '#ff8a3d' },
  { lat: -21.1850, lng: -47.8153, radius: 360, color: '#ff5a3d' },
  { lat: -21.1680, lng: -47.8123, radius: 300, color: '#ffc23d' },
];

export type ModItem = { id: string; type: HazardType; title: string; by: string; why: string };
export const initialModQueue: ModItem[] = [
  { id: 'm1', type: 'inc', title: 'Fumaça saindo de galpão', by: 'Usuário 4', why: '3 denúncias de relato falso' },
  { id: 'm2', type: 'alag', title: 'Alagamento na Rua F', by: 'Usuário 5 · conta nova', why: 'Autor sem histórico' },
];

export const notifications = [
  { type: 'alag' as HazardType, title: 'Novo relato a 380 m', sub: 'Alagamento: ajude a confirmar', ago: '12 min' },
  { type: 'inc' as HazardType, title: 'Incêndio validado', sub: 'Parque A · 1,8 km', ago: '1 h' },
  { type: 'temp' as HazardType, title: 'Relato finalizado', sub: 'Árvore removida na Av. D', ago: 'ontem' },
];
