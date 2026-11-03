import React from 'react';
import { Pressable, Text, View } from 'react-native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { createBottomTabNavigator, BottomTabBarProps } from '@react-navigation/bottom-tabs';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { MaterialIcons } from '@expo/vector-icons';
import { c, f } from '../theme';
import { RootStackParamList, TabParamList } from './types';
import LoginScreen from '../screens/LoginScreen';
import MapScreen from '../screens/MapScreen';
import FiltersScreen from '../screens/FiltersScreen';
import NewReportScreen from '../screens/NewReportScreen';
import DetailScreen from '../screens/DetailScreen';
import ListScreen from '../screens/ListScreen';
import AlertsScreen from '../screens/AlertsScreen';
import SyncScreen from '../screens/SyncScreen';
import ProfileScreen from '../screens/ProfileScreen';

const Stack = createNativeStackNavigator<RootStackParamList>();
const Tab = createBottomTabNavigator<TabParamList>();

const TABS: Record<keyof TabParamList, { label: string; icon: any; outline: any; badge?: number }> = {
  Map: { label: 'Mapa', icon: 'map', outline: 'map' },
  List: { label: 'Relatos', icon: 'list-alt', outline: 'list-alt' },
  Alerts: { label: 'Alertas', icon: 'notifications', outline: 'notifications-none', badge: 2 },
  Profile: { label: 'Perfil', icon: 'person', outline: 'person-outline' },
};

function TabBar({ state, navigation }: BottomTabBarProps) {
  const insets = useSafeAreaInsets();
  return (
    <View style={{ flexDirection: 'row', backgroundColor: c.nav, borderTopWidth: 1, borderTopColor: c.divider, paddingTop: 8, paddingBottom: insets.bottom + 8, paddingHorizontal: 8 }}>
      {state.routes.map((route, i) => {
        const on = state.index === i;
        const tab = TABS[route.name as keyof TabParamList];
        return (
          <Pressable key={route.key} onPress={() => navigation.navigate(route.name)} style={{ flex: 1, alignItems: 'center', gap: 4 }}>
            <View style={{ width: 60, height: 32, borderRadius: 16, alignItems: 'center', justifyContent: 'center', backgroundColor: on ? c.container : 'transparent' }}>
              <MaterialIcons name={on ? tab.icon : tab.outline} size={23} color={on ? c.onContainer : c.ink2} />
              {!!tab.badge && (
                <View style={{ position: 'absolute', top: 2, right: 12, minWidth: 16, height: 16, borderRadius: 8, backgroundColor: c.danger, alignItems: 'center', justifyContent: 'center', paddingHorizontal: 4 }}>
                  <Text style={{ fontFamily: f.extrabold, fontSize: 10, color: '#fff' }}>{tab.badge}</Text>
                </View>
              )}
            </View>
            <Text style={{ fontFamily: on ? f.bold : f.medium, fontSize: 12, color: on ? c.ink : c.ink2 }}>{tab.label}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

function Tabs() {
  return (
    <Tab.Navigator tabBar={p => <TabBar {...p} />} screenOptions={{ headerShown: false }}>
      <Tab.Screen name="Map" component={MapScreen} />
      <Tab.Screen name="List" component={ListScreen} />
      <Tab.Screen name="Alerts" component={AlertsScreen} />
      <Tab.Screen name="Profile" component={ProfileScreen} />
    </Tab.Navigator>
  );
}

export default function RootNavigator() {
  return (
    <Stack.Navigator screenOptions={{ headerShown: false, contentStyle: { backgroundColor: c.bg }, animation: 'slide_from_right' }}>
      <Stack.Screen name="Login" component={LoginScreen} />
      <Stack.Screen name="Tabs" component={Tabs} options={{ animation: 'fade' }} />
      <Stack.Screen name="NewReport" component={NewReportScreen} options={{ animation: 'slide_from_bottom' }} />
      <Stack.Screen name="Detail" component={DetailScreen} />
      <Stack.Screen name="Sync" component={SyncScreen} />
      <Stack.Screen name="Filters" component={FiltersScreen} options={{ presentation: 'transparentModal', animation: 'fade', contentStyle: { backgroundColor: 'transparent' } }} />
    </Stack.Navigator>
  );
}
