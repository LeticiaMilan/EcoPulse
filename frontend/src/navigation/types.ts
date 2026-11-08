import type { NavigatorScreenParams } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';

export type TabParamList = { Map: undefined; List: undefined; Alerts: undefined; Profile: undefined };
export type RootStackParamList = {
  Login: undefined;
  Tabs: NavigatorScreenParams<TabParamList> | undefined;
  NewReport: undefined;
  Detail: undefined;
  Filters: undefined;
  Sync: undefined;
};
export type RootNav = NativeStackNavigationProp<RootStackParamList>;
