import React, { useEffect, useState } from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { c, f, t, hazard, shadow } from '../theme';
import { useApp } from '../store/AppContext';
import { Button, Icon, Toggle } from '../components/ui';
import type { RootNav } from '../navigation/types';
import * as Location from 'expo-location';

export default function ProfileScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const [city, setCity] = useState('Ribeirão Preto, SP');
  useEffect(() => {
    (async () => {
      try {
        const { status } = await Location.getForegroundPermissionsAsync();
        if (status !== 'granted') return;
        const p = await Location.getLastKnownPositionAsync();
        if (!p) return;
        const [g] = await Location.reverseGeocodeAsync(p.coords);
        if (g?.city || g?.subregion) setCity(`${g.city ?? g.subregion}${g.region ? ', ' + g.region : ''}`);
      } catch {}
    })();
  }, []);
  const pend = app.pendingCount;

  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: c.bg }}>
      <ScrollView contentContainerStyle={{ padding: 16, paddingTop: 12, gap: 16 }}>
        <View style={{ flexDirection: 'row', gap: 14, alignItems: 'center', padding: 4, paddingVertical: 8 }}>
          <View style={{ width: 60, height: 60, borderRadius: 22, backgroundColor: c.container, alignItems: 'center', justifyContent: 'center' }}>
            <Text style={{ fontFamily: f.extrabold, fontSize: 22, color: c.onContainer }}>{app.user ? app.user.username.slice(0, 2).toUpperCase() : '?'}</Text>
          </View>
          <View style={{ gap: 4 }}>
            <Text style={t.titleSm}>{app.user ? app.user.username : 'Visitante'}</Text>
            <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13, color: c.ink2 }]}>{app.user ? `ID ${app.user.id.slice(0, 8)} · ${city}` : 'Só consulta do mapa'}</Text>
          </View>
        </View>

        {!app.authed && <Button label="Entrar para relatar e votar" onPress={() => nav.navigate('Login')} />}

        <View style={[{ backgroundColor: c.surface, borderRadius: 22, paddingVertical: 4 }, shadow.card]}>
          <Pressable onPress={() => nav.navigate('Sync')} style={{ flexDirection: 'row', alignItems: 'center', gap: 14, height: 58, paddingHorizontal: 18 }}>
            <Icon name="sync" color={c.ink2} />
            <Text style={{ flex: 1, fontFamily: f.semibold, fontSize: 15, color: c.ink }}>Sincronização</Text>
            {pend > 0 && (
              <View style={{ minWidth: 22, height: 22, paddingHorizontal: 7, borderRadius: 11, backgroundColor: c.offBg, alignItems: 'center', justifyContent: 'center' }}>
                <Text style={{ fontFamily: f.extrabold, fontSize: 12, color: c.offFg }}>{pend}</Text>
              </View>
            )}
            <Icon name="chevron-right" size={20} color={c.ink3} />
          </Pressable>
          <Pressable onPress={() => app.setSimOffline(!app.simOffline)} style={{ flexDirection: 'row', alignItems: 'center', gap: 14, height: 58, paddingHorizontal: 18, borderTopWidth: 1, borderTopColor: c.divider }}>
            <Icon name="wifi-off" color={c.ink2} />
            <View style={{ flex: 1 }}>
              <Text style={{ fontFamily: f.semibold, fontSize: 15, color: c.ink }}>Simular sem internet</Text>
              <Text style={[t.caption, { fontSize: 11 }]}>Modo demonstração</Text>
            </View>
            <Toggle value={app.simOffline} color={c.offTrack} />
          </Pressable>
        </View>

        {false && app.authed && (
          <View style={{ gap: 8 }}>
            <View style={{ flexDirection: 'row', justifyContent: 'space-between', paddingHorizontal: 4 }}>
              <Text style={t.label}>Moderação</Text>
              <Text style={[t.caption, { fontFamily: f.bold }]}>{app.modQueue.length} na fila</Text>
            </View>
            {app.modQueue.map(m => (
              <View key={m.id} style={[{ backgroundColor: c.surface, borderRadius: 20, padding: 14, gap: 12 }, shadow.card]}>
                <View style={{ flexDirection: 'row', gap: 10 }}>
                  <View style={{ width: 12, height: 12, borderRadius: 4, marginTop: 5, backgroundColor: hazard[m.type].solid }} />
                  <View style={{ flex: 1, gap: 3 }}>
                    <Text style={[t.bodyStrong, { fontSize: 15 }]}>{m.title}</Text>
                    <Text style={[t.caption, { fontFamily: f.medium, color: c.ink2 }]}>{m.by} · {m.why}</Text>
                  </View>
                </View>
                <View style={{ flexDirection: 'row', gap: 8 }}>
                  <Pressable onPress={() => app.moderate(m.id, true)} style={{ flex: 1, height: 42, borderRadius: 21, backgroundColor: c.container, alignItems: 'center', justifyContent: 'center' }}>
                    <Text style={{ fontFamily: f.bold, fontSize: 14, color: c.onContainer }}>Aprovar</Text>
                  </Pressable>
                  <Pressable onPress={() => app.moderate(m.id, false)} style={{ flex: 1, height: 42, borderRadius: 21, backgroundColor: c.dangerBg, alignItems: 'center', justifyContent: 'center' }}>
                    <Text style={{ fontFamily: f.bold, fontSize: 14, color: c.dangerText }}>Reprovar</Text>
                  </Pressable>
                </View>
              </View>
            ))}
            {!app.modQueue.length && <Text style={[t.bodySm, { color: c.ink2, textAlign: 'center', padding: 14 }]}>Fila de moderação vazia.</Text>}
          </View>
        )}

        {app.authed && (
          <Pressable onPress={() => { app.logout(); nav.reset({ index: 0, routes: [{ name: 'Login' }] }); }} style={{ height: 48, alignItems: 'center', justifyContent: 'center' }}>
            <Text style={{ fontFamily: f.bold, fontSize: 15, color: c.dangerText }}>Sair da conta</Text>
          </Pressable>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}
