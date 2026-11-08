import React, { useState } from 'react';
import { ActivityIndicator, KeyboardAvoidingView, Pressable, ScrollView, Text, TextInput, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useNavigation } from '@react-navigation/native';
import { c, f, t, r } from '../theme';
import { Button, Icon } from '../components/ui';
import { useApp } from '../store/AppContext';
import type { RootNav } from '../navigation/types';

function Field({ label, focused, ...props }: React.ComponentProps<typeof TextInput> & { label: string; focused: boolean }) {
  return (
    <View style={{ gap: 6 }}>
      <Text style={t.label}>{label}</Text>
      <TextInput
        placeholderTextColor={c.ink3}
        style={{ height: 52, borderRadius: r.input, borderWidth: 1.5, borderColor: focused ? c.primary : c.border, backgroundColor: c.surface, paddingHorizontal: 16, fontFamily: f.medium, fontSize: 16, color: c.ink }}
        {...props}
      />
    </View>
  );
}

export default function LoginScreen() {
  const nav = useNavigation<RootNav>();
  const { login, register, logout } = useApp();
  const [mode, setMode] = useState<'login' | 'signup'>('login');
  const [username, setUsername] = useState('');
  const [pass, setPass] = useState('');
  const [focus, setFocus] = useState<'user' | 'pass' | null>(null);
  const [err, setErr] = useState('');
  const [loading, setLoading] = useState(false);

  const submit = async () => {
    if (username.trim().length < 3) return setErr('O usuário precisa ter pelo menos 3 caracteres.');
    if (pass.length < 4) return setErr('A senha precisa ter pelo menos 4 caracteres.');
    setLoading(true);
    setErr('');
    try {
      if (mode === 'login') await login(username.trim(), pass);
      else await register(username.trim(), pass);
      nav.reset({ index: 0, routes: [{ name: 'Tabs' }] });
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: c.bg }}>
      <KeyboardAvoidingView style={{ flex: 1 }} behavior="padding">
        <ScrollView contentContainerStyle={{ flexGrow: 1, padding: 24, paddingTop: 28 }} keyboardShouldPersistTaps="handled">
          <View style={{ alignItems: 'center', gap: 14, paddingTop: 28, paddingBottom: 32 }}>
            <View style={{ width: 84, height: 84, borderRadius: 28, backgroundColor: c.primary, alignItems: 'center', justifyContent: 'center' }}>
              <Icon name="eco" size={46} color="#fff" />
            </View>
            <Text style={t.display}>EcoPulse</Text>
            <Text style={[t.body, { color: c.ink2, textAlign: 'center', maxWidth: 250 }]}>
              {mode === 'login' ? 'Riscos ambientais da sua cidade, relatados por quem vive nela.' : 'Crie sua conta para relatar e votar.'}
            </Text>
          </View>
          <View style={{ gap: 12 }}>
            <Field label="Usuário" focused={focus === 'user'} value={username}
              onChangeText={v => { setUsername(v); setErr(''); }} onFocus={() => setFocus('user')} onBlur={() => setFocus(null)}
              autoCapitalize="none" autoCorrect={false} placeholder="usuário" />
            <Field label="Senha" focused={focus === 'pass'} value={pass}
              onChangeText={v => { setPass(v); setErr(''); }} onFocus={() => setFocus('pass')} onBlur={() => setFocus(null)}
              secureTextEntry maxLength={72} placeholder="••••••••" />
            {!!err && <Text style={{ fontFamily: f.semibold, fontSize: 13, color: c.dangerText }}>{err}</Text>}
            {loading
              ? <View style={{ height: 54, marginTop: 8, alignItems: 'center', justifyContent: 'center' }}><ActivityIndicator color={c.primary} /></View>
              : <Button label={mode === 'login' ? 'Entrar' : 'Criar conta'} onPress={submit} style={{ marginTop: 8 }} />}
            <Pressable onPress={() => { setMode(mode === 'login' ? 'signup' : 'login'); setErr(''); }} style={{ height: 44, alignItems: 'center', justifyContent: 'center' }}>
              <Text style={{ fontFamily: f.semibold, fontSize: 14, color: c.ink2 }}>
                {mode === 'login' ? 'Não tem conta? ' : 'Já tem conta? '}
                <Text style={{ fontFamily: f.bold, color: c.primaryStrong }}>{mode === 'login' ? 'Criar conta' : 'Entrar'}</Text>
              </Text>
            </Pressable>
            <Button label="Só consultar o mapa, sem conta" variant="ghost" size="md" onPress={() => { logout(); nav.reset({ index: 0, routes: [{ name: 'Tabs' }] }); }} />
          </View>
          <Text style={[t.caption, { marginTop: 'auto', paddingTop: 20, textAlign: 'center', fontFamily: f.medium }]}>
            Qualquer pessoa consulta o mapa. Para criar relatos e votar, é preciso ter uma conta.
          </Text>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}