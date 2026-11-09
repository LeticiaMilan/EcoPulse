import React from 'react';
import { ScrollView, Text, View } from 'react-native';
import MapView, { Circle, Marker, Polygon, PROVIDER_GOOGLE } from 'react-native-maps';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { c, f, t, hazard } from '../theme';
import { useApp } from '../store/AppContext';
import { NEED_VOTES } from '../data/mock';
import { Button, Card, HazardPill, Icon, IconButton, StatusChip } from '../components/ui';
import { mapStyle, ReportPin } from '../components/MapPins';
import type { RootNav } from '../navigation/types';

function Stepper({ done }: { done: number }) {
  const steps = ['Enviado', 'Validado', 'Finalizado'];
  return (
    <View>
      {/* Linhas desenhadas antes (por trás) das bolinhas: ligam o centro de uma etapa ao da próxima */}
      {[1, 2].map(i => (
        <View key={'line' + i} style={{ position: 'absolute', top: 12.5, height: 3, left: `${((i - 0.5) / steps.length) * 100}%`, width: `${100 / steps.length}%`, backgroundColor: done > i ? c.primary : c.border }} />
      ))}
      <View style={{ flexDirection: 'row' }}>
        {steps.map((l, i) => {
          const ok = done > i;
          return (
            <View key={l} style={{ flex: 1, alignItems: 'center', gap: 6 }}>
              <View style={{ width: 28, height: 28, borderRadius: 14, alignItems: 'center', justifyContent: 'center', backgroundColor: ok ? c.primary : c.surface, borderWidth: ok ? 0 : 2, borderColor: c.border }}>
                {ok ? <Icon name="check" size={16} color="#fff" /> : <Text style={{ fontFamily: f.extrabold, fontSize: 13, color: c.ink3 }}>{i + 1}</Text>}
              </View>
              <Text style={{ fontFamily: f.bold, fontSize: 12, color: ok ? c.ink : c.ink3 }}>{l}</Text>
            </View>
          );
        })}
      </View>
    </View>
  );
}

export default function DetailScreen() {
  const nav = useNavigation<RootNav>();
  const app = useApp();
  const rep = app.selected;
  if (!rep) {
    // O relato pode ter saído da lista numa atualização
    return (
      <SafeAreaView style={{ flex: 1, backgroundColor: c.bg, alignItems: 'center', justifyContent: 'center', gap: 12 }}>
        <Text style={t.bodySm}>Relato não encontrado.</Text>
        <Button variant="outline" size="md" label="Voltar" onPress={() => nav.goBack()} />
      </SafeAreaView>
    );
  }
  const areaActive = rep.status === 'val' && rep.area?.status === 'active';
  // valid: aguardando validação · final: validado e em área ativa · waiting: validado, sem área ainda · done: encerrado
  const stage = rep.status === 'pend' ? 'valid' : rep.status === 'val' ? (areaActive ? 'final' : 'waiting') : 'done';
  const my = app.myVoteFor(rep);
  const votable = stage === 'valid' || stage === 'final';
  // Ninguém confirma o próprio relato; já o "foi resolvido?" vale para todos da área
  const canVote = app.authed && !my && votable && !(stage === 'valid' && rep.mine);
  const showVoted = app.authed && votable && (!!my || (stage === 'valid' && rep.mine));
  // Votos de resolução da área ("já resolvido" x "ainda ativo")
  const av = rep.area?.votes ?? { stillHappening: 0, resolved: 0, needed: 3 };
  const openOnMap = () => { app.focusOn(rep.id); nav.navigate('Tabs', { screen: 'Map' }); };

  return (
    <View style={{ flex: 1, backgroundColor: c.bg }}>
      <ScrollView>
        <View style={{ height: 190 }}>
          <MapView style={{ flex: 1 }} provider={PROVIDER_GOOGLE} customMapStyle={mapStyle} liteMode scrollEnabled={false} zoomEnabled={false} pitchEnabled={false} rotateEnabled={false}
            region={{ latitude: rep.lat, longitude: rep.lng, latitudeDelta: 0.008, longitudeDelta: 0.008 }}>
            {areaActive && rep.area && rep.area.polygon.length > 2 && (
              <Polygon coordinates={rep.area.polygon} fillColor={hazard[rep.type].solid + '30'} strokeColor={hazard[rep.type].solid} strokeWidth={2} />
            )}
            <Circle center={{ latitude: rep.lat, longitude: rep.lng }} radius={rep.radius * 1.4} fillColor={hazard[rep.type].solid + '38'} strokeColor={hazard[rep.type].solid + '80'} strokeWidth={1.5} />
            <Marker coordinate={{ latitude: rep.lat, longitude: rep.lng }} anchor={{ x: 0.5, y: 0.5 }} tracksViewChanges={false}><ReportPin report={rep} selected /></Marker>
          </MapView>
          <SafeAreaView edges={['top']} style={{ position: 'absolute', left: 12, top: 10 }}>
            <IconButton icon="arrow-back" onPress={() => nav.goBack()} />
          </SafeAreaView>
          {/* Abre o relato no mapa do próprio app (tocar no mapinha abre o Google Maps) */}
          <SafeAreaView edges={['top']} style={{ position: 'absolute', right: 12, top: 10 }}>
            <IconButton icon="map" onPress={openOnMap} />
          </SafeAreaView>
        </View>

        <View style={{ padding: 20, paddingTop: 18, gap: 18 }}>
          <View style={{ gap: 8 }}>
            <View style={{ flexDirection: 'row', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
              <HazardPill type={rep.type} />
              <StatusChip status={rep.status} />
            </View>
            <Text style={t.titleMd}>{rep.title}</Text>
            <Text style={[t.bodySm, { color: c.ink2 }]}>{rep.addr}{rep.dist ? ` · ${rep.dist}` : ''}</Text>
            <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13 }]}>Relatado por {rep.by} {rep.ago} · gravidade {rep.severity}</Text>
          </View>

          <Stepper done={rep.status === 'pend' ? 1 : rep.status === 'val' ? 2 : 3} />

          <Card>
            <View style={{ gap: 4 }}>
              <Text style={t.headline}>
                {stage === 'valid' ? 'Você confirma esse relato?' : stage === 'final' ? 'A situação já foi resolvida?' : stage === 'waiting' ? 'Relato validado' : rep.rejected ? 'Relato rejeitado' : 'Relato finalizado'}
              </Text>
              <Text style={[t.caption, { fontFamily: f.medium, fontSize: 13, lineHeight: 19, color: c.ink2 }]}>
                {stage === 'valid'
                  ? `Com ${NEED_VOTES} confirmações ele é validado na hora e todos por perto recebem o alerta. Se não chegar a ${NEED_VOTES}, depois de 15 minutos vale a maioria dos votos.`
                  : stage === 'final'
                  ? 'Este relato faz parte de uma área afetada. Com pelo menos 3 votos de "já resolvido" (e mais que os de "ainda ativo") a área é encerrada.'
                  : stage === 'waiting'
                  ? 'A área afetada aparece no mapa quando houver pelo menos 3 relatos validados do mesmo tipo, de pessoas diferentes, a até 500 m um do outro.'
                  : rep.rejected
                  ? 'A comunidade não confirmou este relato. Fica no histórico para consulta.'
                  : 'A comunidade confirmou que a situação foi resolvida. Fica no histórico para consulta.'}
              </Text>
            </View>
            {stage === 'valid' && (
              <View style={{ gap: 6 }}>
                <View style={{ height: 10, borderRadius: 5, backgroundColor: c.sunken, overflow: 'hidden' }}>
                  <View style={{ height: '100%', width: `${Math.min(100, (rep.votes / NEED_VOTES) * 100)}%`, borderRadius: 5, backgroundColor: c.progressPend }} />
                </View>
                <Text style={[t.overline, { fontSize: 12 }]}>{rep.votes}/{NEED_VOTES} {rep.votes === 1 ? 'confirma' : 'confirmam'}</Text>
              </View>
            )}
            {stage === 'final' && (
              <View style={{ gap: 6 }}>
                <View style={{ height: 10, borderRadius: 5, backgroundColor: c.sunken, overflow: 'hidden' }}>
                  <View style={{ height: '100%', width: `${Math.min(100, (av.resolved / (av.needed || 1)) * 100)}%`, borderRadius: 5, backgroundColor: c.primary }} />
                </View>
                <Text style={[t.overline, { fontSize: 12 }]}>{av.resolved}/{av.needed} {av.resolved === 1 ? 'diz' : 'dizem'} já resolvido</Text>
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
                  {stage === 'valid' && rep.mine && !my ? 'Relato seu — quem está por perto é quem vota.' : app.offline ? 'Voto salvo no aparelho. Envia quando a rede voltar.' : 'Seu voto foi registrado.'}
                </Text>
              </View>
            )}
            {!app.authed && votable && <Button variant="outline" size="md" label="Entrar para votar" onPress={() => nav.navigate('Login')} />}
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
