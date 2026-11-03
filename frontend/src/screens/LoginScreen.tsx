import React, { useState } from 'react';
import { KeyboardAvoidingView, ScrollView, Text, TextInput, View } from 'react-native';
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
  const { login, logout } = useApp();
  const [email, setEmail] = useState('exemplo@email.com');
  const [pass, setPass] = useState('');
  const [focus, setFocus] = useState<'email' | 'pass' | null>(null);
  const [err, setErr] = useState('');

  const submit = () => {
    if (!/^\S+@\S+\.\S+$/.test(email)) return setErr('Digite um e-mail válido.');
    if (pass.length < 8) return setErr('A senha precisa ter pelo menos 8 caracteres.');
    login();
    nav.replace('Tabs');
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
            <Text style={[t.body, { color: c.ink2, textAlign: 'center', maxWidth: 250 }]}>Riscos ambientais da sua cidade, relatados por quem vive nela.</Text>
          </View>
          <View style={{ gap: 12 }}>
            <Field label="E-mail" focused={focus === 'email'} value={email} onChangeText={v => { setEmail(v); setErr(''); }} onFocus={() => setFocus('email')} onBlur={() => setFocus(null)} keyboardType="email-address" autoCapitalize="none" autoComplete="email" />
            <Field label="Senha" focused={focus === 'pass'} value={pass} onChangeText={v => { setPass(v); setErr(''); }} onFocus={() => setFocus('pass')} onBlur={() => setFocus(null)} secureTextEntry placeholder="••••••••" />
            {!!err && <Text style={{ fontFamily: f.semibold, fontSize: 13, color: c.dangerText }}>{err}</Text>}
            <Button label="Entrar" onPress={submit} style={{ marginTop: 8 }} />
            <Button label="Só consultar o mapa, sem conta" variant="ghost" size="md" onPress={() => { logout(); nav.replace('Tabs'); }} />
          </View>
          <Text style={[t.caption, { marginTop: 'auto', paddingTop: 20, textAlign: 'center', fontFamily: f.medium }]}>
            Qualquer pessoa consulta o mapa. Para criar relatos e votar, é preciso ter uma conta.
          </Text>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
