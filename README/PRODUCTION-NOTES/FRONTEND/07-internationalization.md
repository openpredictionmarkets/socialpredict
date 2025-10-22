# Internationalization (i18n) Implementation Plan

## Overview
Implement comprehensive internationalization support to make the application accessible to users worldwide, supporting multiple languages, locales, and cultural preferences.

## Current State Analysis
- English-only interface
- Hardcoded text strings throughout components
- No locale-specific formatting
- No RTL (Right-to-Left) language support
- No timezone handling
- No currency/number formatting by locale
- No date/time localization

## Implementation Steps

### Step 1: i18n Library Setup and Configuration
**Timeline: 2-3 days**

Set up React i18next for comprehensive internationalization:

```javascript
// package.json - Add dependencies
{
  "dependencies": {
    "react-i18next": "^13.5.0",
    "i18next": "^23.7.6",
    "i18next-browser-languagedetector": "^7.2.0",
    "i18next-http-backend": "^2.4.2",
    "i18next-icu": "^2.3.0"
  }
}

// src/i18n/index.js - Main i18n configuration
import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import Backend from 'i18next-http-backend'
import LanguageDetector from 'i18next-browser-languagedetector'
import ICU from 'i18next-icu'

// Import locale data for ICU
import en from 'i18next-icu/locale-data/en'
import es from 'i18next-icu/locale-data/es'
import fr from 'i18next-icu/locale-data/fr'
import de from 'i18next-icu/locale-data/de'
import ja from 'i18next-icu/locale-data/ja'
import zh from 'i18next-icu/locale-data/zh'

const localeData = { en, es, fr, de, ja, zh }

i18n
  .use(Backend)
  .use(LanguageDetector)
  .use(ICU)
  .use(initReactI18next)
  .init({
    // Language detection
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'i18nextLng',
      caches: ['localStorage'],
    },

    // Fallback language
    fallbackLng: 'en',
    
    // Debug mode for development
    debug: process.env.NODE_ENV === 'development',

    // Interpolation options
    interpolation: {
      escapeValue: false, // React already escapes
    },

    // Backend configuration
    backend: {
      loadPath: '/locales/{{lng}}/{{ns}}.json',
      addPath: '/locales/add/{{lng}}/{{ns}}',
    },

    // Default namespace
    defaultNS: 'common',
    
    // Available namespaces
    ns: ['common', 'auth', 'markets', 'navigation', 'errors', 'validation'],

    // ICU configuration
    i18nFormat: {
      localeData,
      formats: {
        number: {
          currency: {
            style: 'currency',
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
          },
          percentage: {
            style: 'percent',
            minimumFractionDigits: 1,
            maximumFractionDigits: 1,
          },
        },
        date: {
          short: {
            day: 'numeric',
            month: 'short',
            year: 'numeric',
          },
          long: {
            weekday: 'long',
            day: 'numeric',
            month: 'long',
            year: 'numeric',
          },
        },
        time: {
          short: {
            hour: 'numeric',
            minute: 'numeric',
          },
          long: {
            hour: 'numeric',
            minute: 'numeric',
            second: 'numeric',
            timeZoneName: 'short',
          },
        },
      },
    },

    // React options
    react: {
      useSuspense: true,
      bindI18n: 'languageChanged',
      bindI18nStore: '',
      transEmptyNodeValue: '',
      transSupportBasicHtmlNodes: true,
      transKeepBasicHtmlNodesFor: ['br', 'strong', 'i', 'em'],
      transWrapTextNodes: '',
    },
  })

export default i18n

// src/i18n/resources.js - Type-safe translation keys
export const translationKeys = {
  common: {
    loading: 'loading',
    error: 'error',
    success: 'success',
    cancel: 'cancel',
    save: 'save',
    delete: 'delete',
    edit: 'edit',
    confirm: 'confirm',
    yes: 'yes',
    no: 'no',
    back: 'back',
    next: 'next',
    previous: 'previous',
    close: 'close',
    search: 'search',
    filter: 'filter',
    sort: 'sort',
    selectAll: 'selectAll',
    clearAll: 'clearAll',
  },
  auth: {
    login: 'login',
    logout: 'logout',
    register: 'register',
    username: 'username',
    password: 'password',
    email: 'email',
    forgotPassword: 'forgotPassword',
    resetPassword: 'resetPassword',
    loginSuccess: 'loginSuccess',
    loginError: 'loginError',
    invalidCredentials: 'invalidCredentials',
  },
  markets: {
    title: 'title',
    description: 'description',
    status: 'status',
    volume: 'volume',
    participants: 'participants',
    odds: 'odds',
    placeBet: 'placeBet',
    betAmount: 'betAmount',
    timeRemaining: 'timeRemaining',
    marketClosed: 'marketClosed',
    betPlaced: 'betPlaced',
    createMarket: 'createMarket',
  },
  navigation: {
    home: 'home',
    markets: 'markets',
    profile: 'profile',
    admin: 'admin',
    settings: 'settings',
    help: 'help',
    about: 'about',
  },
  errors: {
    networkError: 'networkError',
    serverError: 'serverError',
    notFound: 'notFound',
    unauthorized: 'unauthorized',
    forbidden: 'forbidden',
    validationError: 'validationError',
    tryAgain: 'tryAgain',
  },
  validation: {
    required: 'required',
    invalidEmail: 'invalidEmail',
    passwordTooShort: 'passwordTooShort',
    passwordMismatch: 'passwordMismatch',
    invalidAmount: 'invalidAmount',
    minimumAmount: 'minimumAmount',
    maximumAmount: 'maximumAmount',
  },
} as const
```

Create translation files for multiple languages:

```json
// public/locales/en/common.json
{
  "loading": "Loading...",
  "error": "Error",
  "success": "Success",
  "cancel": "Cancel",
  "save": "Save",
  "delete": "Delete",
  "edit": "Edit",
  "confirm": "Confirm",
  "yes": "Yes",
  "no": "No",
  "back": "Back",
  "next": "Next",
  "previous": "Previous",
  "close": "Close",
  "search": "Search",
  "filter": "Filter",
  "sort": "Sort",
  "selectAll": "Select All",
  "clearAll": "Clear All",
  "welcome": "Welcome to SocialPredict",
  "welcomeMessage": "Join thousands of users making predictions on future events.",
  "getStarted": "Get Started",
  "learnMore": "Learn More"
}

// public/locales/en/markets.json
{
  "title": "Market Title",
  "description": "Description",
  "status": "Status",
  "volume": "Volume",
  "participants": "Participants",
  "odds": "Odds",
  "placeBet": "Place Bet",
  "betAmount": "Bet Amount",
  "timeRemaining": "Time Remaining",
  "marketClosed": "Market Closed",
  "betPlaced": "Bet placed successfully!",
  "createMarket": "Create Market",
  "marketsList": "Markets List",
  "noMarkets": "No markets available",
  "loadingMarkets": "Loading markets...",
  "marketDetails": "Market Details",
  "outcomeYes": "Yes",
  "outcomeNo": "No",
  "totalVolume": "Total Volume: {amount, number, currency}",
  "participantCount": "{count, plural, =0 {No participants} =1 {1 participant} other {# participants}}",
  "timeLeft": "{duration} remaining",
  "betConfirmation": "Are you sure you want to bet {amount, number, currency} on {outcome}?",
  "minimumBet": "Minimum bet: {amount, number, currency}",
  "maximumBet": "Maximum bet: {amount, number, currency}",
  "insufficientBalance": "Insufficient balance. You have {balance, number, currency} available."
}

// public/locales/es/common.json
{
  "loading": "Cargando...",
  "error": "Error",
  "success": "Éxito",
  "cancel": "Cancelar",
  "save": "Guardar",
  "delete": "Eliminar",
  "edit": "Editar",
  "confirm": "Confirmar",
  "yes": "Sí",
  "no": "No",
  "back": "Atrás",
  "next": "Siguiente",
  "previous": "Anterior",
  "close": "Cerrar",
  "search": "Buscar",
  "filter": "Filtrar",
  "sort": "Ordenar",
  "selectAll": "Seleccionar Todo",
  "clearAll": "Limpiar Todo",
  "welcome": "Bienvenido a SocialPredict",
  "welcomeMessage": "Únete a miles de usuarios haciendo predicciones sobre eventos futuros.",
  "getStarted": "Comenzar",
  "learnMore": "Aprender Más"
}

// public/locales/es/markets.json
{
  "title": "Título del Mercado",
  "description": "Descripción",
  "status": "Estado",
  "volume": "Volumen",
  "participants": "Participantes",
  "odds": "Probabilidades",
  "placeBet": "Hacer Apuesta",
  "betAmount": "Cantidad de Apuesta",
  "timeRemaining": "Tiempo Restante",
  "marketClosed": "Mercado Cerrado",
  "betPlaced": "¡Apuesta realizada con éxito!",
  "createMarket": "Crear Mercado",
  "marketsList": "Lista de Mercados",
  "noMarkets": "No hay mercados disponibles",
  "loadingMarkets": "Cargando mercados...",
  "marketDetails": "Detalles del Mercado",
  "outcomeYes": "Sí",
  "outcomeNo": "No",
  "totalVolume": "Volumen Total: {amount, number, currency}",
  "participantCount": "{count, plural, =0 {Sin participantes} =1 {1 participante} other {# participantes}}",
  "timeLeft": "{duration} restante",
  "betConfirmation": "¿Estás seguro de que quieres apostar {amount, number, currency} en {outcome}?",
  "minimumBet": "Apuesta mínima: {amount, number, currency}",
  "maximumBet": "Apuesta máxima: {amount, number, currency}",
  "insufficientBalance": "Saldo insuficiente. Tienes {balance, number, currency} disponible."
}
```

### Step 2: Translation Hooks and Components
**Timeline: 2 days**

Create reusable translation utilities:

```javascript
// hooks/useTranslation.js - Enhanced translation hook
import { useTranslation as useI18nextTranslation } from 'react-i18next'
import { useMemo } from 'react'

export const useTranslation = (namespace = 'common') => {
  const { t, i18n, ready } = useI18nextTranslation(namespace)

  const translationHelpers = useMemo(() => ({
    // Format currency based on current locale
    formatCurrency: (amount, currency = 'USD') => {
      return new Intl.NumberFormat(i18n.language, {
        style: 'currency',
        currency,
      }).format(amount)
    },

    // Format number based on current locale
    formatNumber: (number, options = {}) => {
      return new Intl.NumberFormat(i18n.language, options).format(number)
    },

    // Format date based on current locale
    formatDate: (date, options = {}) => {
      return new Intl.DateTimeFormat(i18n.language, options).format(new Date(date))
    },

    // Format relative time (e.g., "2 hours ago")
    formatRelativeTime: (date) => {
      const rtf = new Intl.RelativeTimeFormat(i18n.language, { numeric: 'auto' })
      const now = new Date()
      const targetDate = new Date(date)
      const diffInSeconds = (targetDate - now) / 1000

      if (Math.abs(diffInSeconds) < 60) {
        return rtf.format(Math.round(diffInSeconds), 'second')
      } else if (Math.abs(diffInSeconds) < 3600) {
        return rtf.format(Math.round(diffInSeconds / 60), 'minute')
      } else if (Math.abs(diffInSeconds) < 86400) {
        return rtf.format(Math.round(diffInSeconds / 3600), 'hour')
      } else {
        return rtf.format(Math.round(diffInSeconds / 86400), 'day')
      }
    },

    // Get pluralized text with count
    pluralize: (key, count, options = {}) => {
      return t(key, { count, ...options })
    },

    // Check if current language is RTL
    isRTL: () => {
      const rtlLanguages = ['ar', 'he', 'fa', 'ur']
      return rtlLanguages.includes(i18n.language)
    },

    // Get current language info
    getCurrentLanguage: () => ({
      code: i18n.language,
      name: t('languageName', { lng: i18n.language }),
      direction: rtlLanguages.includes(i18n.language) ? 'rtl' : 'ltr',
    }),
  }), [t, i18n])

  return {
    t,
    i18n,
    ready,
    ...translationHelpers,
  }
}

// components/i18n/TranslatedText.jsx
import React from 'react'
import { useTranslation } from '../../hooks/useTranslation'

const TranslatedText = ({ 
  tKey, 
  namespace = 'common',
  values = {},
  components = {},
  fallback = '',
  className = '',
  as = 'span',
}) => {
  const { t } = useTranslation(namespace)
  const Component = as

  const translatedText = t(tKey, { 
    ...values,
    ...components,
    defaultValue: fallback,
  })

  return (
    <Component className={className}>
      {translatedText}
    </Component>
  )
}

export default TranslatedText

// components/i18n/LanguageSwitcher.jsx
import React, { useState } from 'react'
import { useTranslation } from '../../hooks/useTranslation'
import { ChevronDownIcon, GlobeAltIcon } from '@heroicons/react/24/outline'

const availableLanguages = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'es', name: 'Español', flag: '🇪🇸' },
  { code: 'fr', name: 'Français', flag: '🇫🇷' },
  { code: 'de', name: 'Deutsch', flag: '🇩🇪' },
  { code: 'ja', name: '日本語', flag: '🇯🇵' },
  { code: 'zh', name: '中文', flag: '🇨🇳' },
]

const LanguageSwitcher = ({ className = '' }) => {
  const { i18n, t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)

  const currentLanguage = availableLanguages.find(lang => lang.code === i18n.language) 
    || availableLanguages[0]

  const handleLanguageChange = async (languageCode) => {
    await i18n.changeLanguage(languageCode)
    setIsOpen(false)
    
    // Update document direction for RTL languages
    const rtlLanguages = ['ar', 'he', 'fa', 'ur']
    document.dir = rtlLanguages.includes(languageCode) ? 'rtl' : 'ltr'
    
    // Update HTML lang attribute
    document.documentElement.lang = languageCode
  }

  return (
    <div className={`relative ${className}`}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center space-x-2 px-3 py-2 text-sm bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        aria-label={t('selectLanguage')}
        aria-expanded={isOpen}
        aria-haspopup="listbox"
      >
        <GlobeAltIcon className="h-4 w-4" />
        <span>{currentLanguage.flag}</span>
        <span>{currentLanguage.name}</span>
        <ChevronDownIcon className="h-4 w-4" />
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-48 bg-white border border-gray-200 rounded-md shadow-lg z-50">
          <div className="py-1" role="listbox">
            {availableLanguages.map((language) => (
              <button
                key={language.code}
                onClick={() => handleLanguageChange(language.code)}
                className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-100 flex items-center space-x-3 ${
                  language.code === i18n.language ? 'bg-indigo-50 text-indigo-900' : 'text-gray-900'
                }`}
                role="option"
                aria-selected={language.code === i18n.language}
              >
                <span>{language.flag}</span>
                <span>{language.name}</span>
                {language.code === i18n.language && (
                  <span className="ml-auto text-indigo-600">✓</span>
                )}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export default LanguageSwitcher
```

### Step 3: Localized Components
**Timeline: 3-4 days**

Create localized versions of key components:

```javascript
// components/markets/LocalizedMarketCard.jsx
import React from 'react'
import { useTranslation } from '../../hooks/useTranslation'
import TranslatedText from '../i18n/TranslatedText'

const LocalizedMarketCard = ({ market, onBetClick }) => {
  const { t, formatCurrency, formatRelativeTime, pluralize, isRTL } = useTranslation('markets')

  const handleBetClick = () => {
    onBetClick(market.id)
  }

  const timeRemaining = formatRelativeTime(market.closingDate)
  const direction = isRTL() ? 'rtl' : 'ltr'

  return (
    <div 
      className={`market-card border rounded-lg p-4 hover:shadow-lg ${
        direction === 'rtl' ? 'text-right' : 'text-left'
      }`}
      dir={direction}
    >
      <header className="mb-4">
        <h3 className="text-xl font-semibold mb-2">
          {market.title}
        </h3>
        
        <div className="flex items-center mb-2">
          <span 
            className={`inline-block px-2 py-1 text-xs rounded ${
              market.status === 'open' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
            }`}
          >
            <TranslatedText 
              tKey={`status.${market.status}`} 
              fallback={market.status} 
            />
          </span>
        </div>
      </header>

      <div className="text-gray-600 mb-4">
        {market.description}
      </div>

      <div className="grid grid-cols-2 gap-4 text-sm mb-4">
        <div>
          <dt className="font-medium">
            <TranslatedText tKey="volume" />:
          </dt>
          <dd>{formatCurrency(market.totalVolume)}</dd>
        </div>
        
        <div>
          <dt className="font-medium">
            <TranslatedText tKey="participants" />:
          </dt>
          <dd>
            {pluralize('participantCount', market.participantCount, {
              count: market.participantCount
            })}
          </dd>
        </div>
        
        <div>
          <dt className="font-medium">
            <TranslatedText tKey="odds" />:
          </dt>
          <dd>{market.outcomes[0]?.odds || t('common:notAvailable')}</dd>
        </div>
        
        <div>
          <dt className="font-medium">
            <TranslatedText tKey="timeRemaining" />:
          </dt>
          <dd>{timeRemaining}</dd>
        </div>
      </div>

      <button
        onClick={handleBetClick}
        disabled={market.status !== 'open'}
        className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-400 disabled:bg-gray-400 disabled:cursor-not-allowed"
      >
        {market.status === 'open' ? (
          <TranslatedText tKey="placeBet" />
        ) : (
          <TranslatedText tKey="marketClosed" />
        )}
      </button>
    </div>
  )
}

export default LocalizedMarketCard

// components/forms/LocalizedForm.jsx
import React from 'react'
import { useTranslation } from '../../hooks/useTranslation'
import TranslatedText from '../i18n/TranslatedText'

const LocalizedForm = ({ 
  children, 
  title, 
  onSubmit, 
  submitText = 'submit',
  cancelText = 'cancel',
  onCancel,
  loading = false,
}) => {
  const { t, isRTL } = useTranslation('common')
  const direction = isRTL() ? 'rtl' : 'ltr'

  return (
    <form 
      onSubmit={onSubmit}
      className="space-y-6"
      dir={direction}
    >
      {title && (
        <h2 className="text-2xl font-bold text-gray-900">
          <TranslatedText tKey={title} />
        </h2>
      )}
      
      <div className="space-y-4">
        {children}
      </div>
      
      <div className={`flex ${direction === 'rtl' ? 'space-x-reverse' : ''} space-x-4`}>
        <button
          type="submit"
          disabled={loading}
          className="flex-1 bg-indigo-600 text-white py-2 px-4 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50"
        >
          {loading ? (
            <TranslatedText tKey="loading" />
          ) : (
            <TranslatedText tKey={submitText} />
          )}
        </button>
        
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className="flex-1 bg-gray-300 text-gray-700 py-2 px-4 rounded-md hover:bg-gray-400 focus:outline-none focus:ring-2 focus:ring-gray-500"
          >
            <TranslatedText tKey={cancelText} />
          </button>
        )}
      </div>
    </form>
  )
}

export default LocalizedForm

// components/forms/LocalizedValidation.jsx
import React from 'react'
import { useTranslation } from '../../hooks/useTranslation'

const LocalizedValidation = ({ errors = [], field }) => {
  const { t } = useTranslation('validation')

  if (!errors.length) return null

  return (
    <div className="mt-1 text-sm text-red-600" role="alert">
      {errors.map((error, index) => (
        <div key={index}>
          {typeof error === 'string' ? error : t(error.key, error.values)}
        </div>
      ))}
    </div>
  )
}

export default LocalizedValidation
```

### Step 4: Date, Time, and Number Formatting
**Timeline: 2 days**

Implement comprehensive locale-aware formatting:

```javascript
// utils/formatters.js
export class LocaleFormatters {
  constructor(locale = 'en-US', currency = 'USD', timezone = 'UTC') {
    this.locale = locale
    this.currency = currency
    this.timezone = timezone
  }

  updateLocale(locale, currency, timezone) {
    this.locale = locale
    this.currency = currency || this.currency
    this.timezone = timezone || this.timezone
  }

  // Currency formatting
  formatCurrency(amount, options = {}) {
    const { currency = this.currency, ...otherOptions } = options
    
    return new Intl.NumberFormat(this.locale, {
      style: 'currency',
      currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
      ...otherOptions,
    }).format(amount)
  }

  // Number formatting
  formatNumber(number, options = {}) {
    return new Intl.NumberFormat(this.locale, {
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
      ...options,
    }).format(number)
  }

  // Percentage formatting
  formatPercentage(number, options = {}) {
    return new Intl.NumberFormat(this.locale, {
      style: 'percent',
      minimumFractionDigits: 1,
      maximumFractionDigits: 1,
      ...options,
    }).format(number / 100)
  }

  // Date formatting
  formatDate(date, options = {}) {
    const defaultOptions = {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      timeZone: this.timezone,
    }

    return new Intl.DateTimeFormat(this.locale, {
      ...defaultOptions,
      ...options,
    }).format(new Date(date))
  }

  // Time formatting
  formatTime(date, options = {}) {
    const defaultOptions = {
      hour: 'numeric',
      minute: '2-digit',
      timeZone: this.timezone,
    }

    return new Intl.DateTimeFormat(this.locale, {
      ...defaultOptions,
      ...options,
    }).format(new Date(date))
  }

  // DateTime formatting
  formatDateTime(date, options = {}) {
    const defaultOptions = {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      timeZone: this.timezone,
    }

    return new Intl.DateTimeFormat(this.locale, {
      ...defaultOptions,
      ...options,
    }).format(new Date(date))
  }

  // Relative time formatting
  formatRelativeTime(date, options = {}) {
    const rtf = new Intl.RelativeTimeFormat(this.locale, {
      numeric: 'auto',
      ...options,
    })

    const now = new Date()
    const targetDate = new Date(date)
    const diffInSeconds = (targetDate - now) / 1000

    const intervals = [
      { unit: 'year', seconds: 31536000 },
      { unit: 'month', seconds: 2628000 },
      { unit: 'week', seconds: 604800 },
      { unit: 'day', seconds: 86400 },
      { unit: 'hour', seconds: 3600 },
      { unit: 'minute', seconds: 60 },
      { unit: 'second', seconds: 1 },
    ]

    for (const interval of intervals) {
      const count = Math.floor(Math.abs(diffInSeconds) / interval.seconds)
      if (count >= 1) {
        return rtf.format(
          diffInSeconds < 0 ? -count : count,
          interval.unit
        )
      }
    }

    return rtf.format(0, 'second')
  }

  // List formatting
  formatList(items, options = {}) {
    const lf = new Intl.ListFormat(this.locale, {
      style: 'long',
      type: 'conjunction',
      ...options,
    })

    return lf.format(items)
  }

  // Display names
  getDisplayName(code, type = 'language', options = {}) {
    const dn = new Intl.DisplayNames(this.locale, {
      type,
      ...options,
    })

    return dn.of(code)
  }
}

// hooks/useLocaleFormatters.js
import { useMemo } from 'react'
import { useTranslation } from './useTranslation'
import { LocaleFormatters } from '../utils/formatters'

export const useLocaleFormatters = () => {
  const { i18n } = useTranslation()

  const formatters = useMemo(() => {
    // Get user's currency preference from local storage or default
    const userCurrency = localStorage.getItem('userCurrency') || 'USD'
    const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone

    return new LocaleFormatters(i18n.language, userCurrency, userTimezone)
  }, [i18n.language])

  // Update formatters when language changes
  useMemo(() => {
    const userCurrency = localStorage.getItem('userCurrency') || 'USD'
    const userTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone
    formatters.updateLocale(i18n.language, userCurrency, userTimezone)
  }, [i18n.language, formatters])

  return formatters
}

// components/common/LocalizedDate.jsx
import React from 'react'
import { useLocaleFormatters } from '../../hooks/useLocaleFormatters'

const LocalizedDate = ({ 
  date, 
  format = 'date',
  options = {},
  relative = false,
  className = '',
}) => {
  const formatters = useLocaleFormatters()

  const getFormattedDate = () => {
    if (relative) {
      return formatters.formatRelativeTime(date, options)
    }

    switch (format) {
      case 'date':
        return formatters.formatDate(date, options)
      case 'time':
        return formatters.formatTime(date, options)
      case 'datetime':
        return formatters.formatDateTime(date, options)
      default:
        return formatters.formatDate(date, options)
    }
  }

  return (
    <time 
      dateTime={new Date(date).toISOString()} 
      className={className}
      title={formatters.formatDateTime(date)}
    >
      {getFormattedDate()}
    </time>
  )
}

export default LocalizedDate

// components/common/LocalizedNumber.jsx
import React from 'react'
import { useLocaleFormatters } from '../../hooks/useLocaleFormatters'

const LocalizedNumber = ({ 
  value, 
  type = 'number',
  currency,
  options = {},
  className = '',
}) => {
  const formatters = useLocaleFormatters()

  const getFormattedNumber = () => {
    switch (type) {
      case 'currency':
        return formatters.formatCurrency(value, { currency, ...options })
      case 'percentage':
        return formatters.formatPercentage(value, options)
      case 'number':
      default:
        return formatters.formatNumber(value, options)
    }
  }

  return (
    <span className={className}>
      {getFormattedNumber()}
    </span>
  )
}

export default LocalizedNumber
```

### Step 5: RTL Language Support
**Timeline: 2 days**

Implement comprehensive RTL (Right-to-Left) language support:

```javascript
// utils/rtlSupport.js
export const rtlLanguages = ['ar', 'he', 'fa', 'ur', 'dv', 'ps']

export const isRTLLanguage = (language) => {
  return rtlLanguages.some(rtl => language.startsWith(rtl))
}

export const getRTLStyles = (isRTL) => ({
  direction: isRTL ? 'rtl' : 'ltr',
  textAlign: isRTL ? 'right' : 'left',
})

// components/layout/RTLProvider.jsx
import React, { useEffect } from 'react'
import { useTranslation } from '../../hooks/useTranslation'
import { isRTLLanguage } from '../../utils/rtlSupport'

const RTLProvider = ({ children }) => {
  const { i18n } = useTranslation()
  const isRTL = isRTLLanguage(i18n.language)

  useEffect(() => {
    // Update document direction
    document.dir = isRTL ? 'rtl' : 'ltr'
    document.documentElement.setAttribute('dir', isRTL ? 'rtl' : 'ltr')
    
    // Update body class for CSS targeting
    document.body.classList.toggle('rtl', isRTL)
    document.body.classList.toggle('ltr', !isRTL)
    
    // Update HTML lang attribute
    document.documentElement.lang = i18n.language
  }, [isRTL, i18n.language])

  return (
    <div dir={isRTL ? 'rtl' : 'ltr'} className={isRTL ? 'rtl' : 'ltr'}>
      {children}
    </div>
  )
}

export default RTLProvider

// styles/rtl.css - RTL-specific styles
/* RTL-specific utility classes */
.rtl .text-left {
  text-align: right;
}

.rtl .text-right {
  text-align: left;
}

.rtl .float-left {
  float: right;
}

.rtl .float-right {
  float: left;
}

.rtl .ml-2 {
  margin-left: 0;
  margin-right: 0.5rem;
}

.rtl .mr-2 {
  margin-right: 0;
  margin-left: 0.5rem;
}

.rtl .pl-4 {
  padding-left: 0;
  padding-right: 1rem;
}

.rtl .pr-4 {
  padding-right: 0;
  padding-left: 1rem;
}

/* Flexbox RTL support */
.rtl .flex-row {
  flex-direction: row-reverse;
}

.rtl .space-x-4 > :not([hidden]) ~ :not([hidden]) {
  --tw-space-x-reverse: 1;
  margin-right: calc(1rem * var(--tw-space-x-reverse));
  margin-left: calc(1rem * calc(1 - var(--tw-space-x-reverse)));
}

/* Form elements RTL */
.rtl input,
.rtl textarea,
.rtl select {
  text-align: right;
}

/* Navigation RTL */
.rtl .breadcrumb-separator::before {
  content: '\\';
  transform: scaleX(-1);
}

/* Icons that should flip in RTL */
.rtl .flip-rtl {
  transform: scaleX(-1);
}

/* tailwind.config.js - RTL plugin configuration
const plugin = require('tailwindcss/plugin')

module.exports = {
  plugins: [
    plugin(function({ addUtilities, theme, variants }) {
      const rtlUtilities = {
        '.rtl-flip': {
          transform: 'scaleX(-1)',
        },
        '.rtl-space-x-reverse > :not([hidden]) ~ :not([hidden])': {
          '--tw-space-x-reverse': '1',
        },
      }
      
      addUtilities(rtlUtilities, variants('space'))
    })
  ],
}
*/
```

## Directory Structure
```
src/
├── i18n/
│   ├── index.js              # Main i18n configuration
│   ├── resources.js          # Translation key constants
│   └── namespaces.js         # Namespace definitions
├── components/
│   ├── i18n/
│   │   ├── LanguageSwitcher.jsx
│   │   ├── TranslatedText.jsx
│   │   └── RTLProvider.jsx
│   ├── common/
│   │   ├── LocalizedDate.jsx
│   │   ├── LocalizedNumber.jsx
│   │   └── LocalizedValidation.jsx
│   ├── forms/
│   │   └── LocalizedForm.jsx
│   └── markets/
│       └── LocalizedMarketCard.jsx
├── hooks/
│   ├── useTranslation.js     # Enhanced translation hook
│   └── useLocaleFormatters.js
├── utils/
│   ├── formatters.js         # Locale formatters
│   └── rtlSupport.js         # RTL utilities
├── styles/
│   └── rtl.css              # RTL-specific styles
└── locales/                 # Translation files
    ├── en/
    │   ├── common.json
    │   ├── auth.json
    │   ├── markets.json
    │   ├── navigation.json
    │   ├── errors.json
    │   └── validation.json
    ├── es/
    ├── fr/
    ├── de/
    ├── ja/
    └── zh/
```

## Benefits
- Support for multiple languages and locales
- Proper currency, date, and number formatting
- RTL language support
- Type-safe translation keys
- Automatic language detection
- Locale-aware validation messages
- Cultural adaptation (colors, icons, layouts)
- SEO benefits for international markets
- Better user experience for global audience
- Compliance with accessibility standards

## Supported Languages (Initial)
- English (en)
- Spanish (es)
- French (fr)
- German (de)
- Japanese (ja)
- Chinese (zh)

## Features Implemented
- ✅ React i18next integration
- ✅ Multiple language support
- ✅ Locale-aware formatting
- ✅ RTL language support
- ✅ Currency localization
- ✅ Date/time localization
- ✅ Number formatting
- ✅ Pluralization support
- ✅ Language switcher component
- ✅ Translation validation
- ✅ Namespace organization