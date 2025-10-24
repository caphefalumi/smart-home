<template>
  <v-app>
    <!-- Minimalist App Bar -->
    <v-app-bar color="white" flat height="56" class="minimal-app-bar">
      <v-app-bar-nav-icon @click="drawer = !drawer" icon="mdi-menu" color="grey-darken-2"></v-app-bar-nav-icon>
      <v-toolbar-title class="font-weight-bold text-grey-darken-3 ml-2">Smart Home</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-chip :color="isConnected ? 'success' : 'grey-lighten-2'" class="status-chip" size="small" variant="tonal">
        <v-icon size="18" class="mr-1">{{ isConnected ? 'mdi-check-circle' : 'mdi-alert-circle' }}</v-icon>
        {{ isConnected ? 'Online' : 'Offline' }}
      </v-chip>
      <v-btn icon="mdi-theme-light-dark" @click="toggleTheme" variant="text" color="grey-darken-2"></v-btn>
    </v-app-bar>

    <!-- Minimalist Navigation Drawer -->
    <v-navigation-drawer v-model="drawer" temporary width="180" class="minimal-drawer">
      <v-list nav density="compact">
        <v-list-item v-for="item in navigationItems" :key="item.value" :prepend-icon="item.icon" :title="item.title" @click="currentTab = item.value; drawer = false" :active="currentTab === item.value" rounded="lg" class="nav-item-minimal" style="padding: 8px 0;">
          <template #prepend>
            <v-icon size="22" :color="currentTab === item.value ? 'primary' : 'grey-darken-1'">{{ item.icon }}</v-icon>
          </template>
          <span class="nav-title" :style="{ fontWeight: currentTab === item.value ? 'bold' : 'normal', color: currentTab === item.value ? '#1976d2' : '#333' }">{{ item.title }}</span>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <v-main class="minimal-main">
      <v-container fluid class="pa-4">
        <!-- Connection Banner -->
        <v-alert v-if="!isConnected && !isLoading" type="warning" border="start" class="mb-4" dense>
          <v-alert-title>Device Not Connected</v-alert-title>
          Connect to your Arduino device to begin.
        </v-alert>
        <v-alert v-if="isLoading" type="info" border="start" class="mb-4" dense>
          <v-alert-title>Connecting...</v-alert-title>
        </v-alert>

        <!-- Device Connection Card -->
        <v-card class="mb-4 minimal-card" flat>
          <v-card-title class="pa-3 font-weight-bold text-grey-darken-3">Device Connection</v-card-title>
          <v-card-text class="pa-3">
            <v-row align="center">
              <v-col cols="12" md="4">
                <v-select v-model="selectedPort" :items="availablePortsFormatted" item-title="displayName" item-value="path" label="Port" :disabled="isConnected" variant="outlined" density="compact" clearable></v-select>
              </v-col>
              <v-col cols="12" md="8">
                <div class="d-flex gap-2 flex-wrap">
                  <v-btn @click="refreshPorts" :disabled="isConnected" color="grey-darken-2" variant="tonal">Refresh</v-btn>
                  <v-btn v-if="!isConnected" @click="connect" :disabled="!selectedPort || isLoading" :loading="isLoading" color="primary" variant="flat">Connect</v-btn>
                  <v-btn v-else @click="disconnect" color="error" variant="flat">Disconnect</v-btn>
                </div>
              </v-col>
            </v-row>
          </v-card-text>
        </v-card>

        <!-- Minimal Tabs -->
        <v-tabs v-model="currentTab" color="primary" align-tabs="center" class="minimal-tabs mb-4">
          <v-tab v-for="tab in navigationItems" :key="tab.value" :value="tab.value" class="tab-item-minimal">
            <v-icon class="mr-1">{{ tab.icon }}</v-icon>
            {{ tab.title }}
          </v-tab>
        </v-tabs>

        <!-- Tab Content -->
        <v-window v-model="currentTab" class="minimal-window">
          <!-- Dashboard Tab -->
          <v-window-item value="dashboard">
            <v-row class="mb-2">
              <v-col v-for="sensor in sensorTypes" :key="sensor.key" cols="12" sm="6" md="3">
                <v-card flat class="minimal-sensor-card pa-3">
                  <div class="d-flex align-center mb-2">
                    <v-icon size="32" class="mr-2" :color="sensor.color">{{ sensor.icon }}</v-icon>
                    <span class="font-weight-bold">{{ sensor.label }}</span>
                  </div>
                  <div class="text-h4 font-weight-bold mb-1">{{ currentData?.[sensor.key] || 0 }}</div>
                  <div class="text-caption text-grey-darken-1">{{ sensor.unit }}</div>
                </v-card>
              </v-col>
            </v-row>
            <v-card flat class="pa-3 mb-2 minimal-status-card">
              <div class="d-flex align-center">
                <v-icon :color="getOverallStatusColor()" class="mr-2" size="24">{{ getOverallStatusIcon() }}</v-icon>
                <span class="font-weight-bold">{{ getOverallStatus() }}</span>
                <v-spacer></v-spacer>
                <span class="text-caption">{{ formatTime(new Date().toISOString()) }}</span>
              </div>
            </v-card>
          </v-window-item>

          <!-- Controls Tab -->
          <v-window-item value="controls">
            <v-row>
              <v-col v-for="control in deviceControls" :key="control.title" cols="12" sm="6" md="4">
                <v-card flat class="minimal-control-card pa-3">
                  <div class="d-flex align-center mb-2">
                    <v-icon size="32" :color="control.iconColor" class="mr-2">{{ control.icon }}</v-icon>
                    <span class="font-weight-bold">{{ control.title }}</span>
                  </div>
                  <div class="d-flex gap-2">
                    <v-btn @click="sendCommand(control.onCommand)" :color="control.onColor" variant="flat" size="small" :disabled="!isConnected">{{ control.onLabel }}</v-btn>
                    <v-btn @click="sendCommand(control.offCommand)" :color="control.offColor" variant="tonal" size="small" :disabled="!isConnected">{{ control.offLabel }}</v-btn>
                  </div>
                  <v-slider v-if="control.hasAdvanced && control.type === 'servo'" v-model="control.value" :min="0" :max="180" :step="5" thumb-label label="Angle" @update:model-value="sendAdvancedCommand(control, $event)" :disabled="!isConnected" class="mt-2"></v-slider>
                  <v-slider v-if="control.hasAdvanced && control.type === 'pwm'" v-model="control.value" :min="0" :max="255" :step="5" thumb-label label="Level" @update:model-value="sendAdvancedCommand(control, $event)" :disabled="!isConnected" class="mt-2"></v-slider>
                </v-card>
              </v-col>
            </v-row>
            
            <!-- Music Controls Section -->
            <v-divider class="my-4"></v-divider>
            <h3 class="mb-3 text-grey-darken-2">🎵 Music & Entertainment</h3>
            <v-row>
              <v-col v-for="music in musicControls" :key="music.title" cols="12" sm="6" md="4">
                <v-card flat class="minimal-control-card pa-3">
                  <div class="d-flex align-center mb-2">
                    <v-icon size="32" :color="music.iconColor" class="mr-2">{{ music.icon }}</v-icon>
                    <span class="font-weight-bold">{{ music.title }}</span>
                  </div>
                  <v-btn @click="music.action()" :color="music.color" variant="flat" size="small" :disabled="!isConnected" block>
                    {{ music.label }}
                  </v-btn>
                </v-card>
              </v-col>
            </v-row>
          </v-window-item>

          <!-- Analytics Tab -->
          <v-window-item value="analytics">
            <v-row>
              <v-col cols="12" sm="6" md="4">
                <v-select v-model="analyticsHours" :items="analyticsOptions" label="Time Range" variant="outlined" density="compact"></v-select>
              </v-col>
              <v-col cols="12" sm="6" md="4">
                <v-select v-model="selectedSensor" :items="analyticsChartOptions" label="Chart View" variant="outlined" density="compact"></v-select>
              </v-col>
            </v-row>
            <v-row class="mt-2">
              <v-col v-for="sensor in sensorTypes" :key="sensor.key" cols="12" sm="6" md="3">
                <v-card flat class="minimal-analytics-card pa-3">
                  <div class="font-weight-bold mb-1">{{ sensor.label }}</div>
                  <div class="d-flex gap-2">
                    <span class="text-caption">Avg:</span>
                    <span class="font-weight-bold">{{ getStatValue(sensor.key, 'mean') }}</span>
                    <span class="text-caption">Min:</span>
                    <span class="font-weight-bold">{{ getStatValue(sensor.key, 'min') }}</span>
                    <span class="text-caption">Max:</span>
                    <span class="font-weight-bold">{{ getStatValue(sensor.key, 'max') }}</span>
                  </div>
                </v-card>
              </v-col>
            </v-row>
            <v-card flat class="minimal-analytics-card pa-3 mt-2">
              <div class="d-flex align-center justify-space-between mb-2">
                <span class="font-weight-bold">Sensor Trends</span>
                <span class="text-caption text-grey-darken-1">Last {{ analyticsHours }} hours</span>
              </div>
              <SensorTrendChart :labels="trendChartLabels" :datasets="trendChartDatasets" :loading="trendsLoading" />
            </v-card>
          </v-window-item>

          <!-- Rules Tab -->
          <v-window-item value="rules">
            <v-btn @click="showRuleForm = !showRuleForm" :color="showRuleForm ? 'error' : 'primary'" class="mb-3" variant="flat" size="small">
              {{ showRuleForm ? 'Cancel' : 'New Rule' }}
            </v-btn>
            <v-expand-transition>
              <v-card v-if="showRuleForm" flat class="mb-3 pa-3 minimal-rule-form">
                <v-row>
                  <v-col cols="12" md="6">
                    <v-text-field v-model="newRule.name" label="Name" variant="outlined" clearable></v-text-field>
                  </v-col>
                  <v-col cols="12" md="6">
                    <v-select
                      v-model="newRule.sensor"
                      :items="ruleSensorOptions"
                      item-title="label"
                      item-value="key"
                      label="Sensor"
                      variant="outlined"
                    ></v-select>
                  </v-col>
                  <v-col cols="12" md="4">
                    <v-select v-model="newRule.operator" :items="operatorOptions" label="Condition" variant="outlined"></v-select>
                  </v-col>
                  <v-col cols="12" md="4">
                    <v-text-field v-model.number="newRule.threshold" type="number" label="Threshold" variant="outlined"></v-text-field>
                  </v-col>
                  <v-col cols="12" md="4">
                    <v-select v-model="newRule.action" :items="actionOptions" label="Action" variant="outlined"></v-select>
                  </v-col>
                </v-row>
                <div class="d-flex gap-2 mt-2">
                  <v-btn @click="createRule" color="success" variant="flat" size="small">Save</v-btn>
                  <v-btn @click="resetRuleForm" color="grey" variant="tonal" size="small">Reset</v-btn>
                </div>
              </v-card>
            </v-expand-transition>
            <v-row v-if="rules.length">
              <v-col v-for="rule in rules" :key="rule._id" cols="12" md="6">
                <v-card flat class="minimal-rule-card pa-3 mb-2">
                  <div class="d-flex align-center justify-space-between">
                    <span class="font-weight-bold">{{ rule.name }}</span>
                    <v-switch :model-value="rule.enabled" @update:model-value="toggleRule(rule._id?.toString(), $event)" color="success" density="compact" hide-details></v-switch>
                  </div>
                  <div class="text-caption mt-1">IF {{ getSensorLabel(rule.sensor) }} {{ rule.operator }} {{ rule.threshold }} THEN {{ getActionLabel(rule.action) }}</div>
                  <div v-if="rule.description" class="text-caption mt-1">{{ rule.description }}</div>
                  <div class="d-flex align-center justify-space-between mt-1">
                    <span class="text-caption">{{ formatTime(rule.createdAt) }}</span>
                    <v-btn @click="deleteRule(rule._id?.toString())" color="error" size="x-small" icon="mdi-delete-outline" variant="text"></v-btn>
                  </div>
                </v-card>
              </v-col>
            </v-row>
            <v-alert v-else type="info" variant="tonal" class="ma-2">No rules yet.</v-alert>
          </v-window-item>

          <!-- Alerts Tab -->
          <v-window-item value="alerts">
            <v-row>
              <v-col cols="12" sm="6" md="3">
                <v-card flat color="error" dark class="text-center pa-2 minimal-alert-card">
                  <div class="text-h5 font-weight-bold">{{ criticalAlerts }}</div>
                  <div class="text-caption">Critical</div>
                </v-card>
              </v-col>
              <v-col cols="12" sm="6" md="3">
                <v-card flat color="warning" dark class="text-center pa-2 minimal-alert-card">
                  <div class="text-h5 font-weight-bold">{{ warningAlerts }}</div>
                  <div class="text-caption">Warning</div>
                </v-card>
              </v-col>
              <v-col cols="12" sm="6" md="3">
                <v-card flat color="info" dark class="text-center pa-2 minimal-alert-card">
                  <div class="text-h5 font-weight-bold">{{ infoAlerts }}</div>
                  <div class="text-caption">Info</div>
                </v-card>
              </v-col>
              <v-col cols="12" sm="6" md="3">
                <v-card flat color="success" dark class="text-center pa-2 minimal-alert-card">
                  <div class="text-h5 font-weight-bold">{{ recentAlerts.length }}</div>
                  <div class="text-caption">Total</div>
                </v-card>
              </v-col>
            </v-row>
            <v-list v-if="recentAlerts.length" class="py-0 mt-2">
              <v-list-item v-for="(alert, index) in recentAlerts" :key="alert._id" class="minimal-alert-item">
                <v-avatar :color="getAlertColor(alert)" size="32" class="mr-2">
                  <v-icon color="white">{{ getAlertIcon(alert) }}</v-icon>
                </v-avatar>
                <v-list-item-title class="font-weight-bold">{{ formatTime(alert.timestamp) }}</v-list-item-title>
                <v-list-item-subtitle>
                  <div v-for="(msg, idx) in alert.alerts" :key="idx" class="alert-message text-caption">{{ msg }}</div>
                </v-list-item-subtitle>
                <v-chip :color="getAlertSeverity(alert)" size="x-small" variant="tonal">{{ getAlertSeverityText(alert) }}</v-chip>
              </v-list-item>
            </v-list>
            <v-alert v-else type="info" variant="tonal" class="ma-2">No alerts.</v-alert>
          </v-window-item>
        </v-window>
      </v-container>
    </v-main>

    <!-- Minimal FAB -->
    <v-speed-dial v-model="speedDial" location="bottom end" transition="fade-transition" class="minimal-fab">
      <v-btn key="1" size="small" color="primary" icon="mdi-lightbulb" @click="sendCommand('white_light_on')"></v-btn>
      <v-btn key="2" size="small" color="info" icon="mdi-fan" @click="sendCommand('fan_on')"></v-btn>
    </v-speed-dial>

    <!-- Snackbar -->
    <v-snackbar v-model="snackbar" :color="snackbarColor" :timeout="4000" location="top" rounded="lg">
      {{ snackbarText }}
      <template v-slot:actions>
        <v-btn icon="mdi-close" size="small" @click="snackbar = false"></v-btn>
      </template>
    </v-snackbar>
  </v-app>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue';
import { useTheme } from 'vuetify';
import SensorTrendChart from './components/SensorTrendChart.vue';

const theme = useTheme();
const API_URL = 'http://localhost:3000/api';

interface TrendPoint {
  hour: string;
  light?: number;
  gas?: number;
  soil?: number;
  water?: number;
  count?: number;
}

const sensorKeys = ['light', 'gas', 'soil', 'water', 'infrar'] as const;
type SensorKey = typeof sensorKeys[number];

// Enhanced State Management
const isConnected = ref(false);
const selectedPort = ref('');
const availablePorts = ref<any[]>([]);
const currentData = ref<any>({
  light: 0,
  gas: 0,
  soil: 0,
  water: 0,
  infrar: 0,
  btn1: 0,
  btn2: 0
});
const actuatorStates = ref<any>({
  white_light: false,
  yellow_light: false,
  relay: false,
  door_angle: 0,
  window_angle: 0,
  fan: false,
  fan_speed: 0,
  buzzer: false
});
const rules = ref<any[]>([]);
const recentAlerts = ref<any[]>([]);
const statistics = ref<any>({});
const analyticsHours = ref(24);
const selectedSensor = ref('all');
const trendData = ref<TrendPoint[]>([]);
const trendsLoading = ref(false);
const sensorColorMap: Record<SensorKey, { border: string; background: string }> = {
  light: { border: '#ffca28', background: 'rgba(255, 202, 40, 0.2)' },
  gas: { border: '#ef5350', background: 'rgba(239, 83, 80, 0.2)' },
  soil: { border: '#66bb6a', background: 'rgba(102, 187, 106, 0.2)' },
  water: { border: '#42a5f5', background: 'rgba(66, 165, 245, 0.2)' }
};
const showRuleForm = ref(false);
const currentTab = ref('dashboard');
const drawer = ref(false);
const speedDial = ref(false);
const isLoading = ref(false);
const snackbar = ref(false);
const snackbarText = ref('');
const snackbarColor = ref('success');

// Alert statistics
const criticalAlerts = ref(0);
const warningAlerts = ref(0);
const infoAlerts = ref(0);

const newRule = ref({
  name: '',
  sensor: 'gas',
  operator: '>',
  threshold: 0,
  action: 'white_light_on',
  enabled: true,
  description: ''
});

// Navigation Items with badges
const navigationItems = computed(() => [
  { 
    value: 'dashboard', 
    title: 'Dashboard', 
    subtitle: 'Real-time monitoring',
    icon: 'mdi-view-dashboard' 
  },
  { 
    value: 'controls', 
    title: 'Device Control', 
    subtitle: 'Manage actuators',
    icon: 'mdi-tune' 
  },
  { 
    value: 'analytics', 
    title: 'Analytics', 
    subtitle: 'Data insights',
    icon: 'mdi-chart-line' 
  },
  { 
    value: 'rules', 
    title: 'Automation', 
    subtitle: 'Smart rules',
    icon: 'mdi-cog-outline',
    badge: rules.value.filter(r => r.enabled).length || undefined
  },
  { 
    value: 'alerts', 
    title: 'Alerts', 
    subtitle: 'System notifications',
    icon: 'mdi-bell-outline',
    badge: recentAlerts.value.length || undefined
  }
]);

// Enhanced computed properties
const availablePortsFormatted = computed(() => 
  availablePorts.value.map(port => ({
    ...port,
    displayName: `${port.name}`
  }))
);

// Enhanced sensor types with more metadata
const sensorTypes = [
  { 
    key: 'light', 
    label: 'Light Level', 
    icon: 'mdi-lightbulb-outline',
    color: 'amber',
    unit: 'lux',
    thresholds: { low: 200, high: 800 }
  },
  { 
    key: 'gas', 
    label: 'Gas Level', 
    icon: 'mdi-gas-cylinder',
    color: 'red',
    unit: 'ppm',
    thresholds: { low: 300, high: 700 }
  },
  { 
    key: 'soil', 
    label: 'Soil Moisture', 
    icon: 'mdi-sprout-outline',
    color: 'green',
    unit: '%',
    thresholds: { low: 30, high: 70 }
  },
  { 
    key: 'water', 
    label: 'Water Level', 
    icon: 'mdi-water-outline',
    color: 'blue',
    unit: 'mm',
    thresholds: { low: 100, high: 800 }
  },
  {
    key: 'infrar',
    label: 'Infrared Sensor',
    icon: 'mdi-remote',
    color: 'grey',
    unit: '',
    thresholds: { low: 0, high: 1 }
  }
];

const ruleSensorOptions = computed(() =>
  sensorTypes.filter(sensor => sensorKeys.includes(sensor.key as SensorKey))
);

const trendChartLabels = computed(() => trendData.value.map(point => point.hour));

const trendChartDatasets = computed(() => {
  if (!trendData.value.length) {
    return [];
  }

  const keys: SensorKey[] = selectedSensor.value === 'all'
    ? [...sensorKeys]
    : sensorKeys.includes(selectedSensor.value as SensorKey)
      ? [selectedSensor.value as SensorKey]
      : [...sensorKeys];

  return keys.map((key) => {
    const palette = sensorColorMap[key];
    return {
      label: getSensorLabel(key),
      data: trendData.value.map((point) => {
        const rawValue = (point as Record<string, unknown>)[key];
        return typeof rawValue === 'number' ? Math.round(rawValue * 100) / 100 : 0;
      }),
      borderColor: palette.border,
      backgroundColor: palette.background,
      tension: 0.3,
      fill: false
    };
  });
});

// Quick actions for dashboard
const quickActions = [
  { command: 'white_light_on', label: 'Lights ON', icon: 'mdi-lightbulb', color: 'success' },
  { command: 'fan_on', label: 'Fan ON', icon: 'mdi-fan', color: 'info' },
  { command: 'door_open', label: 'Open Door', icon: 'mdi-door-open', color: 'warning' },
  { command: 'buzzer_off', label: 'Silence All', icon: 'mdi-volume-off', color: 'error' }
];

// Enhanced device controls with all backend actuators
const deviceControls = [
  {
    title: 'White Light',
    icon: 'mdi-lightbulb',
    iconColor: 'amber',
    onCommand: 'white_light_on',
    offCommand: 'white_light_off',
    onLabel: 'Turn ON',
    offLabel: 'Turn OFF',
    onIcon: 'mdi-lightbulb-on',
    offIcon: 'mdi-lightbulb-off',
    onColor: 'success',
    offColor: 'error',
    type: 'pwm',
    hasAdvanced: true,
    value: 255
  },
  {
    title: 'Yellow Light',
    icon: 'mdi-lightbulb-variant',
    iconColor: 'yellow-darken-2',
    onCommand: 'yellow_light_on',
    offCommand: 'yellow_light_off',
    onLabel: 'Turn ON',
    offLabel: 'Turn OFF',
    onIcon: 'mdi-lightbulb-on',
    offIcon: 'mdi-lightbulb-off',
    onColor: 'success',
    offColor: 'error',
    type: 'pwm',
    hasAdvanced: true,
    value: 255
  },
  {
    title: 'Smart Relay',
    icon: 'mdi-electric-switch-closed',
    iconColor: 'purple',
    onCommand: 'relay_on',
    offCommand: 'relay_off',
    onLabel: 'Activate',
    offLabel: 'Deactivate',
    onIcon: 'mdi-power-plug',
    offIcon: 'mdi-power-plug-off',
    onColor: 'success',
    offColor: 'error'
  },
  {
    title: 'Ventilation Fan',
    icon: 'mdi-fan',
    iconColor: 'cyan',
    onCommand: 'fan_on',
    offCommand: 'fan_off',
    onLabel: 'Start Fan',
    offLabel: 'Stop Fan',
    onIcon: 'mdi-fan',
    offIcon: 'mdi-fan-off',
    onColor: 'success',
    offColor: 'error',
    type: 'pwm',
    hasAdvanced: true,
    value: 255
  },
  {
    title: 'Main Door',
    icon: 'mdi-door',
    iconColor: 'brown',
    onCommand: 'door_open',
    offCommand: 'door_close',
    onLabel: 'OPEN',
    offLabel: 'CLOSE',
    onIcon: 'mdi-door-open',
    offIcon: 'mdi-door-closed',
    onColor: 'success',
    offColor: 'warning',
    type: 'servo',
    hasAdvanced: true,
    value: 0
  },
  {
    title: 'Window Control',
    icon: 'mdi-window-shutter',
    iconColor: 'teal',
    onCommand: 'window_open',
    offCommand: 'window_close',
    onLabel: 'OPEN',
    offLabel: 'CLOSE',
    onIcon: 'mdi-window-open',
    offIcon: 'mdi-window-closed',
    onColor: 'success',
    offColor: 'warning',
    type: 'servo',
    hasAdvanced: true,
    value: 0
  },
  {
    title: 'Alert Buzzer',
    icon: 'mdi-volume-high',
    iconColor: 'red',
    onCommand: 'buzzer_on',
    offCommand: 'buzzer_off',
    onLabel: 'Sound Alert',
    offLabel: 'Silence',
    onIcon: 'mdi-volume-high',
    offIcon: 'mdi-volume-off',
    onColor: 'error',
    offColor: 'success'
  }
];

// Music controls for entertainment
const musicControls = [
  {
    title: 'Birthday Song',
    icon: 'mdi-cake-variant',
    iconColor: 'pink',
    action: () => playMusic('birthday'),
    label: 'Play Song',
    color: 'success'
  },
  {
    title: 'Ode to Joy',
    icon: 'mdi-music-note',
    iconColor: 'purple',
    action: () => playMusic('ode-to-joy'),
    label: 'Play Classical',
    color: 'info'
  },
  {
    title: 'Stop Music',
    icon: 'mdi-stop',
    iconColor: 'red',
    action: () => stopMusic(),
    label: 'Stop All',
    color: 'error'
  }
];

// Scene controls for automation
const sceneControls = [
  {
    name: 'Welcome Home',
    icon: 'mdi-home-heart',
    color: 'success',
    commands: ['white_light_on', 'door_open', 'fan_on']
  },
  {
    name: 'Security Mode',
    icon: 'mdi-shield-home',
    color: 'warning',
    commands: ['door_close', 'window_close', 'white_light_off']
  },
  {
    name: 'Sleep Mode',
    icon: 'mdi-sleep',
    color: 'indigo',
    commands: ['white_light_off', 'yellow_light_off', 'fan_off']
  },
  {
    name: 'Emergency',
    icon: 'mdi-alarm-light',
    color: 'error',
    commands: ['buzzer_on', 'white_light_on', 'door_open']
  }
];

const analyticsOptions = [
  { title: 'Last Hour', value: 1 },
  { title: 'Last 6 Hours', value: 6 },
  { title: 'Last 24 Hours', value: 24 },
  { title: 'Last Week', value: 168 }
];

const analyticsChartOptions = [
  { title: 'All Sensors', value: 'all' },
  { title: 'Light Level Only', value: 'light' },
  { title: 'Gas Level Only', value: 'gas' },
  { title: 'Soil Moisture Only', value: 'soil' },
  { title: 'Water Level Only', value: 'water' }
];

const operatorOptions = [
  { title: 'Greater than (>)', value: '>' },
  { title: 'Less than (<)', value: '<' },
  { title: 'Greater or equal (≥)', value: '>=' },
  { title: 'Less or equal (≤)', value: '<=' },
  { title: 'Equals (=)', value: '==' }
];

const actionOptions = [
  // Light controls
  { title: 'Turn White Light On', value: 'white_light_on' },
  { title: 'Turn White Light Off', value: 'white_light_off' },
  { title: 'Turn Yellow Light On', value: 'yellow_light_on' },
  { title: 'Turn Yellow Light Off', value: 'yellow_light_off' },
  // Power controls
  { title: 'Activate Relay', value: 'relay_on' },
  { title: 'Deactivate Relay', value: 'relay_off' },
  // Access controls
  { title: 'Open Door', value: 'door_open' },
  { title: 'Close Door', value: 'door_close' },
  { title: 'Open Window', value: 'window_open' },
  { title: 'Close Window', value: 'window_close' },
  // Climate controls
  { title: 'Start Fan', value: 'fan_on' },
  { title: 'Stop Fan', value: 'fan_off' },
  // Alert controls
  { title: 'Sound Alert', value: 'buzzer_on' },
  { title: 'Silence Alert', value: 'buzzer_off' }
];

let refreshInterval: any = null;

// Enhanced Functions
async function refreshPorts() {
  try {
    const response = await fetch(`${API_URL}/serial/ports`);
    availablePorts.value = await response.json();
  } catch (error) {
    console.error('Error fetching ports:', error);
    showSnackbar('Failed to fetch serial ports', 'error');
  }
}

async function connect() {
  try {
    isLoading.value = true;
    const response = await fetch(`${API_URL}/serial/connect`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ port: selectedPort.value, baudRate: 9600 })
    });
    
    if (response.ok) {
      isConnected.value = true;
      startDataRefresh();
      showSnackbar('Successfully connected to Arduino', 'success');
      loadActuatorStates();
    } else {
      const error = await response.json();
      showSnackbar(`Connection failed: ${error.error}`, 'error');
    }
  } catch (error) {
    console.error('Connection error:', error);
    showSnackbar('Failed to connect to Arduino', 'error');
  } finally {
    isLoading.value = false;
  }
}

async function disconnect() {
  try {
    await fetch(`${API_URL}/serial/disconnect`, { method: 'POST' });
    isConnected.value = false;
    stopDataRefresh();
    showSnackbar('Disconnected from Arduino', 'info');
  } catch (error) {
    console.error('Disconnect error:', error);
  }
}

async function sendCommand(command: string) {
  if (!isConnected.value) {
    showSnackbar('Please connect to Arduino first', 'warning');
    return;
  }

  try {
    await fetch(`${API_URL}/serial/command`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command })
    });
    showSnackbar(`Command sent: ${command}`, 'success');
  } catch (error) {
    console.error('Command error:', error);
    showSnackbar('Failed to send command', 'error');
  }
}

async function sendAdvancedCommand(control: any, value: number) {
  if (!isConnected.value) return;
  
  let command = '';
  if (control.type === 'servo') {
    command = control.title.toLowerCase().includes('door') 
      ? `door_angle=${value}` 
      : `window_angle=${value}`;
  } else if (control.type === 'pwm') {
    command = control.title.toLowerCase().includes('yellow') 
      ? `yellow_light_pwm=${value}` 
      : `fan_speed=${value}`;
  }
  
  if (command) {
    await sendCommand(command);
  }
}

// Music control functions
async function playMusic(type: 'birthday' | 'ode-to-joy') {
  if (!isConnected.value) {
    showSnackbar('Please connect to Arduino first', 'warning');
    return;
  }

  try {
    const endpoint = type === 'birthday' ? 'birthday' : 'ode-to-joy';
    const response = await fetch(`${API_URL}/music/${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    });
    
    if (response.ok) {
      const songName = type === 'birthday' ? 'Birthday Song' : 'Ode to Joy';
      showSnackbar(`🎵 Playing ${songName}`, 'success');
    } else {
      throw new Error('Failed to play music');
    }
  } catch (error) {
    console.error('Music play error:', error);
    showSnackbar('Failed to play music', 'error');
  }
}

async function stopMusic() {
  if (!isConnected.value) {
    showSnackbar('Please connect to Arduino first', 'warning');
    return;
  }

  try {
    const response = await fetch(`${API_URL}/music/stop`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    });
    
    if (response.ok) {
      showSnackbar('🔇 Music stopped', 'info');
    } else {
      throw new Error('Failed to stop music');
    }
  } catch (error) {
    console.error('Music stop error:', error);
    showSnackbar('Failed to stop music', 'error');
  }
}

async function executeScene(scene: any) {
  if (!isConnected.value) {
    showSnackbar('Please connect to Arduino first', 'warning');
    return;
  }

  for (const command of scene.commands) {
    await sendCommand(command);
    await new Promise(resolve => setTimeout(resolve, 200)); // Small delay between commands
  }
  showSnackbar(`Scene "${scene.name}" executed`, 'success');
}

async function loadCurrentData() {
  try {
    const response = await fetch(`${API_URL}/sensors/current`);
    if (response.ok) {
      const data = await response.json();
      currentData.value = { ...currentData.value, ...data };
    }
  } catch (error) {
    console.error('Error loading current data:', error);
  }
}

async function loadActuatorStates() {
  try {
    const response = await fetch(`${API_URL}/actuators/states`);
    if (response.ok) {
      actuatorStates.value = await response.json();
    }
  } catch (error) {
    console.error('Error loading actuator states:', error);
  }
}

async function loadRules() {
  try {
    const response = await fetch(`${API_URL}/rules`);
    if (response.ok) {
      const rulesData = await response.json();
      console.log('Loaded rules:', rulesData);
      // Log the structure of each rule to see the _id format
      rulesData.forEach((rule: any, index: number) => {
        console.log(`Rule ${index}:`, {
          _id: rule._id,
          _id_type: typeof rule._id,
          name: rule.name
        });
      });
      rules.value = rulesData;
    }
  } catch (error) {
    console.error('Error loading rules:', error);
  }
}

async function createRule() {
  try {
    const response = await fetch(`${API_URL}/rules`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(newRule.value)
    });
    
    if (response.ok) {
      showRuleForm.value = false;
      resetRuleForm();
      loadRules();
      showSnackbar('Rule created successfully', 'success');
    } else {
      let message = 'Failed to create rule';
      try {
        const errorData = await response.json();
        if (errorData?.error) {
          message = errorData.error;
        }
      } catch (parseError) {
        console.error('Error parsing rule creation error:', parseError);
      }
      throw new Error(message);
    }
  } catch (error) {
    console.error('Error creating rule:', error);
    showSnackbar('Failed to create rule', 'error');
  }
}

function resetRuleForm() {
  newRule.value = {
    name: '',
    sensor: 'gas',
    operator: '>',
    threshold: 0,
    action: 'white_light_on',
    enabled: true,
    description: ''
  };
}

async function toggleRule(id: string, enabled: boolean) {
  try {
    const ruleId = id?.toString() || '';
    if (!ruleId) {
      showSnackbar('Invalid rule ID', 'error');
      return;
    }
    
    console.log('Toggling rule:', {
      originalId: id,
      convertedId: ruleId,
      idType: typeof id,
      enabled: enabled
    });
    
    const response = await fetch(`${API_URL}/rules/${ruleId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled })
    });
    
    if (!response.ok) {
      const errorData = await response.json();
      console.error('Toggle rule error response:', errorData);
      throw new Error(errorData.error || 'Failed to update rule');
    }
    
    await loadRules();
    showSnackbar(enabled ? 'Rule enabled' : 'Rule disabled', 'success');
  } catch (error) {
    console.error('Error toggling rule:', error);
    showSnackbar('Failed to update rule', 'error');
  }
}

async function deleteRule(id: string) {
  if (!confirm('Are you sure you want to delete this rule?')) return;
  
  try {
    const ruleId = id?.toString() || '';
    if (!ruleId) {
      showSnackbar('Invalid rule ID', 'error');
      return;
    }
    
    console.log('Deleting rule:', {
      originalId: id,
      convertedId: ruleId,
      idType: typeof id
    });
    
    const response = await fetch(`${API_URL}/rules/${ruleId}`, { 
      method: 'DELETE' 
    });
    
    if (!response.ok) {
      const errorData = await response.json();
      console.error('Delete rule error response:', errorData);
      throw new Error(errorData.error || 'Failed to delete rule');
    }
    
    await loadRules();
    showSnackbar('Rule deleted successfully', 'success');
  } catch (error) {
    console.error('Error deleting rule:', error);
    showSnackbar('Failed to delete rule', 'error');
  }
}

async function loadAnalytics() {
  try {
    const response = await fetch(`${API_URL}/analytics/statistics?hours=${analyticsHours.value}`);
    if (response.ok) {
      statistics.value = await response.json();
      showSnackbar('Analytics data refreshed', 'success');
    }
  } catch (error) {
    console.error('Error loading analytics:', error);
    showSnackbar('Failed to load analytics', 'error');
  }
}

async function loadTrends() {
  try {
    trendsLoading.value = true;
    const response = await fetch(`${API_URL}/analytics/trends?hours=${analyticsHours.value}`);
    if (!response.ok) {
      throw new Error('Failed to load trends');
    }
    const data = await response.json();
    trendData.value = Array.isArray(data) ? data : [];
  } catch (error) {
    console.error('Error loading trends:', error);
    trendData.value = [];
    showSnackbar('Failed to load trend data', 'error');
  } finally {
    trendsLoading.value = false;
  }
}

async function loadAlerts() {
  try {
    const response = await fetch(`${API_URL}/alerts?limit=20`);
    if (response.ok) {
      recentAlerts.value = await response.json();
      updateAlertCounts();
    }
  } catch (error) {
    console.error('Error loading alerts:', error);
  }
}

function updateAlertCounts() {
  criticalAlerts.value = recentAlerts.value.filter(alert => 
    alert.alerts.some((msg: string) => msg.toLowerCase().includes('danger') || msg.toLowerCase().includes('critical'))
  ).length;
  
  warningAlerts.value = recentAlerts.value.filter(alert => 
    alert.alerts.some((msg: string) => msg.toLowerCase().includes('warning') || msg.toLowerCase().includes('alert'))
  ).length;
  
  infoAlerts.value = recentAlerts.value.length - criticalAlerts.value - warningAlerts.value;
}

async function exportData() {
  try {
    // This would typically generate and download a CSV/JSON file
    showSnackbar('Export feature coming soon', 'info');
  } catch (error) {
    showSnackbar('Export failed', 'error');
  }
}

function startDataRefresh() {
  refreshInterval = setInterval(() => {
    loadCurrentData();
    loadAlerts();
    if (isConnected.value) {
      loadActuatorStates();
    }
  }, 3000);
}

function stopDataRefresh() {
  if (refreshInterval) {
    clearInterval(refreshInterval);
  }
}

// Enhanced UI Helper Functions
function getSensorCardColor(sensor: string, value: number): string {
  if (!value) return 'grey-darken-1';
  
  const sensorConfig = sensorTypes.find(s => s.key === sensor);
  if (!sensorConfig) return 'grey-darken-1';
  
  if (sensor === 'gas' && value > 700) return 'error';
  if (sensor === 'water' && value > 800) return 'warning';
  if (sensor === 'soil' && value > 50) return 'warning';
  if (sensor === 'light' && value < 300) return 'info';
  
  return 'success';
}

function getSensorStatus(sensor: string, value: number): string {
  if (!value) return 'No Data';
  
  switch (sensor) {
    case 'light':
      return value < 300 ? 'Dark Environment' : value > 800 ? 'Very Bright' : 'Good Lighting';
    case 'gas':
      return value > 700 ? 'DANGER LEVEL!' : value > 400 ? 'Elevated' : 'Safe Level';
    case 'soil':
      return value > 70 ? 'Very Dry' : value > 50 ? 'Dry' : value > 30 ? 'Moist' : 'Wet';
    case 'water':
      return value > 800 ? 'Heavy Rain!' : value > 400 ? 'Rain Detected' : 'No Rain';
    default:
      return 'Unknown';
  }
}

function getSensorStatusColor(sensor: string, value: number): string {
  if (!value) return 'grey';
  
  switch (sensor) {
    case 'gas':
      return value > 700 ? 'error' : value > 400 ? 'warning' : 'success';
    case 'water':
      return value > 800 ? 'error' : value > 400 ? 'warning' : 'info';
    case 'soil':
      return value > 50 ? 'warning' : 'success';
    case 'light':
      return value < 300 ? 'info' : 'success';
    default:
      return 'grey';
  }
}

function getOverallStatus(): string {
  const gasLevel = currentData.value?.gas || 0;
  const waterLevel = currentData.value?.water || 0;
  
  if (gasLevel > 700) return 'Critical Alert';
  if (waterLevel > 800) return 'Weather Alert';
  if (!isConnected.value) return 'System Offline';
  
  return 'All Systems Normal';
}

function getOverallStatusIcon(): string {
  const status = getOverallStatus();
  if (status.includes('Critical')) return 'mdi-alert-octagon';
  if (status.includes('Alert')) return 'mdi-weather-rainy';
  if (status.includes('Offline')) return 'mdi-connection';
  return 'mdi-check-circle';
}

function getOverallStatusColor(): string {
  const status = getOverallStatus();
  if (status.includes('Critical')) return 'error';
  if (status.includes('Alert')) return 'warning';
  if (status.includes('Offline')) return 'grey';
  return 'success';
}

// Actuator helper functions
function getActuatorStatusColor(key: string, value: any): string {
  if (typeof value === 'boolean') {
    return value ? 'success' : 'grey';
  }
  if (typeof value === 'number') {
    return value > 0 ? 'success' : 'grey';
  }
  return 'grey';
}

function getActuatorIcon(key: string): string {
  const icons: Record<string, string> = {
    white_light: 'mdi-lightbulb',
    yellow_light: 'mdi-lightbulb-variant',
    relay: 'mdi-electric-switch',
    door_angle: 'mdi-door',
    window_angle: 'mdi-window-shutter',
    fan: 'mdi-fan',
    buzzer: 'mdi-volume-high'
  };
  return icons[key] || 'mdi-help-circle';
}

function formatActuatorName(key: string): string {
  return key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
}

function formatActuatorValue(key: string, value: any): string {
  if (typeof value === 'boolean') {
    return value ? 'ON' : 'OFF';
  }
  if (key.includes('angle')) {
    return `${value}°`;
  }
  if (key.includes('speed')) {
    return `${Math.round((value / 255) * 100)}%`;
  }
  return String(value);
}

// Alert helper functions
function getAlertColor(alert: any): string {
  const message = alert.alerts.join(' ').toLowerCase();
  if (message.includes('danger') || message.includes('critical')) return 'error';
  if (message.includes('warning') || message.includes('alert')) return 'warning';
  return 'info';
}

function getAlertIcon(alert: any): string {
  const message = alert.alerts.join(' ').toLowerCase();
  if (message.includes('danger')) return 'mdi-alert-octagon';
  if (message.includes('gas')) return 'mdi-gas-cylinder';
  if (message.includes('water') || message.includes('rain')) return 'mdi-water-alert';
  if (message.includes('light')) return 'mdi-lightbulb-alert';
  return 'mdi-information';
}

function getAlertClass(alert: any): string {
  const severity = getAlertSeverity(alert);
  return `alert-${severity}`;
}

function getAlertSeverity(alert: any): string {
  const message = alert.alerts.join(' ').toLowerCase();
  if (message.includes('danger') || message.includes('critical')) return 'error';
  if (message.includes('warning') || message.includes('alert')) return 'warning';
  return 'info';
}

function getAlertSeverityText(alert: any): string {
  const severity = getAlertSeverity(alert);
  return severity.charAt(0).toUpperCase() + severity.slice(1);
}

// Rule helper functions
function getSensorLabel(sensorKey: string): string {
  const sensor = sensorTypes.find(s => s.key === sensorKey);
  return sensor ? sensor.label : sensorKey;
}

function getActionLabel(actionValue: string): string {
  const action = actionOptions.find(a => a.value === actionValue);
  return action ? action.title : actionValue;
}

function getStatValue(sensor: string, stat: string): string {
  const key = `${sensor}_${stat}`;
  return statistics.value[key] ? statistics.value[key].toFixed(2) : 'N/A';
}

function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleString();
}

// Theme and UI functions
function toggleTheme() {
  theme.global.name.value = theme.global.current.value.dark ? 'light' : 'dark';
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen();
  } else {
    document.exitFullscreen();
  }
}

function showSnackbar(message: string, color: string) {
  snackbarText.value = message;
  snackbarColor.value = color;
  snackbar.value = true;
}

// Lifecycle
watch(analyticsHours, () => {
  loadAnalytics();
  loadTrends();
});

onMounted(() => {
  refreshPorts();
  loadRules();
  loadAnalytics();
  loadTrends();
  loadAlerts();
});

onUnmounted(() => {
  stopDataRefresh();
});
</script>

<style scoped>
.minimal-app-bar {
  box-shadow: none;
  border-bottom: 1px solid #eee;
}
.minimal-drawer {
  background: #fff;
  border-right: 1px solid #eee;
}
.nav-item-minimal {
  font-size: 1rem;
  color: #333;
  margin-bottom: 2px;
}
.minimal-main {
  background: #fafbfc;
}
.minimal-card, .minimal-sensor-card, .minimal-control-card, .minimal-status-card, .minimal-analytics-card, .minimal-rule-card, .minimal-alert-card {
  background: #fff;
  border-radius: 12px;
  box-shadow: none;
  border: 1px solid #eee;
}
.tab-item-minimal {
  font-size: 1rem;
  color: #333;
}
.minimal-window {
  background: transparent;
}
.minimal-fab {
  z-index: 10;
}
.status-chip {
  font-size: 0.9rem;
  padding: 0 8px;
}
</style>
