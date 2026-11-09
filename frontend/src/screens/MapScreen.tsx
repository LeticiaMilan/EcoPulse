import React, { useEffect, useMemo, useRef } from 'react';
import { FlatList, NativeScrollEvent, NativeSyntheticEvent, Pressable, ScrollView, StyleSheet, Text, useWindowDimensions, View } from 'react-native';
import MapView, { Circle, Marker, Polygon, PROVIDER_GOOGLE } from 'react-native-maps';
import * as Location from 'expo-location';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useIsFocused, useNavigation } from '@react-navigation/native';
import { c, f, t, hazard, shadow } from '../theme';
import { useApp, voteCaption } from '../store/AppContext';
import { AQI_POINTS, HEAT_SPOTS, Report, USER } from '../data/mock';
import { ConnectionPill, HazardIcon, Icon, IconButton, LayerChip, StatusChip } from '../components/ui';
import { AqiBadge, mapStyle, ReportPin } from '../components/MapPins';
import type { RootNav } from '../navigation/types';

export default function MapScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const sel = app.selected;
  const mapRef = useRef<MapView>(null);

  // Vai até a posição atual do aparelho (e guarda para calcular distâncias)
  const goToMe = async (animate = true) => {
    try {
      const { status } = await Location.requestForegroundPermissionsAsync();
      if (status !== 'granted') { if (animate) app.flash('Permita o acesso à localização para ver onde você está'); return; }
      const p = (await Location.getLastKnownPositionAsync()) ?? (await Location.getCurrentPositionAsync({ accuracy: Location.Accuracy.Balanced }));
      const here = { lat: p.coords.latitude, lng: p.coords.longitude };
      app.setHere(here);
      mapRef.current?.animateToRegion({ latitude: here.lat, longitude: here.lng, latitudeDelta: 0.012, longitudeDelta: 0.012 }, animate ? 600 : 0);
      if (animate) {
        // A última posição conhecida pode estar velha: atualiza com a atual
        const fresh = await Location.getCurrentPositionAsync({ accuracy: Location.Accuracy.Balanced });
        app.setHere({ lat: fresh.coords.latitude, lng: fresh.coords.longitude });
        mapRef.current?.animateToRegion({ latitude: fresh.coords.latitude, longitude: fresh.coords.longitude, latitudeDelta: 0.012, longitudeDelta: 0.012 }, 400);
      }
    } catch {
      if (animate) app.flash('Não foi possível obter sua localização');
    }
  };

  // Ao abrir, centraliza no usuário
  useEffect(() => { goToMe(false); }, []);

  // Áreas afetadas ativas (uma por área, mesmo com vários relatos dentro)
  const areas = useMemo(() => {
    const seen = new Map<string, { id: string; color: string; coords: { latitude: number; longitude: number }[] }>();
    for (const rep of app.visibleReports) {
      if (rep.area?.status === 'active' && rep.area.polygon.length > 2 && !seen.has(rep.area.id)) {
        seen.set(rep.area.id, { id: rep.area.id, color: hazard[rep.type].solid, coords: rep.area.polygon });
      }
    }
    return [...seen.values()];
  }, [app.visibleReports]);

  // O mapa nativo às vezes não remove polígonos/marcadores quando a lista muda.
  // Esta chave muda quando entra/sai relato ou muda status, e força redesenhar as camadas.
  const overlayKey = useMemo(() => {
    let h = 0;
    for (const r of app.visibleReports) {
      const k = r.id + r.status + (r.area?.status ?? '');
      for (let i = 0; i < k.length; i++) h = (h * 31 + k.charCodeAt(i)) | 0;
    }
    return `${h}:${areas.length}`;
  }, [app.visibleReports, areas]);

  // ---------- carrossel de relatos ----------
  const { width } = useWindowDimensions();
  const CARD_W = width - 24 - 28; // deixa aparecer um pedaço do próximo cartão
  const STEP = CARD_W + 10;
  const listRef = useRef<FlatList<Report>>(null);
  const items = app.visibleReports;

  const centerOn = (rep: Report) =>
    mapRef.current?.animateToRegion({ latitude: rep.lat, longitude: rep.lng, latitudeDelta: 0.01, longitudeDelta: 0.01 }, 500);
  const scrollTo = (index: number) => {
    if (index >= 0 && index < items.length) listRef.current?.scrollToOffset({ offset: index * STEP, animated: true });
  };
  const openDetail = (rep: Report) => { app.setSelectedId(rep.id); nav.navigate('Detail'); };

  // Tocar no pino do mapa leva o carrossel até o relato
  const onMarker = (rep: Report) => { app.setSelectedId(rep.id); scrollTo(items.findIndex(r => r.id === rep.id)); };

  // Ao parar de arrastar, seleciona o cartão visível e centraliza o mapa nele
  const onSwipeEnd = (e: NativeSyntheticEvent<NativeScrollEvent>) => {
    const i = Math.min(Math.max(Math.round(e.nativeEvent.contentOffset.x / STEP), 0), items.length - 1);
    const rep = items[i];
    if (rep && rep.id !== app.selectedId) { app.setSelectedId(rep.id); centerOn(rep); }
  };

  // "Ver no mapa" vindo de outra tela.
  // O pedido chega enquanto o mapa ainda está escondido (atrás do detalhe ou em outra aba),
  // e o mapa nativo ignora animações nesse estado. Por isso só centraliza quando a tela do
  // mapa volta a ficar visível, com um pequeno atraso para o mapa terminar de aparecer.
  const isFocused = useIsFocused();
  const handledFocus = useRef(0);
  useEffect(() => {
    if (!isFocused || !app.focusTick || handledFocus.current === app.focusTick) return;
    handledFocus.current = app.focusTick;
    const target = app.selected;
    if (!target) return;
    const timer = setTimeout(() => {
      centerOn(target);
      scrollTo(items.findIndex(r => r.id === target.id));
    }, 350);
    return () => clearTimeout(timer);
  }, [isFocused, app.focusTick]);

  return (
    <View style={{ flex: 1, backgroundColor: c.mapBase }}>
      <MapView
        ref={mapRef}
        style={StyleSheet.absoluteFill}
        provider={PROVIDER_GOOGLE}
        customMapStyle={mapStyle}
        initialRegion={{ latitude: USER.lat, longitude: USER.lng, latitudeDelta: 0.028, longitudeDelta: 0.028 }}
        showsUserLocation
        showsMyLocationButton={false}
        toolbarEnabled={false}
        mapPadding={{ top: 150, right: 0, bottom: 150, left: 0 }}
      >
        {app.layers.heat && HEAT_SPOTS.map((h, i) => (
          <Circle key={'h' + i} center={{ latitude: h.lat, longitude: h.lng }} radius={h.radius} fillColor={h.color + '40'} strokeWidth={0} />
        ))}
        <React.Fragment key={overlayKey}>
        {app.layers.reports && areas.map(a => (
          <Polygon key={'area' + a.id} coordinates={a.coords} fillColor={a.color + '30'} strokeColor={a.color} strokeWidth={2} />
        ))}
        {app.layers.reports && app.visibleReports.map(rep => (
          <Circle key={'z' + rep.id} center={{ latitude: rep.lat, longitude: rep.lng }} radius={rep.radius}
            fillColor={hazard[rep.type].solid + (rep.status === 'fin' ? '14' : '33')} strokeColor={hazard[rep.type].solid + '80'} strokeWidth={1.5}
            lineDashPattern={rep.status === 'pend' ? [8, 6] : undefined} />
        ))}
        {app.layers.reports && app.visibleReports.map(rep => {
          const on = rep.id === app.selectedId;
          return (
            <Marker key={rep.id + (on ? '-on' : '')} coordinate={{ latitude: rep.lat, longitude: rep.lng }} anchor={{ x: 0.5, y: 0.5 }} tracksViewChanges={false} onPress={() => onMarker(rep)}>
              <ReportPin report={rep} selected={on} />
            </Marker>
          );
        })}
        </React.Fragment>
        {app.layers.aqi && AQI_POINTS.map((p, i) => (
          <Marker key={'a' + i} coordinate={{ latitude: p.lat, longitude: p.lng }} anchor={{ x: 0.5, y: 0.5 }} tracksViewChanges={false}>
            <AqiBadge v={p.v} />
          </Marker>
        ))}
      </MapView>

      <SafeAreaView edges={['top']} style={{ position: 'absolute', left: 12, right: 12, top: 0, gap: 10, paddingTop: 10 }} pointerEvents="box-none">
        <View style={{ flexDirection: 'row', gap: 8 }}>
          <View style={[{ flex: 1, height: 50, borderRadius: 25, backgroundColor: c.surface, flexDirection: 'row', alignItems: 'center', gap: 10, paddingHorizontal: 16 }, shadow.floating]}>
            <Icon name="search" color={c.ink2} />
            <Text style={[t.body, { color: c.ink3 }]}>Buscar rua ou bairro</Text>
          </View>
          <IconButton icon="tune" size={50} onPress={() => nav.navigate('Filters')}>
            {app.filtersActive && <View style={{ position: 'absolute', top: 10, right: 10, width: 9, height: 9, borderRadius: 5, backgroundColor: c.primary, borderWidth: 2, borderColor: '#fff' }} />}
          </IconButton>
        </View>
        <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={{ gap: 8, paddingVertical: 4, paddingHorizontal: 2 }}>
          <LayerChip label="Relatos" icon="place" active={app.layers.reports} onPress={() => app.toggleLayer('reports')} />
          <LayerChip label="Qualidade do ar" icon="air" active={app.layers.aqi} onPress={() => app.toggleLayer('aqi')} />
          <LayerChip label="Ilhas de calor" icon="device-thermostat" active={app.layers.heat} onPress={() => app.toggleLayer('heat')} />
        </ScrollView>
        <ConnectionPill />
      </SafeAreaView>

      <View style={{ position: 'absolute', left: 12, right: 12, bottom: 12, alignItems: 'flex-end', gap: 12 }} pointerEvents="box-none">
        <IconButton icon="my-location" size={50} onPress={() => goToMe(true)} />
        <Pressable onPress={() => nav.navigate('NewReport')} style={({ pressed }) => [{ height: 56, paddingLeft: 18, paddingRight: 22, borderRadius: 18, backgroundColor: c.primary, flexDirection: 'row', alignItems: 'center', gap: 8, transform: [{ scale: pressed ? 0.97 : 1 }] }, shadow.raised]}>
          <Icon name="add-location-alt" size={24} color="#fff" />
          <Text style={{ fontFamily: f.bold, fontSize: 16, color: '#fff' }}>Relatar</Text>
        </Pressable>
        {items.length > 0 && (
          <FlatList
            ref={listRef}
            style={{ alignSelf: 'stretch', marginHorizontal: -12, marginVertical: -8 }}
            contentContainerStyle={{ paddingHorizontal: 12, paddingVertical: 8, gap: 10 }}
            data={items}
            extraData={app.selectedId}
            keyExtractor={r => r.id}
            horizontal
            showsHorizontalScrollIndicator={false}
            snapToInterval={STEP}
            decelerationRate="fast"
            onMomentumScrollEnd={onSwipeEnd}
            renderItem={({ item: r, index }) => {
              const on = r.id === sel?.id;
              return (
                <Pressable onPress={() => openDetail(r)} style={[{ width: CARD_W, backgroundColor: c.surface, borderRadius: 22, padding: 14, flexDirection: 'row', gap: 12, alignItems: 'center', borderWidth: 2, borderColor: on ? c.primary : 'transparent' }, shadow.raised]}>
                  <HazardIcon type={r.type} size={48} solid />
                  <View style={{ flex: 1, gap: 4 }}>
                    <Text style={[t.overline, { fontSize: 12, textTransform: 'uppercase', letterSpacing: 0.6 }]} numberOfLines={1}>{hazard[r.type].name}{r.dist ? ` · ${r.dist}` : ''}</Text>
                    <Text style={t.bodyStrong} numberOfLines={1}>{r.title}</Text>
                    <View style={{ flexDirection: 'row', gap: 6, alignItems: 'center' }}>
                      <StatusChip status={r.status} />
                      <Text style={t.caption} numberOfLines={1}>{voteCaption(r)}</Text>
                    </View>
                  </View>
                  <View style={{ alignItems: 'flex-end', gap: 8 }}>
                    <Text style={[t.caption, { fontFamily: f.bold }]}>{index + 1}/{items.length}</Text>
                    <Pressable onPress={() => openDetail(r)} hitSlop={6} style={({ pressed }) => ({ flexDirection: 'row', alignItems: 'center', gap: 2, height: 32, paddingLeft: 12, paddingRight: 6, borderRadius: 16, backgroundColor: c.container, opacity: pressed ? 0.8 : 1 })}>
                      <Text style={{ fontFamily: f.bold, fontSize: 13, color: c.onContainer }}>Ver relato</Text>
                      <Icon name="chevron-right" size={18} color={c.onContainer} />
                    </Pressable>
                  </View>
                </Pressable>
              );
            }}
          />
        )}
      </View>
    </View>
  );
}
