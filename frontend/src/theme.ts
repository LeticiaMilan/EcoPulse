// EcoPulse · paleta "Moderno" (valores HEX convertidos do design)
export const c = {
  bg: '#f6f7f9',
  surface: '#ffffff',
  nav: '#ffffff',
  subtle: '#f3f4f7',
  sunken: '#eef0f3',
  canvas: '#eceef2',
  border: '#dfe2e8',
  divider: '#e8eaee',
  disabled: '#cdd1d9',
  ink: '#0e1116',
  ink2: '#5b6270',
  ink3: '#858b96',
  inkInverse: '#f6f7f9',

  primary: '#008145',
  primaryStrong: '#006734',
  container: '#b1efca',
  onContainer: '#004318',
  subtleBrand: '#d5f9e3',
  accent: '#4bdc95',

  pendBg: '#fee9be',
  pendFg: '#714000',
  valBg: '#c6f4d7',
  valFg: '#005f2e',
  finBg: '#eaecf0',
  progressPend: '#d09100',

  offBg: '#ffe2a8',
  offFg: '#783b00',
  offTrack: '#bc7300',
  syncBg: '#d1e9ff',
  syncFg: '#1d4587',
  danger: '#ce232c',
  dangerText: '#a82428',
  dangerBg: '#ffe0db',
  location: '#226adf',

  alertBg: '#ffe4d9',
  alertBorder: '#ffb7a3',
  alertFg: '#942c11',

  mapPark: '#d2f1dc',
  mapWater: '#c6e0f8',
  mapBase: '#eceff3',
};

export type HazardType = 'alag' | 'temp' | 'inc' | 'desl';
export type Status = 'pend' | 'val' | 'fin';

export const hazard: Record<HazardType, { name: string; icon: any; solid: string; bg: string; fg: string }> = {
  alag: { name: 'Alagamento', icon: 'flood', solid: '#007cdc', bg: '#cdefff', fg: '#0055aa' },
  temp: { name: 'Tempestade', icon: 'thunderstorm', solid: '#855dd8', bg: '#eee4ff', fg: '#5f38a6' },
  inc: { name: 'Incêndio', icon: 'local-fire-department', solid: '#d33d34', bg: '#ffdbd3', fg: '#a01111' },
  desl: { name: 'Deslizamento', icon: 'landslide', solid: '#b26600', bg: '#ffe6c0', fg: '#854200' },
};
export const HAZARD_TYPES = Object.keys(hazard) as HazardType[];

export const status: Record<Status, { label: string; bg: string; fg: string }> = {
  pend: { label: 'Aguardando validação', bg: c.pendBg, fg: c.pendFg },
  val: { label: 'Ativo · validado', bg: c.valBg, fg: c.valFg },
  fin: { label: 'Finalizado', bg: c.finBg, fg: c.ink2 },
};

// AQI: 0 bom · 1 moderado · 2 sensíveis · 3 ruim
export const aqi = [
  { label: 'Bom', bg: '#b4eec7', fg: '#004922', border: '#349d62' },
  { label: 'Moderado', bg: '#eedea1', fg: '#4b3800', border: '#9f8400' },
  { label: 'Sensíveis', bg: '#ffd2a9', fg: '#5e2a00', border: '#be7125' },
  { label: 'Ruim', bg: '#ffcac3', fg: '#65201e', border: '#c8635d' },
];
export const aqiLevel = (v: number) => (v <= 50 ? 0 : v <= 100 ? 1 : v <= 150 ? 2 : 3);

// Fontes (no Android, fontWeight não funciona com fontes customizadas — use a família do peso)
export const f = {
  regular: 'Figtree_400Regular',
  medium: 'Figtree_500Medium',
  semibold: 'Figtree_600SemiBold',
  bold: 'Figtree_700Bold',
  extrabold: 'Figtree_800ExtraBold',
  mono: 'IBMPlexMono_500Medium',
};

export const t = {
  display: { fontFamily: f.extrabold, fontSize: 30, lineHeight: 35, letterSpacing: -0.6, color: c.ink },
  titleLg: { fontFamily: f.extrabold, fontSize: 26, lineHeight: 31, letterSpacing: -0.3, color: c.ink },
  titleMd: { fontFamily: f.extrabold, fontSize: 24, lineHeight: 29, letterSpacing: -0.2, color: c.ink },
  titleSm: { fontFamily: f.extrabold, fontSize: 20, lineHeight: 25, color: c.ink },
  headline: { fontFamily: f.extrabold, fontSize: 18, lineHeight: 23, color: c.ink },
  bodyStrong: { fontFamily: f.bold, fontSize: 16, lineHeight: 22, color: c.ink },
  body: { fontFamily: f.medium, fontSize: 15, lineHeight: 22, color: c.ink },
  bodySm: { fontFamily: f.medium, fontSize: 14, lineHeight: 20, color: c.ink },
  label: { fontFamily: f.bold, fontSize: 13, lineHeight: 17, color: c.ink2 },
  caption: { fontFamily: f.semibold, fontSize: 12, lineHeight: 16, color: c.ink3 },
  overline: { fontFamily: f.bold, fontSize: 11, lineHeight: 13, color: c.ink2 },
  mono: { fontFamily: f.mono, fontSize: 11, lineHeight: 15, color: c.ink3 },
};

export const r = { chip: 8, iconSm: 12, segmented: 14, input: 16, fab: 18, card: 20, cardLg: 22, sheet: 28, pill: 999 };
export const s = { 4: 4, 6: 6, 8: 8, 10: 10, 12: 12, 14: 14, 16: 16, 20: 20, 24: 24, 32: 32 };

export const shadow = {
  card: { shadowColor: '#0f1423', shadowOpacity: 0.06, shadowRadius: 2, shadowOffset: { width: 0, height: 1 }, elevation: 1 },
  subtle: { shadowColor: '#0f1423', shadowOpacity: 0.15, shadowRadius: 3, shadowOffset: { width: 0, height: 1 }, elevation: 2 },
  floating: { shadowColor: '#0f1423', shadowOpacity: 0.14, shadowRadius: 10, shadowOffset: { width: 0, height: 2 }, elevation: 4 },
  raised: { shadowColor: '#0f1423', shadowOpacity: 0.22, shadowRadius: 20, shadowOffset: { width: 0, height: 6 }, elevation: 8 },
};
