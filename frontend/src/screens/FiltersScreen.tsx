import React from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { c, f, t, r, hazard, HAZARD_TYPES, Status } from '../theme';
import { useApp } from '../store/AppContext';
import { Button, FilterChip, SectionLabel, Segmented } from '../components/ui';

const STATUSES: [Status, string][] = [['pend', 'Aguardando validação'], ['val', 'Ativo'], ['fin', 'Finalizado']];

export default function FiltersScreen() {
  const nav = useNavigation();
  const insets = useSafeAreaInsets();
  const { filters, setFilters, clearFilters, visibleReports } = useApp();

  return (
    <View style={{ flex: 1, justifyContent: 'flex-end' }}>
      <Pressable onPress={() => nav.goBack()} style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(14,17,22,.38)' }} />
      <View style={{ backgroundColor: c.bg, borderTopLeftRadius: r.sheet, borderTopRightRadius: r.sheet, maxHeight: '92%', paddingBottom: insets.bottom + 18 }}>
        <View style={{ alignSelf: 'center', width: 36, height: 4, borderRadius: 2, backgroundColor: c.disabled, marginTop: 10 }} />
        <ScrollView contentContainerStyle={{ padding: 20, paddingTop: 14, gap: 18 }}>
          <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
            <Text style={t.titleSm}>Filtros</Text>
            <Pressable onPress={clearFilters} hitSlop={10}><Text style={{ fontFamily: f.bold, fontSize: 14, color: c.primaryStrong }}>Limpar</Text></Pressable>
          </View>

          <View style={{ gap: 10 }}>
            <SectionLabel>Tipo de risco</SectionLabel>
            <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }}>
              {HAZARD_TYPES.map(k => {
                const on = filters.types[k];
                return (
                  <Pressable key={k} onPress={() => setFilters({ ...filters, types: { ...filters.types, [k]: !on } })}
                    style={{ width: '48.5%', height: 46, borderRadius: 14, borderWidth: on ? 2 : 1.5, borderColor: on ? hazard[k].solid : c.border, backgroundColor: on ? c.surface : 'transparent', flexDirection: 'row', alignItems: 'center', gap: 10, paddingHorizontal: 12 }}>
                    <View style={{ width: 14, height: 14, borderRadius: 5, backgroundColor: on ? hazard[k].solid : c.disabled }} />
                    <Text style={{ fontFamily: f.bold, fontSize: 14, color: c.ink }}>{hazard[k].name}</Text>
                  </Pressable>
                );
              })}
            </View>
          </View>

          <View style={{ gap: 10 }}>
            <SectionLabel>Situação</SectionLabel>
            <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 8 }}>
              {STATUSES.map(([k, l]) => (
                <FilterChip key={k} label={l} active={filters.statuses[k]} onPress={() => setFilters({ ...filters, statuses: { ...filters.statuses, [k]: !filters.statuses[k] } })} />
              ))}
            </View>
          </View>

          <View style={{ gap: 10 }}>
            <SectionLabel>Distância de você</SectionLabel>
            <Segmented options={[1, 2, 5, 10].map(k => ({ key: k, label: `${k} km` }))} value={filters.radiusKm} onChange={k => setFilters({ ...filters, radiusKm: k })} />
          </View>

          <View style={{ gap: 10 }}>
            <SectionLabel>Período</SectionLabel>
            <Segmented options={[{ key: '24h', label: '24 horas' }, { key: '7d', label: '7 dias' }, { key: '30d', label: '30 dias' }] as const as any} value={filters.period} onChange={(k: any) => setFilters({ ...filters, period: k })} />
          </View>

          <Button label={`Mostrar ${visibleReports.length} relatos`} onPress={() => nav.goBack()} />
        </ScrollView>
      </View>
    </View>
  );
}
