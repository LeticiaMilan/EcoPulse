import React from 'react';
import { Text, View } from 'react-native';
import { c, f, hazard, aqi, aqiLevel, shadow } from '../theme';
import type { Report } from '../data/mock';
import { Icon } from './ui';

export function ReportPin({ report, selected }: { report: Report; selected: boolean }) {
  const d = selected ? 42 : 34;
  const color = report.status === 'fin' ? c.ink3 : hazard[report.type].solid;
  return (
    <View style={{ padding: 6 }}>
      <View style={[{ width: d, height: d, borderRadius: d / 2, backgroundColor: color, borderWidth: 3, borderColor: '#fff', alignItems: 'center', justifyContent: 'center' }, shadow.floating, selected && { transform: [{ scale: 1.05 }] }]}>
        <Icon name={hazard[report.type].icon} size={selected ? 22 : 18} color="#fff" />
      </View>
    </View>
  );
}

export function AqiBadge({ v }: { v: number }) {
  const a = aqi[aqiLevel(v)];
  return (
    <View style={{ width: 46, height: 46, borderRadius: 14, backgroundColor: a.bg, borderWidth: 2, borderColor: a.border, alignItems: 'center', justifyContent: 'center' }}>
      <Text style={{ fontFamily: f.extrabold, fontSize: 13, color: a.fg, lineHeight: 14 }}>{v}</Text>
      <Text style={{ fontFamily: f.bold, fontSize: 8, color: a.fg, letterSpacing: 0.4 }}>AQI</Text>
    </View>
  );
}

export const mapStyle = [
  { elementType: 'geometry', stylers: [{ color: c.mapBase }] },
  { elementType: 'labels.text.fill', stylers: [{ color: c.ink2 }] },
  { elementType: 'labels.text.stroke', stylers: [{ color: '#ffffff' }] },
  { featureType: 'poi', elementType: 'labels.icon', stylers: [{ visibility: 'off' }] },
  { featureType: 'poi.park', elementType: 'geometry', stylers: [{ color: c.mapPark }] },
  { featureType: 'road', elementType: 'geometry', stylers: [{ color: '#ffffff' }] },
  { featureType: 'road', elementType: 'geometry.stroke', stylers: [{ color: c.border }] },
  { featureType: 'transit', stylers: [{ visibility: 'off' }] },
  { featureType: 'water', elementType: 'geometry', stylers: [{ color: c.mapWater }] },
];
