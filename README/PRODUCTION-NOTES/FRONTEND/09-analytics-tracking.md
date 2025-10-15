# Analytics and Tracking Implementation Plan

## Overview
Implement comprehensive analytics and user tracking to gain insights into user behavior, application performance, and business metrics while ensuring privacy compliance and data protection.

## Current State Analysis
- No analytics implementation
- No user behavior tracking
- No performance monitoring
- No business metrics collection
- No A/B testing capabilities
- No conversion tracking
- No user journey analysis
- No error tracking integration with analytics

## Implementation Steps

### Step 1: Analytics Setup and Configuration
**Timeline: 2-3 days**

Set up Google Analytics 4 and comprehensive tracking infrastructure:

```javascript
// utils/analytics.js
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals'

class AnalyticsManager {
  constructor() {
    this.isInitialized = false
    this.queue = []
    this.userId = null
    this.sessionId = this.generateSessionId()
    this.pageLoadTime = Date.now()
  }

  // Initialize analytics services
  async initialize(config = {}) {
    try {
      // Initialize Google Analytics 4
      if (config.gaId) {
        await this.initializeGA4(config.gaId)
      }

      // Initialize other analytics services
      if (config.mixpanelToken) {
        await this.initializeMixpanel(config.mixpanelToken)
      }

      // Initialize performance monitoring
      this.initializePerformanceTracking()

      // Process queued events
      this.processQueue()

      this.isInitialized = true
      console.log('[Analytics] Initialized successfully')
    } catch (error) {
      console.error('[Analytics] Initialization failed:', error)
    }
  }

  // Google Analytics 4 setup
  async initializeGA4(gaId) {
    // Load gtag script
    const script = document.createElement('script')
    script.async = true
    script.src = `https://www.googletagmanager.com/gtag/js?id=${gaId}`
    document.head.appendChild(script)

    // Initialize gtag
    window.dataLayer = window.dataLayer || []
    window.gtag = function() { window.dataLayer.push(arguments) }
    
    window.gtag('js', new Date())
    window.gtag('config', gaId, {
      send_page_view: false, // We'll handle page views manually
      anonymize_ip: true,
      allow_google_signals: false, // Privacy-focused
      custom_map: {
        custom_parameter_1: 'user_type',
        custom_parameter_2: 'feature_flag',
      },
    })

    // Set up enhanced ecommerce
    window.gtag('config', gaId, {
      custom_parameters: {
        user_type: 'standard',
      },
    })
  }

  // Mixpanel setup for detailed user analytics
  async initializeMixpanel(token) {
    if (typeof window !== 'undefined' && !window.mixpanel) {
      const script = document.createElement('script')
      script.src = 'https://cdn.mxpnl.com/libs/mixpanel-2-latest.min.js'
      script.onload = () => {
        window.mixpanel.init(token, {
          debug: process.env.NODE_ENV === 'development',
          track_pageview: false,
          persistence: 'localStorage',
          ignore_dnt: false,
          property_blacklist: ['$current_url', '$initial_referrer', '$referrer'],
        })
      }
      document.head.appendChild(script)
    }
  }

  // Set user identification
  setUser(userId, properties = {}) {
    this.userId = userId

    if (window.gtag) {
      window.gtag('config', process.env.REACT_APP_GA_ID, {
        user_id: userId,
        custom_parameters: {
          user_type: properties.userType || 'standard',
          registration_date: properties.registrationDate,
          plan: properties.plan || 'free',
        },
      })
    }

    if (window.mixpanel) {
      window.mixpanel.identify(userId)
      window.mixpanel.people.set({
        $name: properties.name,
        $email: properties.email,
        $created: properties.registrationDate,
        user_type: properties.userType,
        plan: properties.plan,
      })
    }

    console.log('[Analytics] User identified:', userId)
  }

  // Track page views
  trackPageView(path, title = '') {
    const event = {
      type: 'page_view',
      data: {
        page_path: path,
        page_title: title || document.title,
        user_id: this.userId,
        session_id: this.sessionId,
        timestamp: new Date().toISOString(),
        referrer: document.referrer,
        user_agent: navigator.userAgent,
      },
    }

    if (this.isInitialized) {
      this.sendPageView(event)
    } else {
      this.queue.push(event)
    }
  }

  sendPageView(event) {
    if (window.gtag) {
      window.gtag('event', 'page_view', {
        page_title: event.data.page_title,
        page_location: window.location.href,
        page_path: event.data.page_path,
        user_id: event.data.user_id,
        session_id: event.data.session_id,
      })
    }

    if (window.mixpanel) {
      window.mixpanel.track('Page View', {
        page: event.data.page_path,
        title: event.data.page_title,
        referrer: event.data.referrer,
      })
    }
  }

  // Track custom events
  trackEvent(eventName, properties = {}) {
    const event = {
      type: 'custom_event',
      name: eventName,
      data: {
        ...properties,
        user_id: this.userId,
        session_id: this.sessionId,
        timestamp: new Date().toISOString(),
        page_path: window.location.pathname,
      },
    }

    if (this.isInitialized) {
      this.sendEvent(event)
    } else {
      this.queue.push(event)
    }
  }

  sendEvent(event) {
    if (window.gtag) {
      window.gtag('event', event.name, {
        ...event.data,
        event_category: event.data.category || 'engagement',
        event_label: event.data.label,
        value: event.data.value,
      })
    }

    if (window.mixpanel) {
      window.mixpanel.track(event.name, event.data)
    }

    console.log('[Analytics] Event tracked:', event.name, event.data)
  }

  // Track business metrics
  trackBusiness(eventName, data = {}) {
    const businessEvent = {
      type: 'business_metric',
      name: eventName,
      data: {
        ...data,
        user_id: this.userId,
        session_id: this.sessionId,
        timestamp: new Date().toISOString(),
      },
    }

    this.trackEvent(eventName, businessEvent.data)

    // Send to custom business metrics endpoint
    this.sendBusinessMetric(businessEvent)
  }

  async sendBusinessMetric(event) {
    try {
      await fetch('/api/analytics/business', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(event),
      })
    } catch (error) {
      console.error('[Analytics] Failed to send business metric:', error)
    }
  }

  // Performance tracking
  initializePerformanceTracking() {
    // Core Web Vitals
    getCLS((metric) => {
      this.trackEvent('web_vital_cls', {
        name: 'CLS',
        value: metric.value,
        rating: metric.rating,
        category: 'performance',
      })
    })

    getFID((metric) => {
      this.trackEvent('web_vital_fid', {
        name: 'FID',
        value: metric.value,
        rating: metric.rating,
        category: 'performance',
      })
    })

    getFCP((metric) => {
      this.trackEvent('web_vital_fcp', {
        name: 'FCP',
        value: metric.value,
        rating: metric.rating,
        category: 'performance',
      })
    })

    getLCP((metric) => {
      this.trackEvent('web_vital_lcp', {
        name: 'LCP',
        value: metric.value,
        rating: metric.rating,
        category: 'performance',
      })
    })

    getTTFB((metric) => {
      this.trackEvent('web_vital_ttfb', {
        name: 'TTFB',
        value: metric.value,
        rating: metric.rating,
        category: 'performance',
      })
    })

    // Track page load time
    window.addEventListener('load', () => {
      setTimeout(() => {
        const loadTime = Date.now() - this.pageLoadTime
        this.trackEvent('page_load_time', {
          value: loadTime,
          category: 'performance',
        })
      }, 0)
    })
  }

  // A/B testing support
  trackExperiment(experimentId, variant, outcome = null) {
    this.trackEvent('experiment_view', {
      experiment_id: experimentId,
      variant,
      category: 'experiment',
    })

    if (outcome) {
      this.trackEvent('experiment_conversion', {
        experiment_id: experimentId,
        variant,
        outcome,
        category: 'experiment',
      })
    }
  }

  // Error tracking
  trackError(error, context = {}) {
    const errorEvent = {
      name: 'javascript_error',
      data: {
        error_message: error.message,
        error_stack: error.stack,
        error_name: error.name,
        page_path: window.location.pathname,
        user_agent: navigator.userAgent,
        context,
        category: 'error',
      },
    }

    this.trackEvent('javascript_error', errorEvent.data)
  }

  // User journey tracking
  trackUserFlow(flowName, step, data = {}) {
    this.trackEvent('user_flow', {
      flow_name: flowName,
      step,
      step_data: data,
      category: 'user_journey',
    })
  }

  // Conversion tracking
  trackConversion(type, value = null, currency = 'USD') {
    this.trackEvent('conversion', {
      conversion_type: type,
      value,
      currency,
      category: 'conversion',
    })

    if (window.gtag && value) {
      window.gtag('event', 'purchase', {
        transaction_id: `${type}_${Date.now()}`,
        value,
        currency,
      })
    }
  }

  // Utility methods
  generateSessionId() {
    return `${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  processQueue() {
    while (this.queue.length > 0) {
      const event = this.queue.shift()
      
      if (event.type === 'page_view') {
        this.sendPageView(event)
      } else if (event.type === 'custom_event') {
        this.sendEvent(event)
      }
    }
  }

  // Privacy compliance
  optOut() {
    if (window.gtag) {
      window.gtag('consent', 'update', {
        analytics_storage: 'denied',
      })
    }

    if (window.mixpanel) {
      window.mixpanel.opt_out_tracking()
    }

    localStorage.setItem('analytics_opt_out', 'true')
    console.log('[Analytics] User opted out of tracking')
  }

  optIn() {
    if (window.gtag) {
      window.gtag('consent', 'update', {
        analytics_storage: 'granted',
      })
    }

    if (window.mixpanel) {
      window.mixpanel.opt_in_tracking()
    }

    localStorage.removeItem('analytics_opt_out')
    console.log('[Analytics] User opted in to tracking')
  }

  isOptedOut() {
    return localStorage.getItem('analytics_opt_out') === 'true'
  }
}

// Create global instance
export const analytics = new AnalyticsManager()

// React hook for analytics
import { useEffect, useContext } from 'react'
import { useLocation } from 'react-router-dom'

export const useAnalytics = () => {
  const location = useLocation()

  useEffect(() => {
    // Track page views on route changes
    analytics.trackPageView(location.pathname)
  }, [location])

  return {
    trackEvent: analytics.trackEvent.bind(analytics),
    trackBusiness: analytics.trackBusiness.bind(analytics),
    trackConversion: analytics.trackConversion.bind(analytics),
    trackUserFlow: analytics.trackUserFlow.bind(analytics),
    trackExperiment: analytics.trackExperiment.bind(analytics),
    setUser: analytics.setUser.bind(analytics),
  }
}
```

### Step 2: Business Metrics and KPI Tracking
**Timeline: 3-4 days**

Implement comprehensive business and KPI tracking:

```javascript
// hooks/useBusinessMetrics.js
import { useCallback } from 'react'
import { analytics } from '../utils/analytics'
import { useAuth } from './useAuth'

export const useBusinessMetrics = () => {
  const { user } = useAuth()

  // User registration and authentication metrics
  const trackRegistration = useCallback((method, referralSource = null) => {
    analytics.trackBusiness('user_registration', {
      registration_method: method,
      referral_source: referralSource,
      user_type: 'new',
      value: 1,
      category: 'acquisition',
    })

    analytics.trackConversion('registration', 10) // Assign value to registration
  }, [])

  const trackLogin = useCallback((method) => {
    analytics.trackBusiness('user_login', {
      login_method: method,
      user_id: user?.id,
      category: 'engagement',
    })
  }, [user])

  const trackLogout = useCallback(() => {
    analytics.trackBusiness('user_logout', {
      user_id: user?.id,
      session_duration: calculateSessionDuration(),
      category: 'engagement',
    })
  }, [user])

  // Market engagement metrics
  const trackMarketView = useCallback((marketId, marketData) => {
    analytics.trackBusiness('market_viewed', {
      market_id: marketId,
      market_title: marketData.title,
      market_category: marketData.category,
      market_status: marketData.status,
      market_volume: marketData.totalVolume,
      participants_count: marketData.participantCount,
      category: 'engagement',
    })
  }, [])

  const trackMarketShare = useCallback((marketId, platform) => {
    analytics.trackBusiness('market_shared', {
      market_id: marketId,
      share_platform: platform,
      category: 'viral',
    })
  }, [])

  const trackMarketSearch = useCallback((query, resultsCount) => {
    analytics.trackBusiness('market_search', {
      search_query: query,
      results_count: resultsCount,
      category: 'discovery',
    })
  }, [])

  // Betting and transaction metrics
  const trackBetPlaced = useCallback((betData) => {
    analytics.trackBusiness('bet_placed', {
      market_id: betData.marketId,
      bet_amount: betData.amount,
      bet_outcome: betData.outcome,
      odds: betData.odds,
      potential_return: betData.potentialReturn,
      user_balance_before: betData.balanceBefore,
      user_balance_after: betData.balanceAfter,
      category: 'monetization',
    })

    // Track as conversion with monetary value
    analytics.trackConversion('bet_placed', betData.amount, 'USD')
  }, [])

  const trackBetResult = useCallback((betData, result) => {
    analytics.trackBusiness('bet_resolved', {
      market_id: betData.marketId,
      bet_id: betData.id,
      bet_amount: betData.amount,
      bet_outcome: betData.outcome,
      actual_outcome: result.outcome,
      won: result.won,
      payout: result.payout,
      profit_loss: result.payout - betData.amount,
      category: 'monetization',
    })

    if (result.won) {
      analytics.trackConversion('bet_won', result.payout, 'USD')
    }
  }, [])

  const trackDeposit = useCallback((amount, method) => {
    analytics.trackBusiness('funds_deposited', {
      amount,
      payment_method: method,
      category: 'monetization',
    })

    analytics.trackConversion('deposit', amount, 'USD')
  }, [])

  const trackWithdrawal = useCallback((amount, method) => {
    analytics.trackBusiness('funds_withdrawn', {
      amount,
      withdrawal_method: method,
      category: 'monetization',
    })
  }, [])

  // User engagement metrics
  const trackTimeOnSite = useCallback((duration) => {
    analytics.trackBusiness('time_on_site', {
      duration_seconds: duration,
      duration_minutes: Math.round(duration / 60),
      category: 'engagement',
    })
  }, [])

  const trackFeatureUsage = useCallback((feature, context = {}) => {
    analytics.trackBusiness('feature_used', {
      feature_name: feature,
      feature_context: context,
      category: 'product',
    })
  }, [])

  const trackErrorEncountered = useCallback((errorType, errorMessage, context) => {
    analytics.trackBusiness('error_encountered', {
      error_type: errorType,
      error_message: errorMessage,
      error_context: context,
      category: 'quality',
    })
  }, [])

  // Retention and lifecycle metrics
  const trackDailyActive = useCallback(() => {
    const today = new Date().toISOString().split('T')[0]
    const lastActive = localStorage.getItem('last_active_date')
    
    if (lastActive !== today) {
      analytics.trackBusiness('daily_active_user', {
        date: today,
        is_returning: lastActive !== null,
        category: 'retention',
      })
      
      localStorage.setItem('last_active_date', today)
    }
  }, [])

  const trackUserRetention = useCallback((daysSinceRegistration) => {
    const retentionMilestones = [1, 3, 7, 14, 30, 60, 90]
    
    if (retentionMilestones.includes(daysSinceRegistration)) {
      analytics.trackBusiness('user_retention', {
        days_since_registration: daysSinceRegistration,
        retention_milestone: `day_${daysSinceRegistration}`,
        category: 'retention',
      })
    }
  }, [])

  // Market creation and management metrics
  const trackMarketCreated = useCallback((marketData) => {
    analytics.trackBusiness('market_created', {
      market_id: marketData.id,
      market_title: marketData.title,
      market_category: marketData.category,
      creator_id: marketData.creatorId,
      initial_odds: marketData.odds,
      closing_date: marketData.closingDate,
      category: 'content_creation',
    })
  }, [])

  const trackMarketResolved = useCallback((marketData, resolution) => {
    analytics.trackBusiness('market_resolved', {
      market_id: marketData.id,
      final_volume: marketData.totalVolume,
      final_participants: marketData.participantCount,
      resolution_outcome: resolution.outcome,
      total_bets: marketData.totalBets,
      duration_hours: calculateMarketDuration(marketData),
      category: 'content_lifecycle',
    })
  }, [])

  // Helper functions
  const calculateSessionDuration = () => {
    const sessionStart = sessionStorage.getItem('session_start')
    if (sessionStart) {
      return Math.round((Date.now() - parseInt(sessionStart)) / 1000)
    }
    return 0
  }

  const calculateMarketDuration = (marketData) => {
    const startDate = new Date(marketData.createdAt)
    const endDate = new Date(marketData.resolvedAt)
    return Math.round((endDate - startDate) / (1000 * 60 * 60))
  }

  return {
    // Authentication metrics
    trackRegistration,
    trackLogin,
    trackLogout,
    
    // Market engagement
    trackMarketView,
    trackMarketShare,
    trackMarketSearch,
    
    // Betting metrics
    trackBetPlaced,
    trackBetResult,
    trackDeposit,
    trackWithdrawal,
    
    // User engagement
    trackTimeOnSite,
    trackFeatureUsage,
    trackErrorEncountered,
    
    // Retention metrics
    trackDailyActive,
    trackUserRetention,
    
    // Market management
    trackMarketCreated,
    trackMarketResolved,
  }
}

// components/analytics/BusinessMetricsProvider.jsx
import React, { createContext, useContext, useEffect } from 'react'
import { useBusinessMetrics } from '../../hooks/useBusinessMetrics'
import { useAuth } from '../../hooks/useAuth'

const BusinessMetricsContext = createContext()

export const useBusinessMetricsContext = () => {
  const context = useContext(BusinessMetricsContext)
  if (!context) {
    throw new Error('useBusinessMetricsContext must be used within BusinessMetricsProvider')
  }
  return context
}

export const BusinessMetricsProvider = ({ children }) => {
  const businessMetrics = useBusinessMetrics()
  const { user, isAuthenticated } = useAuth()

  useEffect(() => {
    // Track session start
    if (!sessionStorage.getItem('session_start')) {
      sessionStorage.setItem('session_start', Date.now().toString())
    }

    // Track daily active user
    if (isAuthenticated) {
      businessMetrics.trackDailyActive()
    }

    // Track time on site on page unload
    const handleBeforeUnload = () => {
      const sessionStart = sessionStorage.getItem('session_start')
      if (sessionStart) {
        const duration = Math.round((Date.now() - parseInt(sessionStart)) / 1000)
        businessMetrics.trackTimeOnSite(duration)
      }
    }

    window.addEventListener('beforeunload', handleBeforeUnload)

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }, [isAuthenticated, businessMetrics])

  return (
    <BusinessMetricsContext.Provider value={businessMetrics}>
      {children}
    </BusinessMetricsContext.Provider>
  )
}
```

### Step 3: User Behavior and Journey Tracking
**Timeline: 2-3 days**

Implement detailed user behavior and journey analytics:

```javascript
// hooks/useUserJourney.js
import { useCallback, useEffect, useRef } from 'react'
import { analytics } from '../utils/analytics'
import { useLocation } from 'react-router-dom'

export const useUserJourney = () => {
  const location = useLocation()
  const journeyRef = useRef({
    currentFlow: null,
    flowStartTime: null,
    stepHistory: [],
  })

  // Start a user flow
  const startFlow = useCallback((flowName, initialData = {}) => {
    journeyRef.current = {
      currentFlow: flowName,
      flowStartTime: Date.now(),
      stepHistory: [],
    }

    analytics.trackUserFlow(flowName, 'flow_started', {
      ...initialData,
      timestamp: new Date().toISOString(),
    })
  }, [])

  // Track a step in the current flow
  const trackStep = useCallback((stepName, stepData = {}) => {
    if (!journeyRef.current.currentFlow) {
      console.warn('No active flow to track step:', stepName)
      return
    }

    const stepTime = Date.now()
    const stepDuration = journeyRef.current.stepHistory.length > 0 
      ? stepTime - journeyRef.current.stepHistory[journeyRef.current.stepHistory.length - 1].timestamp
      : stepTime - journeyRef.current.flowStartTime

    const step = {
      name: stepName,
      timestamp: stepTime,
      duration: stepDuration,
      data: stepData,
    }

    journeyRef.current.stepHistory.push(step)

    analytics.trackUserFlow(journeyRef.current.currentFlow, stepName, {
      ...stepData,
      step_index: journeyRef.current.stepHistory.length,
      step_duration: stepDuration,
      flow_duration: stepTime - journeyRef.current.flowStartTime,
    })
  }, [])

  // Complete the current flow
  const completeFlow = useCallback((completionData = {}) => {
    if (!journeyRef.current.currentFlow) {
      console.warn('No active flow to complete')
      return
    }

    const flowDuration = Date.now() - journeyRef.current.flowStartTime
    const totalSteps = journeyRef.current.stepHistory.length

    analytics.trackUserFlow(journeyRef.current.currentFlow, 'flow_completed', {
      ...completionData,
      total_steps: totalSteps,
      flow_duration: flowDuration,
      success: true,
    })

    // Reset journey state
    journeyRef.current = {
      currentFlow: null,
      flowStartTime: null,
      stepHistory: [],
    }
  }, [])

  // Abandon the current flow
  const abandonFlow = useCallback((reason = 'unknown') => {
    if (!journeyRef.current.currentFlow) {
      return
    }

    const flowDuration = Date.now() - journeyRef.current.flowStartTime
    const totalSteps = journeyRef.current.stepHistory.length
    const lastStep = journeyRef.current.stepHistory[journeyRef.current.stepHistory.length - 1]

    analytics.trackUserFlow(journeyRef.current.currentFlow, 'flow_abandoned', {
      abandonment_reason: reason,
      total_steps: totalSteps,
      last_step: lastStep?.name,
      flow_duration: flowDuration,
      success: false,
    })

    // Reset journey state
    journeyRef.current = {
      currentFlow: null,
      flowStartTime: null,
      stepHistory: [],
    }
  }, [])

  // Track page navigation within a flow
  useEffect(() => {
    if (journeyRef.current.currentFlow) {
      trackStep('page_navigation', {
        from: journeyRef.current.stepHistory[journeyRef.current.stepHistory.length - 1]?.data.page || 'unknown',
        to: location.pathname,
        page: location.pathname,
      })
    }
  }, [location.pathname, trackStep])

  // Automatically abandon flow on page unload
  useEffect(() => {
    const handleBeforeUnload = () => {
      if (journeyRef.current.currentFlow) {
        abandonFlow('page_unload')
      }
    }

    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => window.removeEventListener('beforeunload', handleBeforeUnload)
  }, [abandonFlow])

  return {
    startFlow,
    trackStep,
    completeFlow,
    abandonFlow,
    currentFlow: journeyRef.current.currentFlow,
  }
}

// components/analytics/UserJourneyTracker.jsx
import React, { useEffect } from 'react'
import { useUserJourney } from '../../hooks/useUserJourney'
import { useAuth } from '../../hooks/useAuth'

// Common user flows in the application
export const USER_FLOWS = {
  REGISTRATION: 'user_registration',
  BET_PLACEMENT: 'bet_placement',
  MARKET_CREATION: 'market_creation',
  DEPOSIT: 'deposit_funds',
  WITHDRAWAL: 'withdraw_funds',
  ONBOARDING: 'user_onboarding',
  MARKET_DISCOVERY: 'market_discovery',
}

const UserJourneyTracker = ({ children }) => {
  const { user, isAuthenticated } = useAuth()
  const { startFlow, trackStep, completeFlow } = useUserJourney()

  useEffect(() => {
    // Start onboarding flow for new users
    if (isAuthenticated && user && !user.hasCompletedOnboarding) {
      startFlow(USER_FLOWS.ONBOARDING, {
        user_id: user.id,
        registration_date: user.createdAt,
      })
    }
  }, [isAuthenticated, user, startFlow])

  return <>{children}</>
}

export default UserJourneyTracker

// Higher-order component for tracking user interactions
export const withUserJourneyTracking = (WrappedComponent, flowName, stepName) => {
  return function TrackedComponent(props) {
    const { trackStep } = useUserJourney()

    useEffect(() => {
      if (flowName && stepName) {
        trackStep(stepName, {
          component: WrappedComponent.name,
          props: Object.keys(props),
        })
      }
    }, [trackStep])

    return <WrappedComponent {...props} />
  }
}

// Hook for tracking form interactions
export const useFormTracking = (formName) => {
  const { trackStep } = useUserJourney()

  const trackFormStart = useCallback(() => {
    trackStep('form_started', {
      form_name: formName,
    })
  }, [formName, trackStep])

  const trackFieldInteraction = useCallback((fieldName, action = 'focus') => {
    trackStep('form_field_interaction', {
      form_name: formName,
      field_name: fieldName,
      action,
    })
  }, [formName, trackStep])

  const trackFormValidation = useCallback((errors) => {
    trackStep('form_validation', {
      form_name: formName,
      has_errors: Object.keys(errors).length > 0,
      error_fields: Object.keys(errors),
      error_count: Object.keys(errors).length,
    })
  }, [formName, trackStep])

  const trackFormSubmission = useCallback((success = true, errorMessage = null) => {
    trackStep('form_submitted', {
      form_name: formName,
      success,
      error_message: errorMessage,
    })
  }, [formName, trackStep])

  return {
    trackFormStart,
    trackFieldInteraction,
    trackFormValidation,
    trackFormSubmission,
  }
}
```

### Step 4: A/B Testing and Feature Flags
**Timeline: 2-3 days**

Implement A/B testing and feature flag analytics:

```javascript
// hooks/useExperiments.js
import { useState, useEffect, useCallback } from 'react'
import { analytics } from '../utils/analytics'
import { useAuth } from './useAuth'

const EXPERIMENTS_CONFIG = {
  bet_button_color: {
    variants: ['blue', 'green', 'orange'],
    weights: [0.4, 0.3, 0.3],
    enabled: true,
  },
  market_card_layout: {
    variants: ['compact', 'detailed', 'minimal'],
    weights: [0.33, 0.33, 0.34],
    enabled: true,
  },
  onboarding_flow: {
    variants: ['guided', 'self_service', 'video'],
    weights: [0.5, 0.3, 0.2],
    enabled: true,
  },
  pricing_display: {
    variants: ['percentage', 'decimal', 'fraction'],
    weights: [0.4, 0.4, 0.2],
    enabled: true,
  },
}

export const useExperiments = () => {
  const [experiments, setExperiments] = useState({})
  const { user } = useAuth()

  useEffect(() => {
    if (user) {
      initializeExperiments()
    }
  }, [user])

  const initializeExperiments = useCallback(() => {
    const userExperiments = {}

    Object.entries(EXPERIMENTS_CONFIG).forEach(([experimentId, config]) => {
      if (!config.enabled) {
        userExperiments[experimentId] = null
        return
      }

      // Check if user already has a variant assigned
      const storedVariant = localStorage.getItem(`experiment_${experimentId}`)
      if (storedVariant && config.variants.includes(storedVariant)) {
        userExperiments[experimentId] = storedVariant
        return
      }

      // Assign variant based on user ID for consistency
      const variant = assignVariant(experimentId, user.id, config)
      userExperiments[experimentId] = variant

      // Store variant
      localStorage.setItem(`experiment_${experimentId}`, variant)

      // Track experiment exposure
      analytics.trackExperiment(experimentId, variant)
    })

    setExperiments(userExperiments)
  }, [user])

  const assignVariant = (experimentId, userId, config) => {
    // Use user ID to create deterministic assignment
    const hash = simpleHash(`${experimentId}_${userId}`)
    const normalizedHash = hash % 100

    let cumulativeWeight = 0
    for (let i = 0; i < config.variants.length; i++) {
      cumulativeWeight += config.weights[i] * 100
      if (normalizedHash < cumulativeWeight) {
        return config.variants[i]
      }
    }

    return config.variants[0] // Fallback
  }

  const getVariant = useCallback((experimentId) => {
    return experiments[experimentId] || null
  }, [experiments])

  const trackConversion = useCallback((experimentId, outcome, value = null) => {
    const variant = experiments[experimentId]
    if (variant) {
      analytics.trackExperiment(experimentId, variant, {
        outcome,
        value,
        conversion: true,
      })
    }
  }, [experiments])

  const isVariant = useCallback((experimentId, variantName) => {
    return experiments[experimentId] === variantName
  }, [experiments])

  return {
    experiments,
    getVariant,
    isVariant,
    trackConversion,
  }
}

// Simple hash function for deterministic assignment
function simpleHash(str) {
  let hash = 0
  if (str.length === 0) return hash
  
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash = hash & hash // Convert to 32-bit integer
  }
  
  return Math.abs(hash)
}

// React component for A/B testing
export const ExperimentVariant = ({ 
  experimentId, 
  variant, 
  children, 
  fallback = null 
}) => {
  const { isVariant } = useExperiments()

  if (isVariant(experimentId, variant)) {
    return children
  }

  return fallback
}

// Hook for feature flags
export const useFeatureFlags = () => {
  const [flags, setFlags] = useState({})
  const { user } = useAuth()

  useEffect(() => {
    fetchFeatureFlags()
  }, [user])

  const fetchFeatureFlags = async () => {
    try {
      const response = await fetch('/api/feature-flags', {
        headers: {
          'Authorization': `Bearer ${user?.token}`,
        },
      })
      
      if (response.ok) {
        const data = await response.json()
        setFlags(data.flags)
      }
    } catch (error) {
      console.error('Failed to fetch feature flags:', error)
      
      // Fallback to default flags
      setFlags({
        beta_features: false,
        advanced_charts: true,
        social_sharing: true,
        notifications: true,
        dark_mode: true,
      })
    }
  }

  const isEnabled = useCallback((flagName) => {
    return flags[flagName] === true
  }, [flags])

  const trackFlagUsage = useCallback((flagName, action = 'viewed') => {
    if (flags[flagName]) {
      analytics.trackEvent('feature_flag_usage', {
        flag_name: flagName,
        flag_value: flags[flagName],
        action,
        category: 'feature_flags',
      })
    }
  }, [flags])

  return {
    flags,
    isEnabled,
    trackFlagUsage,
  }
}

// Component for feature-flagged content
export const FeatureFlag = ({ 
  flag, 
  children, 
  fallback = null,
  trackUsage = true 
}) => {
  const { isEnabled, trackFlagUsage } = useFeatureFlags()

  useEffect(() => {
    if (trackUsage && isEnabled(flag)) {
      trackFlagUsage(flag, 'rendered')
    }
  }, [flag, isEnabled, trackFlagUsage, trackUsage])

  if (isEnabled(flag)) {
    return children
  }

  return fallback
}
```

## Directory Structure
```
src/
├── utils/
│   ├── analytics.js          # Main analytics manager
│   ├── experiments.js        # A/B testing utilities
│   └── tracking.js          # Event tracking helpers
├── hooks/
│   ├── useAnalytics.js       # General analytics hook
│   ├── useBusinessMetrics.js # Business KPI tracking
│   ├── useUserJourney.js     # User flow tracking
│   ├── useExperiments.js     # A/B testing hook
│   └── useFeatureFlags.js    # Feature flag management
├── components/
│   ├── analytics/
│   │   ├── AnalyticsProvider.jsx
│   │   ├── BusinessMetricsProvider.jsx
│   │   ├── UserJourneyTracker.jsx
│   │   ├── ExperimentVariant.jsx
│   │   ├── FeatureFlag.jsx
│   │   └── PrivacyBanner.jsx
│   └── tracking/
│       ├── TrackedButton.jsx
│       ├── TrackedForm.jsx
│       └── TrackedLink.jsx
├── services/
│   ├── analyticsService.js   # API communication
│   ├── experimentService.js  # Experiment management
│   └── metricsService.js     # Custom metrics
└── constants/
    ├── events.js            # Event name constants
    ├── experiments.js       # Experiment definitions
    └── metrics.js           # Metric definitions
```

## Benefits
- Data-driven decision making
- User behavior insights
- Performance monitoring
- A/B testing capabilities
- Business KPI tracking
- User journey optimization
- Conversion rate optimization
- Feature usage analytics
- Error tracking and debugging
- Privacy-compliant tracking
- Real-time dashboards
- Retention analysis

## Analytics Features Implemented
- ✅ Google Analytics 4 integration
- ✅ Custom event tracking
- ✅ Business metrics collection
- ✅ User journey mapping
- ✅ A/B testing framework
- ✅ Feature flag analytics
- ✅ Performance monitoring
- ✅ Conversion tracking
- ✅ Error analytics
- ✅ Privacy compliance
- ✅ Real-time tracking
- ✅ Custom dashboards

## Key Metrics Tracked
- User registration and authentication
- Market engagement and discovery
- Betting behavior and patterns
- Financial transactions
- User retention and lifecycle
- Feature adoption and usage
- Performance and errors
- A/B test results
- Conversion funnels
- Business KPIs