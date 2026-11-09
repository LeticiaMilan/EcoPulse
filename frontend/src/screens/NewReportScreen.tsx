import React, { useEffect, useRef, useState } from 'react';
import { KeyboardAvoidingView, Modal, Pressable, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import MapView, { Circle, PROVIDER_GOOGLE, Region } from 'react-native-maps';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import * as Location from 'expo-location';
import { c, f, t, r, hazard, HAZARD_TYPES, HazardType } from '../theme';
import { distanceM, useApp } from '../store/AppContext';
import { NEED_VOTES, Report, Severity, USER } from '../data/mock';
import { Button, Icon, IconButton, ScreenHeader, SectionLabel, Segmented } from '../components/ui';
import { mapStyle } from '../components/MapPins';
import type { RootNav } from '../navigation/types';

// ===================== AJUSTE MANUAL DO LOCAL DO RELATO =====================
// A ideia: por padrão o relato usa a posição do aparelho (GPS). Com o ajuste ligado,
// a pessoa toca em "Ajustar" e escolhe o local num mapa (pino fixo no centro), para
// relatar algo perto dela que não está exatamente onde ela está. O local escolhido
// precisa ficar a até MAX_DISTANCIA_AJUSTE_KM da posição real, para ninguém marcar
// relatos em lugares onde não está.
//
// >>> PARA TRAVAR O RELATO SÓ NA LOCALIZAÇÃO DO USUÁRIO: troque para `false`. <<<
// Com `false`, o "Ajustar" some, o mapa de escolha não abre e o relato usa sempre o GPS.
const PERMITIR_AJUSTE_DO_LOCAL = true;
// Distância máxima (em km) entre a posição do usuário e o local escolhido no mapa
const MAX_DISTANCIA_AJUSTE_KM = 5;
// ==============================================================================

export default function NewReportScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const [type, setType] = useState<HazardType>('alag');
  const [sev, setSev] = useState<Severity>('Média');
  const [desc, setDesc] = useState('');
  const [focus, setFocus] = useState(false);
  const [done, setDone] = useState<Report | null>(null);
  // source: 'none' = ainda sem local · 'gps' = posição do aparelho · 'map' = escolhido no mapa
  const start = app.here ?? { lat: USER.lat, lng: USER.lng };
  const [loc, setLoc] = useState<{ lat: number; lng: number; acc: number; addr: string; source: 'none' | 'gps' | 'map' }>(
    { lat: start.lat, lng: start.lng, acc: 0, addr: 'Buscando sua localização…', source: 'none' });
  const [picking, setPicking] = useState(false);
  const [pick, setPick] = useState({ lat: start.lat, lng: start.lng });
  const pickMapRef = useRef<MapView>(null);

  const addressOf = async (lat: number, lng: number) => {
    const [g] = await Location.reverseGeocodeAsync({ latitude: lat, longitude: lng }).catch(() => []);
    const street = g ? `${g.street ?? ''}${g.streetNumber ? ', ' + g.streetNumber : ''}` : '';
    const area = g ? g.district ?? g.city ?? '' : '';
    return street || area ? [street, area].filter(Boolean).join(' · ') : `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
  };

  // Começa com a posição do aparelho (sem sobrescrever um local que a pessoa já escolheu no mapa)
  useEffect(() => {
    (async () => {
      try {
        const { status } = await Location.requestForegroundPermissionsAsync();
        if (status !== 'granted') {
          setLoc(l => (l.source === 'none' ? { ...l, addr: 'Localização desligada · escolha o local no mapa' } : l));
          return;
        }
        const p = await Location.getCurrentPositionAsync({ accuracy: Location.Accuracy.High });
        const addr = await addressOf(p.coords.latitude, p.coords.longitude);
        app.setHere({ lat: p.coords.latitude, lng: p.coords.longitude });
        setLoc(l => (l.source === 'map' ? l : { lat: p.coords.latitude, lng: p.coords.longitude, acc: Math.round(p.coords.accuracy ?? 10), addr, source: 'gps' }));
      } catch {
        setLoc(l => (l.source === 'none' ? { ...l, addr: 'Não foi possível obter sua localização · escolha o local no mapa' } : l));
      }
    })();
  }, []);

  // Distância (km) entre a posição real do usuário e o ponto do pino; null = sem GPS
  const pickKm = app.here ? distanceM(app.here, pick.lat, pick.lng) / 1000 : null;
  const pickTooFar = pickKm !== null && pickKm > MAX_DISTANCIA_AJUSTE_KM;

  const openPicker = () => {
    if (!PERMITIR_AJUSTE_DO_LOCAL) return;
    // Sem a posição real não dá para conferir o limite de distância
    if (!app.here) return app.flash('Ative a localização: o local precisa ficar a até ' + MAX_DISTANCIA_AJUSTE_KM + ' km de você');
    setPick({ lat: loc.lat, lng: loc.lng });
    setPicking(true);
  };
  const confirmPick = async () => {
    if (pickTooFar) return app.flash(`Escolha um local a até ${MAX_DISTANCIA_AJUSTE_KM} km da sua posição`);
    const chosen = { ...pick };
    setPicking(false);
    setLoc({ lat: chosen.lat, lng: chosen.lng, acc: 0, addr: 'Buscando endereço…', source: 'map' });
    const addr = await addressOf(chosen.lat, chosen.lng);
    setLoc(l => (l.source === 'map' && l.lat === chosen.lat && l.lng === chosen.lng ? { ...l, addr } : l));
  };
  const pickerToMe = () => {
    const h = app.here;
    if (!h) return app.flash('Sua localização ainda não está disponível');
    pickMapRef.current?.animateToRegion({ latitude: h.lat, longitude: h.lng, latitudeDelta: 0.004, longitudeDelta: 0.004 }, 400);
  };

  const submit = () => {
    if (!app.authed) return nav.navigate('Login');
    if (loc.source === 'none') {
      if (!PERMITIR_AJUSTE_DO_LOCAL || !app.here) return app.flash('Ative a localização para enviar o relato');
      app.flash('Escolha o local do relato no mapa');
      return openPicker();
    }
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
            <Pressable onPress={openPicker} disabled={!PERMITIR_AJUSTE_DO_LOCAL} style={{ backgroundColor: c.surface, borderWidth: 1.5, borderColor: c.border, borderRadius: r.input, padding: 12, paddingHorizontal: 14, flexDirection: 'row', gap: 12, alignItems: 'center' }}>
              <View style={{ width: 40, height: 40, borderRadius: 12, backgroundColor: c.syncBg, alignItems: 'center', justifyContent: 'center' }}><Icon name={loc.source === 'map' ? 'place' : 'my-location'} color={c.location} /></View>
              <View style={{ flex: 1, gap: 3 }}>
                <Text style={t.bodyStrong} numberOfLines={1}>{loc.addr}</Text>
                {loc.source !== 'none' && (
                  <Text style={t.mono}>{loc.lat.toFixed(5)}, {loc.lng.toFixed(5)} · {loc.source === 'map' ? 'escolhido no mapa' : `±${loc.acc} m · GPS`}</Text>
                )}
              </View>
              {PERMITIR_AJUSTE_DO_LOCAL && <Text style={{ fontFamily: f.bold, fontSize: 13, color: c.primaryStrong }}>Ajustar</Text>}
            </Pressable>
            {PERMITIR_AJUSTE_DO_LOCAL && (
              <Text style={[t.caption, { paddingHorizontal: 4 }]}>Você pode ajustar o local para até {MAX_DISTANCIA_AJUSTE_KM} km da sua posição.</Text>
            )}
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

      {/* Escolher o local no mapa: arraste o mapa até o pino apontar para o lugar */}
      <Modal visible={picking} animationType="slide" onRequestClose={() => setPicking(false)}>
        <View style={{ flex: 1, backgroundColor: c.mapBase }}>
          <MapView
            ref={pickMapRef}
            style={StyleSheet.absoluteFill}
            provider={PROVIDER_GOOGLE}
            customMapStyle={mapStyle}
            initialRegion={{ latitude: pick.lat, longitude: pick.lng, latitudeDelta: 0.004, longitudeDelta: 0.004 }}
            showsUserLocation
            showsMyLocationButton={false}
            toolbarEnabled={false}
            onRegionChangeComplete={(reg: Region) => setPick({ lat: reg.latitude, lng: reg.longitude })}
          >
            {/* Limite de distância: o local precisa ficar dentro deste círculo */}
            {app.here && (
              <Circle center={{ latitude: app.here.lat, longitude: app.here.lng }} radius={MAX_DISTANCIA_AJUSTE_KM * 1000}
                strokeColor={c.primary} strokeWidth={2} fillColor={c.primary + '10'} />
            )}
          </MapView>
          {/* Pino fixo no centro: a ponta do ícone fica exatamente no centro do mapa */}
          <View pointerEvents="none" style={{ position: 'absolute', left: 0, right: 0, top: 0, bottom: 0, alignItems: 'center', justifyContent: 'center' }}>
            <View style={{ marginBottom: 44 }}>
              <Icon name="place" size={48} color={c.danger} />
            </View>
          </View>
          <SafeAreaView edges={['top']} style={{ position: 'absolute', left: 12, right: 12, top: 0, paddingTop: 10, flexDirection: 'row', gap: 10, alignItems: 'center' }} pointerEvents="box-none">
            <IconButton icon="close" onPress={() => setPicking(false)} />
            <View style={{ flex: 1, backgroundColor: c.surface, borderRadius: 22, paddingHorizontal: 14, paddingVertical: 10 }}>
              <Text style={t.bodyStrong}>Escolha o local do relato</Text>
              <Text style={t.caption}>Arraste o mapa até o pino apontar para o lugar (até {MAX_DISTANCIA_AJUSTE_KM} km de você)</Text>
            </View>
          </SafeAreaView>
          <SafeAreaView edges={['bottom']} style={{ position: 'absolute', left: 12, right: 12, bottom: 12, gap: 10, alignItems: 'flex-end' }} pointerEvents="box-none">
            <IconButton icon="my-location" size={50} onPress={pickerToMe} />
            <View style={{ alignSelf: 'stretch', backgroundColor: c.surface, borderRadius: 22, padding: 14, gap: 10 }}>
              {pickTooFar ? (
                <View style={{ flexDirection: 'row', gap: 8, alignItems: 'center', backgroundColor: c.dangerBg, borderRadius: 14, padding: 10 }}>
                  <Icon name="warning" size={20} color={c.dangerText} />
                  <Text style={{ flex: 1, fontFamily: f.semibold, fontSize: 13, color: c.dangerText }}>
                    Esse ponto está a {pickKm!.toFixed(1).replace('.', ',')} km de você. O limite é {MAX_DISTANCIA_AJUSTE_KM} km.
                  </Text>
                </View>
              ) : (
                <Text style={t.mono}>
                  {pick.lat.toFixed(5)}, {pick.lng.toFixed(5)}{pickKm !== null ? ` · ${pickKm < 1 ? Math.round(pickKm * 1000) + ' m' : pickKm.toFixed(1).replace('.', ',') + ' km'} de você` : ''}
                </Text>
              )}
              <Button label="Confirmar local" icon="check" disabled={pickTooFar} onPress={confirmPick} />
            </View>
          </SafeAreaView>
        </View>
      </Modal>
    </SafeAreaView>
  );
}
