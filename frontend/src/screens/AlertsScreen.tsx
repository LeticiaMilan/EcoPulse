import React, { useEffect, useState } from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import Slider from '@react-native-community/slider';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useIsFocused, useNavigation } from '@react-navigation/native';
import * as Location from 'expo-location';
import { c, f, t, hazard, HAZARD_TYPES, shadow } from '../theme';
import { useApp } from '../store/AppContext';
import type { Report } from '../data/mock';
import { Button, Card, HazardIcon, Toggle } from '../components/ui';
import type { RootNav } from '../navigation/types';

export default function AlertsScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const isFocused = useIsFocused();
  const [locState, setLocState] = useState<'loading' | 'ok' | 'denied'>(app.here ? 'ok' : 'loading');

  // Usa a localização real do aparelho (atualiza sempre que a aba é aberta)
  useEffect(() => {
    if (!isFocused) return;
    (async () => {
      try {
        const { status } = await Location.requestForegroundPermissionsAsync();
        if (status !== 'granted') { if (!app.here) setLocState('denied'); return; }
        // Mostra rápido com a última posição conhecida e depois troca pela posição atual,
        // porque a última conhecida pode ser de outro lugar (ex.: de horas atrás)
        const last = await Location.getLastKnownPositionAsync();
        if (last) { app.setHere({ lat: last.coords.latitude, lng: last.coords.longitude }); setLocState('ok'); }
        const now = await Location.getCurrentPositionAsync({ accuracy: Location.Accuracy.Balanced });
        app.setHere({ lat: now.coords.latitude, lng: now.coords.longitude });
        setLocState('ok');
      } catch {
        if (!app.here) setLocState('denied');
      }
    })();
  }, [isFocused]);

  // Relatos dentro do raio de alerta (calculados no AppContext, que também conta o número do sininho)
  const nearby = app.nearbyAlerts;

  // Com a aba aberta, tudo o que aparece aqui conta como visto (inclusive o que chegar enquanto ela está aberta)
  useEffect(() => { if (isFocused) app.markAlertsSeen(); }, [isFocused, nearby]);

  // Risco ativo = relato validado (de preferência dentro de uma área afetada ativa)
  const active = nearby.filter(x => x.r.status === 'val');
  const top = active.find(x => x.r.area?.status === 'active') ?? active[0];
  const recent = [...nearby].sort((a, b) => (b.r.datetime ?? '').localeCompare(a.r.datetime ?? '')).slice(0, 6);

  const fmt = (m: number) => (m < 1000 ? `${Math.round(m / 10) * 10} m` : `${(m / 1000).toFixed(1).replace('.', ',')} km`);
  const openOnMap = (r: Report) => { app.focusOn(r.id); nav.navigate('Tabs', { screen: 'Map' }); };
  const radiusText = `${String(app.alertRadius).replace('.', ',')} km`;

  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: c.bg }}>
      <ScrollView contentContainerStyle={{ padding: 16, paddingTop: 12, gap: 16 }}>
        <Text style={[t.titleLg, { paddingHorizontal: 4 }]}>Alertas</Text>

        {top ? (
          <View style={{ backgroundColor: c.alertBg, borderRadius: 24, padding: 18, gap: 12, borderWidth: 1.5, borderColor: c.alertBorder }}>
            <View style={{ flexDirection: 'row', gap: 12, alignItems: 'center' }}>
              <HazardIcon type={top.r.type} size={44} solid />
              <View style={{ gap: 2, flex: 1 }}>
                <Text style={{ fontFamily: f.extrabold, fontSize: 12, color: c.alertFg, textTransform: 'uppercase', letterSpacing: 0.7 }}>
                  {active.length > 1 ? `${active.length} riscos ativos perto de você` : 'Risco ativo perto de você'}
                </Text>
                <Text style={[t.headline, { fontSize: 17 }]}>{hazard[top.r.type].name} a {fmt(top.m)}</Text>
              </View>
            </View>
            <Text style={[t.bodySm, { color: c.ink2 }]}>
              {top.r.title}. {top.r.addr}. {top.r.area?.status === 'active' ? 'Área afetada confirmada pela comunidade.' : 'Relato validado pela comunidade.'}
            </Text>
            <Button size="md" variant="dark" label="Ver no mapa" onPress={() => openOnMap(top.r)} />
          </View>
        ) : (
          <View style={{ backgroundColor: c.container, borderRadius: 24, padding: 18, gap: 6 }}>
            <Text style={{ fontFamily: f.extrabold, fontSize: 12, color: c.onContainer, textTransform: 'uppercase', letterSpacing: 0.7 }}>
              {locState === 'denied' ? 'Localização desligada' : locState === 'loading' ? 'Buscando sua localização…' : 'Tudo tranquilo por aqui'}
            </Text>
            <Text style={[t.bodySm, { color: c.onContainer }]}>
              {locState === 'denied'
                ? 'Permita o acesso à localização para ver os riscos perto de você.'
                : locState === 'loading'
                ? 'Os alertas usam onde você está agora.'
                : `Nenhum risco validado num raio de ${radiusText}.`}
            </Text>
          </View>
        )}

        <Card style={{ gap: 12, paddingHorizontal: 18, paddingVertical: 16 }}>
          <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <Text style={[t.bodyStrong, { fontFamily: f.extrabold }]}>Me avise num raio de</Text>
            <Text style={[t.headline, { color: c.primaryStrong }]}>{String(app.alertRadius).replace('.', ',')} km</Text>
          </View>
          <Slider minimumValue={0.5} maximumValue={5} step={0.5} value={app.alertRadius} onValueChange={app.setAlertRadius}
            minimumTrackTintColor={c.primary} maximumTrackTintColor={c.disabled} thumbTintColor={c.primary} />
          <View>
            {HAZARD_TYPES.map(k => (
              <Pressable key={k} onPress={() => app.toggleAlertType(k)} style={{ flexDirection: 'row', alignItems: 'center', gap: 12, height: 48, borderTopWidth: 1, borderTopColor: c.divider }}>
                <View style={{ width: 12, height: 12, borderRadius: 4, backgroundColor: hazard[k].solid }} />
                <Text style={{ flex: 1, fontFamily: f.semibold, fontSize: 15, color: c.ink }}>{hazard[k].name}</Text>
                <Toggle value={app.alertTypes[k]} />
              </Pressable>
            ))}
          </View>
        </Card>

        <View style={{ gap: 8 }}>
          <Text style={[t.label, { paddingHorizontal: 4 }]}>Recentes num raio de {radiusText}</Text>
          {recent.map(({ r, m }) => (
            <Pressable key={r.id} onPress={() => openOnMap(r)} style={[{ flexDirection: 'row', gap: 12, padding: 14, paddingVertical: 12, backgroundColor: c.surface, borderRadius: 18 }, shadow.card]}>
              <View style={{ width: 10, height: 10, borderRadius: 5, marginTop: 5, backgroundColor: hazard[r.type].solid }} />
              <View style={{ flex: 1, gap: 2 }}>
                <Text style={[t.bodySm, { fontFamily: f.bold }]}>
                  {r.status === 'pend' ? `Novo relato a ${fmt(m)}` : r.status === 'val' ? `${hazard[r.type].name} validado` : r.rejected ? 'Relato rejeitado' : 'Relato finalizado'}
                </Text>
                <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13, color: c.ink2 }]} numberOfLines={1}>
                  {r.status === 'pend' ? `${hazard[r.type].name}: ajude a confirmar` : `${r.title} · ${fmt(m)}`}
                </Text>
              </View>
              <Text style={[t.caption, { fontFamily: f.medium }]}>{r.ago}</Text>
            </Pressable>
          ))}
          {!recent.length && (
            <Text style={[t.bodySm, { color: c.ink2, textAlign: 'center', padding: 16 }]}>
              {app.here ? 'Nenhum relato recente perto de você.' : 'Sem localização, não dá para saber o que está perto.'}
            </Text>
          )}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
