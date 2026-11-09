import React from 'react';
import { ScrollView, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { c, f, t, shadow } from '../theme';
import { useApp } from '../store/AppContext';
import { Button, Card, HazardIcon, Icon, ScreenHeader } from '../components/ui';

function Version({ title, sev, note, meta }: { title: string; sev: string; note: string; meta: string }) {
  return (
    <View style={{ flex: 1, backgroundColor: c.subtle, borderRadius: 14, padding: 12, gap: 6 }}>
      <Text style={[t.overline, { textTransform: 'uppercase', letterSpacing: 0.5 }]}>{title}</Text>
      <Text style={[t.bodySm, { fontFamily: f.bold }]}>{sev}</Text>
      <Text style={[t.caption, { fontFamily: f.medium, color: c.ink2 }]}>{note}</Text>
      <Text style={[t.mono, { fontSize: 10 }]}>{meta}</Text>
    </View>
  );
}

export default function SyncScreen() {
  const nav = useNavigation();
  const app = useApp();
  const pending = app.reports.filter(r => r.sync === 'pending');
  const total = app.pendingCount;
  const [bg, fg] = total ? [c.offBg, c.offFg] : [c.container, c.onContainer];

  const syncNow = () => {
    if (app.offline || app.syncing) return;
    app.runSync(); // envia o que estiver na fila e atualiza a lista do servidor
  };

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: c.bg }}>
      <ScreenHeader title="Sincronização" onBack={() => nav.goBack()} />
      <ScrollView contentContainerStyle={{ paddingHorizontal: 16, paddingBottom: 20, paddingTop: 4, gap: 14 }}>
        <Card>
          <View style={{ flexDirection: 'row', gap: 12, alignItems: 'center' }}>
            <View style={{ width: 48, height: 48, borderRadius: 16, backgroundColor: bg, alignItems: 'center', justifyContent: 'center' }}>
              <Icon name={app.syncing ? 'sync' : total ? 'cloud-upload' : 'cloud-done'} size={26} color={fg} />
            </View>
            <View style={{ gap: 2 }}>
              <Text style={t.headline}>{app.syncing ? 'Sincronizando…' : total ? `${total} ${total === 1 ? 'item aguardando' : 'itens aguardando'}` : 'Tudo sincronizado'}</Text>
              <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13, color: c.ink2 }]}>Última sincronização: {app.lastSync}</Text>
            </View>
          </View>
          <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13, lineHeight: 19, color: c.ink2 }]}>Acontece em segundo plano. Você pode continuar usando o app enquanto isso.</Text>
          <Button size="md" variant="dark" disabled={app.offline} label={app.offline ? 'Aguardando conexão…' : app.syncing ? 'Sincronizando…' : 'Sincronizar agora'} onPress={syncNow} />
        </Card>

        {app.conflict === 'open' && (
          <Card style={{ borderWidth: 1.5, borderColor: c.offBg }}>
            <View style={{ flexDirection: 'row', gap: 10, alignItems: 'center' }}>
              <Icon name="merge" color={c.offFg} />
              <Text style={[t.bodyStrong, { fontFamily: f.extrabold, flex: 1 }]}>Versões diferentes do mesmo relato</Text>
            </View>
            <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13, lineHeight: 19, color: c.ink2 }]}>“Terra deslizando na calçada” foi editado no seu aparelho (offline) e no servidor ao mesmo tempo.</Text>
            <View style={{ flexDirection: 'row', gap: 8 }}>
              <Version title="No aparelho" sev="Gravidade alta" note="“Terra cobrindo metade da rua”" meta="09:02 · v3" />
              <Version title="No servidor" sev="Gravidade média" note="“Terra só na calçada”" meta="09:05 · v3′" />
            </View>
            <View style={{ flexDirection: 'row', gap: 8 }}>
              <Button style={{ flex: 1 }} size="md" label="Minha versão" onPress={() => app.resolveConflict('mine')} />
              <Button style={{ flex: 1 }} size="md" variant="secondary" label="Do servidor" onPress={() => app.resolveConflict('server')} />
            </View>
            <Text style={[t.caption, { fontFamily: f.medium, lineHeight: 17 }]}>Votos nunca conflitam: são somados. Nos outros campos, a versão mais recente é sugerida.</Text>
          </Card>
        )}

        <View style={{ gap: 8 }}>
          <Text style={[t.label, { paddingHorizontal: 4 }]}>Fila do aparelho</Text>
          {pending.map((r, i) => (
            <View key={r.id} style={[{ flexDirection: 'row', gap: 12, alignItems: 'center', padding: 12, paddingHorizontal: 14, backgroundColor: c.surface, borderRadius: 18 }, shadow.card]}>
              <HazardIcon type={r.type} size={38} />
              <View style={{ flex: 1, gap: 2 }}>
                <Text style={[t.bodySm, { fontFamily: f.bold }]} numberOfLines={1}>{r.title}</Text>
                <Text style={[t.mono, { fontSize: 10.5 }]}>{app.offline ? `${r.id} · aguardando rede` : app.syncing ? `${r.id} · enviando…` : `${r.id} · tentativa ${i + 1} de 5`}</Text>
              </View>
            </View>
          ))}
          {!pending.length && <Text style={[t.bodySm, { color: c.ink2, textAlign: 'center', padding: 16 }]}>Nada esperando envio.</Text>}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
