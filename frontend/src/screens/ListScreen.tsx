import React, { useState } from 'react';
import { FlatList, Pressable, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { c, f, t, hazard, shadow } from '../theme';
import { useApp, voteCaption } from '../store/AppContext';
import { HazardIcon, Icon, IconButton, StatusChip } from '../components/ui';
import type { RootNav } from '../navigation/types';

export default function ListScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const [tab, setTab] = useState<'near' | 'mine'>('near');
  const items = tab === 'mine' ? app.reports.filter(r => r.mine) : app.visibleReports;
  const empty = tab === 'mine' ? (app.authed ? 'Você ainda não criou relatos.' : 'Entre na sua conta para ver seus relatos.') : 'Nenhum relato com esses filtros.';

  return (
    <SafeAreaView edges={['top']} style={{ flex: 1, backgroundColor: c.bg }}>
      <View style={{ paddingHorizontal: 20, paddingTop: 12, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' }}>
        <Text style={t.titleLg}>Relatos</Text>
        <IconButton icon="tune" variant="sunken" onPress={() => nav.navigate('Filters')} />
      </View>
      <View style={{ flexDirection: 'row', gap: 6, paddingHorizontal: 20, paddingTop: 14, paddingBottom: 8 }}>
        {([['near', 'Perto de mim'], ['mine', 'Meus relatos']] as const).map(([k, l]) => {
          const on = tab === k;
          return (
            <Pressable key={k} onPress={() => setTab(k)} style={{ height: 38, paddingHorizontal: 16, borderRadius: 19, justifyContent: 'center', backgroundColor: on ? c.container : 'transparent', borderWidth: on ? 0 : 1.5, borderColor: c.border }}>
              <Text style={{ fontFamily: f.bold, fontSize: 14, color: on ? c.onContainer : c.ink2 }}>{l}</Text>
            </Pressable>
          );
        })}
      </View>
      <FlatList
        data={items}
        keyExtractor={r => r.id}
        contentContainerStyle={{ padding: 16, paddingTop: 8, gap: 10 }}
        ListEmptyComponent={<Text style={[t.bodySm, { color: c.ink2, textAlign: 'center', padding: 40 }]}>{empty}</Text>}
        renderItem={({ item: r }) => (
          <Pressable onPress={() => { app.setSelectedId(r.id); nav.navigate('Detail'); }} style={({ pressed }) => [{ backgroundColor: c.surface, borderRadius: 20, padding: 14, flexDirection: 'row', gap: 12, transform: [{ scale: pressed ? 0.99 : 1 }] }, shadow.card]}>
            <HazardIcon type={r.type} />
            <View style={{ flex: 1, gap: 5 }}>
              <View style={{ flexDirection: 'row', justifyContent: 'space-between', gap: 8 }}>
                <Text style={[t.overline, { fontSize: 12, flexShrink: 1 }]} numberOfLines={1}>{hazard[r.type].name}{r.dist ? ` · ${r.dist}` : ''}</Text>
                <Text style={[t.caption, { fontFamily: f.medium }]}>{r.ago}</Text>
              </View>
              <Text style={[t.bodyStrong, { fontSize: 15, lineHeight: 20 }]}>{r.title}</Text>
              <View style={{ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
                <StatusChip status={r.status} />
                <Text style={[t.caption, { fontFamily: f.medium }]}>{voteCaption(r)}</Text>
                {r.sync === 'pending' && (
                  <View style={{ flexDirection: 'row', alignItems: 'center', gap: 3 }}>
                    <Icon name="cloud-upload" size={15} color={c.offFg} />
                    <Text style={{ fontFamily: f.bold, fontSize: 11, color: c.offFg }}>Não enviado</Text>
                  </View>
                )}
              </View>
            </View>
          </Pressable>
        )}
      />
    </SafeAreaView>
  );
}
