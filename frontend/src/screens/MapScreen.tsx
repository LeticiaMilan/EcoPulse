import React, { useEffect } from 'react';
import { Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import MapView, { Circle, Marker, PROVIDER_GOOGLE } from 'react-native-maps';
import * as Location from 'expo-location';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { c, f, t, hazard, shadow } from '../theme';
import { useApp } from '../store/AppContext';
import { AQI_POINTS, HEAT_SPOTS, NEED_VOTES, USER } from '../data/mock';
import { ConnectionPill, HazardIcon, Icon, IconButton, LayerChip, StatusChip } from '../components/ui';
import { AqiBadge, mapStyle, ReportPin } from '../components/MapPins';
import type { RootNav } from '../navigation/types';

export default function MapScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const sel = app.selected;

  useEffect(() => { Location.requestForegroundPermissionsAsync().catch(() => {}); }, []);

  return (
    <View style={{ flex: 1, backgroundColor: c.mapBase }}>
      <MapView
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
        {app.layers.reports && app.visibleReports.map(rep => (
          <Circle key={'z' + rep.id} center={{ latitude: rep.lat, longitude: rep.lng }} radius={rep.radius}
            fillColor={hazard[rep.type].solid + (rep.status === 'fin' ? '14' : '33')} strokeColor={hazard[rep.type].solid + '80'} strokeWidth={1.5}
            lineDashPattern={rep.status === 'pend' ? [8, 6] : undefined} />
        ))}
        {app.layers.aqi && AQI_POINTS.map((p, i) => (
          <Marker key={'a' + i} coordinate={{ latitude: p.lat, longitude: p.lng }} anchor={{ x: 0.5, y: 0.5 }} tracksViewChanges={false}>
            <AqiBadge v={p.v} />
          </Marker>
        ))}
        {app.layers.reports && app.visibleReports.map(rep => {
          const on = rep.id === app.selectedId;
          return (
            <Marker key={rep.id + (on ? '-on' : '')} coordinate={{ latitude: rep.lat, longitude: rep.lng }} anchor={{ x: 0.5, y: 0.5 }} tracksViewChanges={false} onPress={() => app.setSelectedId(rep.id)}>
              <ReportPin report={rep} selected={on} />
            </Marker>
          );
        })}
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
        <Pressable onPress={() => nav.navigate('NewReport')} style={({ pressed }) => [{ height: 56, paddingLeft: 18, paddingRight: 22, borderRadius: 18, backgroundColor: c.primary, flexDirection: 'row', alignItems: 'center', gap: 8, transform: [{ scale: pressed ? 0.97 : 1 }] }, shadow.raised]}>
          <Icon name="add-location-alt" size={24} color="#fff" />
          <Text style={{ fontFamily: f.bold, fontSize: 16, color: '#fff' }}>Relatar</Text>
        </Pressable>
        {sel && (
          <Pressable onPress={() => nav.navigate('Detail')} style={[{ width: '100%', backgroundColor: c.surface, borderRadius: 22, padding: 14, flexDirection: 'row', gap: 12, alignItems: 'center' }, shadow.raised]}>
            <HazardIcon type={sel.type} size={48} solid />
            <View style={{ flex: 1, gap: 4 }}>
              <Text style={[t.overline, { fontSize: 12, textTransform: 'uppercase', letterSpacing: 0.6 }]}>{hazard[sel.type].name} · {sel.dist}</Text>
              <Text style={t.bodyStrong} numberOfLines={1}>{sel.title}</Text>
              <View style={{ flexDirection: 'row', gap: 6, alignItems: 'center' }}>
                <StatusChip status={sel.status} />
                <Text style={t.caption}>{sel.status === 'fin' ? 'encerrado' : `${sel.votes}/${NEED_VOTES} votos`}</Text>
              </View>
            </View>
            <Icon name="chevron-right" color={c.ink3} />
          </Pressable>
        )}
      </View>
    </View>
  );
}
