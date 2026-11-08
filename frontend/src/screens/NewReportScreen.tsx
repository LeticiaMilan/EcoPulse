import React, { useEffect, useState } from 'react';
import { KeyboardAvoidingView, Pressable, ScrollView, Text, TextInput, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import * as Location from 'expo-location';
import { c, f, t, r, hazard, HAZARD_TYPES, HazardType } from '../theme';
import { useApp } from '../store/AppContext';
import { NEED_VOTES, Report, Severity, USER } from '../data/mock';
import { Button, Icon, ScreenHeader, SectionLabel, Segmented } from '../components/ui';
import type { RootNav } from '../navigation/types';

export default function NewReportScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const [type, setType] = useState<HazardType>('alag');
  const [sev, setSev] = useState<Severity>('Média');
  const [desc, setDesc] = useState('');
  const [focus, setFocus] = useState(false);
  const [done, setDone] = useState<Report | null>(null);
  const [loc, setLoc] = useState({ lat: USER.lat, lng: USER.lng, acc: 6, addr: 'R. Glicério, 210 · Liberdade' });

  useEffect(() => {
    (async () => {
      try {
        const { status } = await Location.requestForegroundPermissionsAsync();
        if (status !== 'granted') return;
        const p = await Location.getCurrentPositionAsync({ accuracy: Location.Accuracy.High });
        const [g] = await Location.reverseGeocodeAsync(p.coords).catch(() => []);
        setLoc({ lat: p.coords.latitude, lng: p.coords.longitude, acc: Math.round(p.coords.accuracy ?? 10), addr: g ? `${g.street ?? ''}${g.streetNumber ? ', ' + g.streetNumber : ''} · ${g.district ?? g.city ?? ''}` : loc.addr });
      } catch {}
    })();
  }, []);

  const submit = () => {
    if (!app.authed) return nav.navigate('Login');
    setDone(app.submitReport({ type, severity: sev, description: desc || hazard[type].name + ' relatado', lat: loc.lat, lng: loc.lng, addr: loc.addr }));
  };

  if (done) {
    const off = done.sync === 'pending';
    return (
      <SafeAreaView style={{ flex: 1, backgroundColor: c.bg }}>
        <ScreenHeader title="Novo relato" icon="close" onBack={() => nav.goBack()} />
        <View style={{ flex: 1, alignItems: 'center', justifyContent: 'center', gap: 16, padding: 24 }}>
          <View style={{ width: 88, height: 88, borderRadius: 30, alignItems: 'center', justifyContent: 'center', backgroundColor: off ? c.offBg : c.container }}>
            <Icon name={off ? 'cloud-upload' : 'check'} size={48} color={off ? c.offFg : c.onContainer} />
          </View>
          <Text style={[t.titleMd, { textAlign: 'center' }]}>{off ? 'Relato salvo no aparelho' : 'Relato enviado!'}</Text>
          <Text style={[t.body, { color: c.ink2, textAlign: 'center', maxWidth: 280 }]}>
            {off ? 'Assim que a internet voltar, ele é enviado sozinho. Você pode continuar usando o app normalmente.' : `Ele aparece para a comunidade como "aguardando validação" até receber ${NEED_VOTES} confirmações.`}
          </Text>
          <Text style={[t.mono, { backgroundColor: c.subtle, borderRadius: 10, paddingHorizontal: 12, paddingVertical: 8, overflow: 'hidden' }]}>
            {off ? `outbox · ${done.id} · aguardando rede` : `ReportCreated · ${done.id} · entregue`}
          </Text>
          <View style={{ alignSelf: 'stretch', gap: 8, marginTop: 12 }}>
            <Button label="Ver relato" onPress={() => nav.replace('Detail')} />
            <Button label="Voltar ao mapa" variant="ghost" size="md" onPress={() => nav.goBack()} />
          </View>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: c.bg }}>
      <ScreenHeader title="Novo relato" icon="close" onBack={() => nav.goBack()} />
      <KeyboardAvoidingView style={{ flex: 1 }} behavior="padding">
        <ScrollView contentContainerStyle={{ paddingHorizontal: 20, paddingBottom: 20, gap: 20 }} keyboardShouldPersistTaps="handled">
          {app.offline && (
            <View style={{ flexDirection: 'row', gap: 10, backgroundColor: c.offBg, borderRadius: 16, paddingHorizontal: 14, paddingVertical: 12 }}>
              <Icon name="cloud-off" size={20} color={c.offFg} />
              <Text style={{ flex: 1, fontFamily: f.medium, fontSize: 13, lineHeight: 19, color: c.offFg }}>Você está sem internet. O relato fica salvo no aparelho e é enviado sozinho quando a conexão voltar.</Text>
            </View>
          )}

          <View style={{ gap: 10 }}>
            <SectionLabel>O que está acontecendo?</SectionLabel>
            <View style={{ flexDirection: 'row', flexWrap: 'wrap', gap: 10 }}>
              {HAZARD_TYPES.map(k => {
                const on = k === type, h = hazard[k];
                return (
                  <Pressable key={k} onPress={() => setType(k)} style={{ width: '48%', height: 96, borderRadius: 18, borderWidth: on ? 2.5 : 1.5, borderColor: on ? h.solid : c.border, backgroundColor: on ? h.bg + '66' : c.surface, padding: 12, justifyContent: 'space-between' }}>
                    <View style={{ width: 38, height: 38, borderRadius: 12, alignItems: 'center', justifyContent: 'center', backgroundColor: on ? h.solid : h.bg }}>
                      <Icon name={h.icon} size={22} color={on ? '#fff' : h.fg} />
                    </View>
                    <Text style={t.bodyStrong}>{h.name}</Text>
                  </Pressable>
                );
              })}
            </View>
          </View>

          <View style={{ gap: 10 }}>
            <SectionLabel>Local</SectionLabel>
            <View style={{ backgroundColor: c.surface, borderWidth: 1.5, borderColor: c.border, borderRadius: r.input, padding: 12, paddingHorizontal: 14, flexDirection: 'row', gap: 12, alignItems: 'center' }}>
              <View style={{ width: 40, height: 40, borderRadius: 12, backgroundColor: c.syncBg, alignItems: 'center', justifyContent: 'center' }}><Icon name="my-location" color={c.location} /></View>
              <View style={{ flex: 1, gap: 3 }}>
                <Text style={t.bodyStrong} numberOfLines={1}>{loc.addr}</Text>
                <Text style={t.mono}>{loc.lat.toFixed(5)}, {loc.lng.toFixed(5)} · ±{loc.acc} m · GPS</Text>
              </View>
              <Text style={{ fontFamily: f.bold, fontSize: 13, color: c.primaryStrong }}>Ajustar</Text>
            </View>
          </View>

          <View style={{ gap: 10 }}>
            <SectionLabel>Gravidade</SectionLabel>
            <Segmented options={(['Baixa', 'Média', 'Alta'] as Severity[]).map(k => ({ key: k, label: k }))} value={sev} onChange={setSev} />
          </View>

          <View style={{ gap: 10 }}>
            <SectionLabel>Descreva rapidamente</SectionLabel>
            <TextInput value={desc} onChangeText={setDesc} multiline onFocus={() => setFocus(true)} onBlur={() => setFocus(false)}
              placeholder="Ex.: água na altura do joelho, carros não passam" placeholderTextColor={c.ink3}
              style={{ minHeight: 88, borderRadius: r.input, borderWidth: 1.5, borderColor: focus ? c.primary : c.border, backgroundColor: c.surface, padding: 14, paddingTop: 12, fontFamily: f.medium, fontSize: 15, color: c.ink, textAlignVertical: 'top' }} />
          </View>

          <Pressable onPress={() => app.flash('Câmera: integrar expo-image-picker')} style={{ height: 64, borderRadius: r.input, borderWidth: 1.5, borderStyle: 'dashed', borderColor: c.disabled, flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
            <Icon name="add-a-photo" color={c.ink2} />
            <Text style={{ fontFamily: f.semibold, fontSize: 14, color: c.ink2 }}>Adicionar foto (opcional)</Text>
          </Pressable>
        </ScrollView>
        <View style={{ paddingHorizontal: 20, paddingTop: 12, paddingBottom: 16, borderTopWidth: 1, borderTopColor: c.divider, gap: 8 }}>
          <Button label={app.offline ? 'Salvar no aparelho' : 'Enviar relato'} icon={app.offline ? 'save' : 'send'} onPress={submit} />
          {!app.authed && <Text style={[t.caption, { textAlign: 'center' }]}>Você precisa entrar para enviar relatos.</Text>}
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
