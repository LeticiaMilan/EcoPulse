import React from 'react';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { NavigationContainer, NavigationIndependentTree, DefaultTheme } from '@react-navigation/native';
import { useFonts, Figtree_400Regular, Figtree_500Medium, Figtree_600SemiBold, Figtree_700Bold, Figtree_800ExtraBold } from '@expo-google-fonts/figtree';
import { IBMPlexMono_500Medium } from '@expo-google-fonts/ibm-plex-mono';
import { AppProvider } from '../store/AppContext';
import RootNavigator from '../navigation/RootNavigator';
import { Toast } from '../components/ui';
import { c } from '../theme';

export default function RootLayout() {
  const [loaded] = useFonts({ Figtree_400Regular, Figtree_500Medium, Figtree_600SemiBold, Figtree_700Bold, Figtree_800ExtraBold, IBMPlexMono_500Medium });
  if (!loaded) return null;
  return (
    <SafeAreaProvider>
      <AppProvider>
        <NavigationIndependentTree>
          <NavigationContainer theme={{ ...DefaultTheme, colors: { ...DefaultTheme.colors, background: c.bg, primary: c.primary } }}>
            <RootNavigator />
          </NavigationContainer>
        </NavigationIndependentTree>
        <Toast />
      </AppProvider>
      <StatusBar style="dark" />
    </SafeAreaProvider>
  );
}
