import React from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import Slider from '@react-native-community/slider';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { c, f, t, hazard, HAZARD_TYPES, shadow } from '../theme';
import { useApp } from '../store/AppContext';
import { notifications, initialReports } from '../data/mock';
import { Button, Card, HazardIcon, Toggle } from '../components/ui';
import type { RootNav } from '../navigation/types';

export default function AlertsScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const openAlert = () => { app.setSelectedId('r2'); nav.navigate('Tabs', { screen: 'Map' }); };

  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: c.bg }}>
      <ScrollView contentContainerStyle={{ padding: 16, paddingTop: 12, gap: 16 }}>
        <Text style={[t.titleLg, { paddingHorizontal: 4 }]}>Alertas</Text>

        <View style={{ backgroundColor: c.alertBg, borderRadius: 24, padding: 18, gap: 12, borderWidth: 1.5, borderColor: c.alertBorder }}>
          <View style={{ flexDirection: 'row', gap: 12, alignItems: 'center' }}>
            <HazardIcon type="desl" size={44} solid />
            <View style={{ gap: 2, flex: 1 }}>
              <Text style={{ fontFamily: f.extrabold, fontSize: 12, color: c.alertFg, textTransform: 'uppercase', letterSpacing: 0.7 }}>Risco ativo perto de você</Text>
              <Text style={[t.headline, { fontSize: 17 }]}>Deslizamento a 1,2 km</Text>
            </View>
          </View>
          <Text style={[t.bodySm, { color: c.ink2 }]}>Rua C, 820. Confirmado por 5 moradores. Evite a encosta e a Rua G.</Text>
          <View style={{ flexDirection: 'row', gap: 8 }}>
            <Button style={{ flex: 1 }} size="md" variant="dark" label="Ver no mapa" onPress={openAlert} />
            <Button style={{ flex: 1, backgroundColor: 'transparent' }} size="md" variant="secondary" icon="alt-route" label="Rota segura" onPress={openAlert} />
          </View>
        </View>

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
          <Text style={[t.label, { paddingHorizontal: 4 }]}>Recentes</Text>
          {notifications.map((n, i) => (
            <View key={i} style={[{ flexDirection: 'row', gap: 12, padding: 14, paddingVertical: 12, backgroundColor: c.surface, borderRadius: 18 }, shadow.card]}>
              <View style={{ width: 10, height: 10, borderRadius: 5, marginTop: 5, backgroundColor: hazard[n.type].solid }} />
              <View style={{ flex: 1, gap: 2 }}>
                <Text style={[t.bodySm, { fontFamily: f.bold }]}>{n.title}</Text>
                <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13, color: c.ink2 }]}>{n.sub}</Text>
              </View>
              <Text style={[t.caption, { fontFamily: f.medium }]}>{n.ago}</Text>
            </View>
          ))}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
