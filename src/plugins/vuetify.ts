import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { mdi } from 'vuetify/iconsets/mdi'
import '@mdi/font/css/materialdesignicons.css'

export default createVuetify({
  components,
  directives,
  icons: {
    defaultSet: 'mdi',
    sets: {
      mdi,
    },
  },
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary: '#667eea',
          secondary: '#764ba2',
          accent: '#00d4ff',
          error: '#ef4444',
          info: '#3b82f6',
          success: '#10b981',
          warning: '#f59e0b',
          background: '#f8fafc',
          surface: '#ffffff',
          'surface-variant': '#f1f5f9',
          'on-surface': '#1e293b',
          'on-primary': '#ffffff',
          'on-secondary': '#ffffff',
          'primary-darken-1': '#5a67d8',
          'secondary-darken-1': '#6b46c1',
        },
        variables: {
          'border-color': '#e2e8f0',
          'border-opacity': 0.12,
          'high-emphasis-opacity': 0.87,
          'medium-emphasis-opacity': 0.60,
          'disabled-opacity': 0.38,
          'activated-opacity': 0.12,
          'hover-opacity': 0.04,
          'focus-opacity': 0.12,
          'selected-opacity': 0.08,
          'pressed-opacity': 0.12,
          'dragged-opacity': 0.08,
          'kbd-background-color': '#212529',
          'kbd-color': '#ffffff',
          'code-background-color': '#f5f5f5',
        }
      },
      dark: {
        colors: {
          primary: '#667eea',
          secondary: '#764ba2',
          accent: '#00d4ff',
          error: '#f87171',
          info: '#60a5fa',
          success: '#34d399',
          warning: '#fbbf24',
          background: '#0f172a',
          surface: '#1e293b',
          'surface-variant': '#334155',
          'on-surface': '#f1f5f9',
          'on-primary': '#ffffff',
          'on-secondary': '#ffffff',
          'primary-darken-1': '#5a67d8',
          'secondary-darken-1': '#6b46c1',
        },
        variables: {
          'border-color': '#475569',
          'border-opacity': 0.12,
          'high-emphasis-opacity': 1,
          'medium-emphasis-opacity': 0.70,
          'disabled-opacity': 0.50,
          'activated-opacity': 0.12,
          'hover-opacity': 0.04,
          'focus-opacity': 0.12,
          'selected-opacity': 0.08,
          'pressed-opacity': 0.12,
          'dragged-opacity': 0.08,
          'kbd-background-color': '#212529',
          'kbd-color': '#ffffff',
          'code-background-color': '#2d3748',
        }
      },
    },
  },
  defaults: {
    VCard: {
      elevation: 2,
      rounded: 'lg',
    },
    VBtn: {
      style: 'text-transform: none; font-weight: 600;',
      rounded: 'lg',
    },
    VChip: {
      rounded: 'lg',
    },
    VAlert: {
      rounded: 'lg',
    },
    VTextField: {
      variant: 'outlined',
      density: 'comfortable',
    },
    VSelect: {
      variant: 'outlined',
      density: 'comfortable',
    },
    VTextarea: {
      variant: 'outlined',
      density: 'comfortable',
    },
    VTab: {
      style: 'text-transform: none; font-weight: 600;',
    },
  },
})
