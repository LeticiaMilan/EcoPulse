import React from 'react';
import { ScrollView, Text, View } from 'react-native';
import MapView, { Circle, Marker, PROVIDER_GOOGLE } from 'react-native-maps';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { c, f, t, hazard } from '../theme';
import { useApp } from '../store/AppContext';
import { NEED_VOTES } from '../data/mock';
import { Button, Card, HazardPill, Icon, IconButton, StatusChip } from '../components/ui';
import { mapStyle, ReportPin } from '../components/MapPins';
import type { RootNav } from '../navigation/types';

function Stepper({ done }: { done: number }) {
  return (
    <View style={{ flexDirection: 'row' }}>
      {['Enviado', 'Validado', 'Finalizado'].map((l, i) => {
        const ok = done > i;
        return (
          <View key={l} style={{ flex: 1, alignItems: 'center', gap: 6 }}>
            {i > 0 && <View style={{ position: 'absolute', top: 13, right: '50%', width: '100%', height: 3, backgroundColor: ok ? c.primary : c.border }} />}
            <View style={{ width: 28, height: 28, borderRadius: 14, alignItems: 'center', justifyContent: 'center', backgroundColor: ok ? c.primary : c.surface, borderWidth: ok ? 0 : 2, borderColor: c.border }}>
              {ok ? <Icon name="check" size={16} color="#fff" /> : <Text style={{ fontFamily: f.extrabold, fontSize: 13, color: c.ink3 }}>{i + 1}</Text>}
            </View>
            <Text style={{ fontFamily: f.bold, fontSize: 12, color: ok ? c.ink : c.ink3 }}>{l}</Text>
          </View>
        );
      })}
    </View>
  );
}

export default function DetailScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const rep = app.selected;
  const stage = rep.status === 'pend' ? 'valid' : rep.status === 'val' ? 'final' : 'done';
  const my = app.myVotes[rep.id];
  const canVote = app.authed && !my && stage !== 'done' && !rep.mine;
  const showVoted = !!my || (rep.mine && stage !== 'done' && app.authed);

  return (
    <View style={{ flex: 1, backgroundColor: c.bg }}>
      <ScrollView>
        <View style={{ height: 190 }}>
          <MapView style={{ flex: 1 }} provider={PROVIDER_GOOGLE} customMapStyle={mapStyle} liteMode scrollEnabled={false} zoomEnabled={false} pitchEnabled={false} rotateEnabled={false}
            region={{ latitude: rep.lat, longitude: rep.lng, latitudeDelta: 0.008, longitudeDelta: 0.008 }}>
            <Circle center={{ latitude: rep.lat, longitude: rep.lng }} radius={rep.radius * 1.4} fillColor={hazard[rep.type].solid + '38'} strokeColor={hazard[rep.type].solid + '80'} strokeWidth={1.5} />
            <Marker coordinate={{ latitude: rep.lat, longitude: rep.lng }} anchor={{ x: 0.5, y: 0.5 }} tracksViewChanges={false}><ReportPin report={rep} selected /></Marker>
          </MapView>
          <SafeAreaView edges={['top']} style={{ position: 'absolute', left: 12, top: 10 }}>
            <IconButton icon="arrow-back" onPress={() => nav.goBack()} />
          </SafeAreaView>
        </View>

        <View style={{ padding: 20, paddingTop: 18, gap: 18 }}>
          <View style={{ gap: 8 }}>
            <View style={{ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
              <HazardPill type={rep.type} />
              <StatusChip status={rep.status} />
            </View>
            <Text style={t.titleMd}>{rep.title}</Text>
            <Text style={[t.bodySm, { color: c.ink2 }]}>{rep.addr} · {rep.dist}</Text>
            <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13 }]}>Relatado por {rep.by} {rep.ago} · gravidade {rep.severity}</Text>
          </View>

          <Stepper done={rep.status === 'pend' ? 1 : rep.status === 'val' ? 2 : 3} />

          <Card>
            <View style={{ gap: 4 }}>
              <Text style={t.headline}>{stage === 'valid' ? 'Você confirma esse relato?' : stage === 'final' ? 'A situação já foi resolvida?' : 'Relato finalizado'}</Text>
              <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13, lineHeight: 19, color: c.ink2 }]}>
                {stage === 'valid' ? `Com ${NEED_VOTES} confirmações ele é validado e todos por perto recebem o alerta.` : stage === 'final' ? `Com ${NEED_VOTES} votos de "já resolvido" o relato sai do mapa.` : 'A comunidade confirmou que a situação foi resolvida. Fica no histórico para consulta.'}
              </Text>
            </View>
            {stage !== 'done' && (
              <View style={{ gap: 6 }}>
                <View style={{ height: 10, borderRadius: 5, backgroundColor: c.sunken, overflow: 'hidden' }}>
                  <View style={{ height: '100%', width: `${(rep.votes / NEED_VOTES) * 100}%`, borderRadius: 5, backgroundColor: stage === 'valid' ? c.progressPend : c.primary }} />
                </View>
                <View style={{ flexDirection: 'row', justifyContent: 'space-between' }}>
                  <Text style={[t.overline, { fontSize: 12 }]}>{rep.votes}/{NEED_VOTES} votos</Text>
                  <Text style={[t.overline, { fontSize: 12 }]}>faltam {NEED_VOTES - rep.votes}</Text>
                </View>
              </View>
            )}
            {canVote && (
              <View style={{ flexDirection: 'row', gap: 10 }}>
                <Button style={{ flex: 1, borderRadius: 16 }} size="md" icon="thumb-up" label={stage === 'valid' ? 'Confirmo' : 'Já resolvido'} onPress={() => app.vote(rep.id, true)} />
                <Button style={{ flex: 1, borderRadius: 16 }} size="md" variant="secondary" label={stage === 'valid' ? 'Não vi isso' : 'Ainda ativo'} onPress={() => app.vote(rep.id, false)} />
              </View>
            )}
            {showVoted && (
              <View style={{ flexDirection: 'row', gap: 8, alignItems: 'center', backgroundColor: c.subtleBrand, borderRadius: 14, padding: 12, paddingHorizontal: 14 }}>
                <Icon name="how-to-vote" size={20} color={c.valFg} />
                <Text style={{ flex: 1, fontFamily: f.semibold, fontSize: 14, color: c.valFg }}>
                  {rep.mine && !my ? 'Relato seu — quem está por perto é quem vota.' : my && rep.sync === 'pending' ? 'Voto salvo no aparelho. Envia quando a rede voltar.' : 'Seu voto foi registrado.'}
                </Text>
              </View>
            )}
            {!app.authed && stage !== 'done' && <Button variant="outline" size="md" label="Entrar para votar" onPress={() => nav.navigate('Login')} />}
          </Card>

          <View style={{ flexDirection: 'row', gap: 8, alignItems: 'center' }}>
            <Icon name={rep.sync === 'pending' ? 'cloud-upload' : 'cloud-done'} size={16} color={c.ink3} />
            <Text style={t.mono}>{rep.sync === 'pending' ? `${rep.id} · na fila local (outbox) · não enviado` : `${rep.id} · sincronizado · ${app.lastSync}`}</Text>
          </View>
        </View>
      </ScrollView>
    </View>
  );
}
