import React from 'react';
import { Pressable, Text, View, StyleProp, ViewStyle } from 'react-native';
import { MaterialIcons } from '@expo/vector-icons';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { c, f, t, r, shadow, hazard, status as statusTok, HazardType, Status } from '../theme';
import { useApp } from '../store/AppContext';

export type IconName = React.ComponentProps<typeof MaterialIcons>['name'];
export const Icon = ({ name, size = 22, color = c.ink }: { name: IconName; size?: number; color?: string }) => <MaterialIcons name={name} size={size} color={color} />;

type BtnVariant = 'primary' | 'outline' | 'ghost' | 'dark' | 'secondary';
const btnColors: Record<BtnVariant | 'muted', { bg: string; fg: string; border?: string }> = {
  primary: { bg: c.primary, fg: '#fff' },
  outline: { bg: c.surface, fg: c.primaryStrong, border: c.primary },
  ghost: { bg: 'transparent', fg: c.primaryStrong },
  dark: { bg: c.ink, fg: c.inkInverse },
  secondary: { bg: c.surface, fg: c.ink, border: c.border },
  muted: { bg: c.sunken, fg: c.ink3 },
};

export function Button({ label, icon, onPress, variant = 'primary', size = 'lg', disabled, style }: {
  label: string; icon?: IconName; onPress?: () => void; variant?: BtnVariant; size?: 'lg' | 'md' | 'sm'; disabled?: boolean; style?: StyleProp<ViewStyle>;
}) {
  const v = btnColors[disabled ? 'muted' : variant];
  const h = size === 'lg' ? 54 : size === 'md' ? 46 : 42;
  return (
    <Pressable
      onPress={disabled ? undefined : onPress}
      style={({ pressed }) => [{
        height: h, borderRadius: h / 2, backgroundColor: v.bg, borderWidth: v.border ? 1.5 : 0, borderColor: v.border,
        flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 8, paddingHorizontal: 18,
        transform: [{ scale: pressed ? 0.98 : 1 }],
      }, style]}
    >
      {icon && <Icon name={icon} size={20} color={v.fg} />}
      <Text style={{ fontFamily: f.bold, fontSize: size === 'lg' ? 16 : 14, color: v.fg }}>{label}</Text>
    </Pressable>
  );
}

export function IconButton({ icon, onPress, variant = 'floating', size = 44, children }: {
  icon: IconName; onPress?: () => void; variant?: 'floating' | 'sunken' | 'plain'; size?: number; children?: React.ReactNode;
}) {
  const bg = variant === 'floating' ? c.surface : variant === 'sunken' ? c.sunken : 'transparent';
  return (
    <Pressable onPress={onPress} hitSlop={6} style={[{ width: size, height: size, borderRadius: size / 2, backgroundColor: bg, alignItems: 'center', justifyContent: 'center' }, variant === 'floating' && shadow.floating]}>
      <Icon name={icon} size={23} />
      {children}
    </Pressable>
  );
}

export function LayerChip({ label, icon, active, onPress }: { label: string; icon: IconName; active: boolean; onPress: () => void }) {
  return (
    <Pressable onPress={onPress} style={[{ height: 36, paddingHorizontal: 14, borderRadius: 18, flexDirection: 'row', alignItems: 'center', gap: 6, backgroundColor: active ? c.ink : c.surface }, shadow.floating]}>
      <Icon name={icon} size={17} color={active ? c.inkInverse : c.ink} />
      <Text style={{ fontFamily: f.bold, fontSize: 13, color: active ? c.inkInverse : c.ink }}>{label}</Text>
    </Pressable>
  );
}

export function FilterChip({ label, active, onPress }: { label: string; active: boolean; onPress: () => void }) {
  return (
    <Pressable onPress={onPress} style={{ height: 36, paddingHorizontal: 14, borderRadius: 18, justifyContent: 'center', backgroundColor: active ? c.ink : 'transparent', borderWidth: active ? 0 : 1.5, borderColor: c.border }}>
      <Text style={{ fontFamily: f.bold, fontSize: 13, color: active ? c.inkInverse : c.ink }}>{label}</Text>
    </Pressable>
  );
}

export function StatusChip({ status }: { status: Status }) {
  const s = statusTok[status];
  return <Text style={{ fontFamily: f.bold, fontSize: 11, paddingHorizontal: 8, paddingVertical: 3, borderRadius: r.chip, overflow: 'hidden', backgroundColor: s.bg, color: s.fg }}>{s.label}</Text>;
}

export function HazardIcon({ type, size = 46, solid }: { type: HazardType; size?: number; solid?: boolean }) {
  const h = hazard[type];
  return (
    <View style={{ width: size, height: size, borderRadius: size * 0.3, backgroundColor: solid ? h.solid : h.bg, alignItems: 'center', justifyContent: 'center' }}>
      <Icon name={h.icon} size={size * 0.52} color={solid ? '#fff' : h.fg} />
    </View>
  );
}

export function HazardPill({ type }: { type: HazardType }) {
  const h = hazard[type];
  return (
    <View style={{ flexDirection: 'row', alignItems: 'center', gap: 5, paddingLeft: 7, paddingRight: 10, paddingVertical: 4, borderRadius: 9, backgroundColor: h.bg }}>
      <Icon name={h.icon} size={16} color={h.fg} />
      <Text style={{ fontFamily: f.extrabold, fontSize: 12, color: h.fg }}>{h.name}</Text>
    </View>
  );
}

export function Segmented<T extends string | number>({ options, value, onChange }: { options: { key: T; label: string }[]; value: T; onChange: (k: T) => void }) {
  return (
    <View style={{ flexDirection: 'row', backgroundColor: c.sunken, borderRadius: r.segmented, padding: 4, gap: 4 }}>
      {options.map(o => {
        const on = o.key === value;
        return (
          <Pressable key={String(o.key)} onPress={() => onChange(o.key)} style={[{ flex: 1, height: 40, borderRadius: 11, alignItems: 'center', justifyContent: 'center' }, on && { backgroundColor: c.surface, ...shadow.subtle }]}>
            <Text style={{ fontFamily: f.bold, fontSize: 14, color: on ? c.ink : c.ink2 }}>{o.label}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

export function Toggle({ value, color = c.primary }: { value: boolean; color?: string }) {
  return (
    <View style={{ width: 44, height: 26, borderRadius: 13, backgroundColor: value ? color : c.disabled, justifyContent: 'center' }}>
      <View style={[{ position: 'absolute', left: value ? 21 : 3, width: 20, height: 20, borderRadius: 10, backgroundColor: '#fff' }, shadow.subtle]} />
    </View>
  );
}

export const Card = ({ children, style }: { children: React.ReactNode; style?: StyleProp<ViewStyle> }) => (
  <View style={[{ backgroundColor: c.surface, borderRadius: r.cardLg, padding: 18, gap: 14 }, shadow.card, style]}>{children}</View>
);

export const SectionLabel = ({ children }: { children: React.ReactNode }) => <Text style={t.label}>{children}</Text>;

export function ScreenHeader({ title, onBack, icon = 'arrow-back' }: { title: string; onBack: () => void; icon?: IconName }) {
  return (
    <View style={{ height: 60, flexDirection: 'row', alignItems: 'center', gap: 6, paddingHorizontal: 8 }}>
      <IconButton icon={icon} variant="plain" onPress={onBack} />
      <Text style={t.titleSm}>{title}</Text>
    </View>
  );
}

export function ConnectionPill() {
  const { offline, syncing, pendingCount, conflict, simOffline, setSimOffline } = useApp();
  const n = pendingCount - (conflict === 'open' ? 1 : 0);
  const [bg, fg] = offline ? [c.offBg, c.offFg] : syncing ? [c.syncBg, c.syncFg] : [c.surface, c.primaryStrong];
  const label = syncing ? 'Sincronizando relatos…' : offline ? `Offline · ${n} ${n === 1 ? 'item' : 'itens'} no aparelho` : 'Online · tudo sincronizado';
  return (
    <Pressable onPress={() => setSimOffline(!simOffline)} style={[{ alignSelf: 'flex-start', height: 30, paddingHorizontal: 12, borderRadius: 15, flexDirection: 'row', alignItems: 'center', gap: 7, backgroundColor: bg }, shadow.floating]}>
      {syncing ? <Icon name="sync" size={16} color={fg} /> : <View style={{ width: 8, height: 8, borderRadius: 4, backgroundColor: fg }} />}
      <Text style={{ fontFamily: f.bold, fontSize: 12, color: fg }}>{label}</Text>
    </Pressable>
  );
}

export function Toast() {
  const { toast } = useApp();
  const insets = useSafeAreaInsets();
  if (!toast) return null;
  return (
    <View pointerEvents="none" style={[{ position: 'absolute', left: 16, right: 16, bottom: insets.bottom + 90, backgroundColor: c.ink, borderRadius: 16, paddingHorizontal: 16, paddingVertical: 14, flexDirection: 'row', alignItems: 'center', gap: 10 }, shadow.raised]}>
      <Icon name="check-circle" size={20} color={c.accent} />
      <Text style={{ fontFamily: f.medium, fontSize: 14, color: c.inkInverse, flex: 1 }}>{toast}</Text>
    </View>
  );
}
