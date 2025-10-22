# Monitoring and Observability Implementation Plan

## Overview
Implement comprehensive monitoring, logging, and observability solutions to ensure application performance, reliability, and quick issue resolution with real-time insights into user experience and system health.

## Current State Analysis
- No application monitoring
- Limited error tracking
- No performance monitoring
- No user experience monitoring
- No real-time alerting
- No centralized logging
- No observability dashboards
- No incident response system

## Implementation Steps

### Step 1: Application Performance Monitoring (APM)
**Timeline: 3-4 days**

Set up comprehensive APM with Sentry and custom monitoring:

```javascript
// utils/monitoring.js
import * as Sentry from '@sentry/react'
import { BrowserTracing } from '@sentry/tracing'
import { config } from './configManager'

class MonitoringManager {
  constructor() {
    this.isInitialized = false
    this.performanceObserver = null
    this.errorBoundaryCount = 0
    this.userSessionId = this.generateSessionId()
  }

  // Initialize monitoring services
  initialize() {
    try {
      this.initializeSentry()
      this.initializePerformanceMonitoring()
      this.initializeUserExperienceMonitoring()
      this.initializeCustomMetrics()
      
      this.isInitialized = true
      console.log('[Monitoring] Initialized successfully')
    } catch (error) {
      console.error('[Monitoring] Initialization failed:', error)
    }
  }

  // Sentry setup for error tracking and performance
  initializeSentry() {
    if (!config.get('SENTRY_DSN')) {
      console.warn('[Monitoring] Sentry DSN not configured')
      return
    }

    Sentry.init({
      dsn: config.get('SENTRY_DSN'),
      environment: config.get('APP_ENV'),
      
      // Performance monitoring
      integrations: [
        new BrowserTracing({
          // Set sampling rate for performance monitoring
          tracePropagationTargets: [
            'localhost',
            config.get('API_BASE_URL'),
            /^https:\/\/.*\.socialpredict\.com/,
          ],
        }),
      ],
      
      // Performance transaction sample rate
      tracesSampleRate: config.isProduction() ? 0.1 : 1.0,
      
      // Session replay for debugging
      replaysSessionSampleRate: config.isProduction() ? 0.01 : 0.1,
      replaysOnErrorSampleRate: 1.0,
      
      // Custom error filtering
      beforeSend(event) {
        // Filter out development errors
        if (config.isDevelopment() && event.level === 'warning') {
          return null
        }
        
        // Filter out network errors from ad blockers
        if (event.exception?.values?.[0]?.value?.includes('Network request failed')) {
          return null
        }
        
        return event
      },
      
      // Custom context
      initialScope: {
        tags: {
          component: 'frontend',
          version: process.env.REACT_APP_VERSION || 'unknown',
        },
        user: {
          session_id: this.userSessionId,
        },
      },
    })
  }

  // Performance monitoring setup
  initializePerformanceMonitoring() {
    // Core Web Vitals monitoring
    this.monitorCoreWebVitals()
    
    // Resource timing monitoring
    this.monitorResourceTiming()
    
    // Long task monitoring
    this.monitorLongTasks()
    
    // Memory usage monitoring
    this.monitorMemoryUsage()
    
    // Bundle size monitoring
    this.monitorBundleSize()
  }

  monitorCoreWebVitals() {
    if ('web-vitals' in window) {
      import('web-vitals').then(({ getCLS, getFID, getFCP, getLCP, getTTFB }) => {
        getCLS((metric) => this.sendMetric('web-vital', 'CLS', metric))
        getFID((metric) => this.sendMetric('web-vital', 'FID', metric))
        getFCP((metric) => this.sendMetric('web-vital', 'FCP', metric))
        getLCP((metric) => this.sendMetric('web-vital', 'LCP', metric))
        getTTFB((metric) => this.sendMetric('web-vital', 'TTFB', metric))
      })
    }
  }

  monitorResourceTiming() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'resource') {
            this.analyzeResourceTiming(entry)
          }
        }
      })
      
      observer.observe({ entryTypes: ['resource'] })
      this.performanceObserver = observer
    }
  }

  analyzeResourceTiming(entry) {
    const { name, duration, transferSize, encodedBodySize } = entry
    
    // Track slow resources
    if (duration > 1000) {
      this.sendMetric('performance', 'slow-resource', {
        url: name,
        duration,
        size: transferSize,
        type: entry.initiatorType,
      })
    }
    
    // Track large resources
    if (transferSize > 1024 * 1024) { // 1MB
      this.sendMetric('performance', 'large-resource', {
        url: name,
        size: transferSize,
        compressed_size: encodedBodySize,
        compression_ratio: encodedBodySize / transferSize,
      })
    }
  }

  monitorLongTasks() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.entryType === 'longtask' && entry.duration > 50) {
            this.sendMetric('performance', 'long-task', {
              duration: entry.duration,
              start_time: entry.startTime,
              attribution: entry.attribution?.[0]?.name || 'unknown',
            })
          }
        }
      })
      
      try {
        observer.observe({ entryTypes: ['longtask'] })
      } catch (error) {
        console.warn('[Monitoring] Long task monitoring not supported')
      }
    }
  }

  monitorMemoryUsage() {
    if ('memory' in performance) {
      setInterval(() => {
        const memInfo = performance.memory
        
        this.sendMetric('performance', 'memory-usage', {
          used: memInfo.usedJSHeapSize,
          total: memInfo.totalJSHeapSize,
          limit: memInfo.jsHeapSizeLimit,
          usage_percentage: (memInfo.usedJSHeapSize / memInfo.jsHeapSizeLimit) * 100,
        })
        
        // Alert on high memory usage
        if (memInfo.usedJSHeapSize / memInfo.jsHeapSizeLimit > 0.9) {
          this.sendAlert('high-memory-usage', {
            usage_percentage: (memInfo.usedJSHeapSize / memInfo.jsHeapSizeLimit) * 100,
          })
        }
      }, 30000) // Check every 30 seconds
    }
  }

  monitorBundleSize() {
    // Monitor chunk loading times
    const originalImport = window.__webpack_require__?.cache || {}
    
    if (typeof window.performance?.getEntriesByType === 'function') {
      const resources = window.performance.getEntriesByType('resource')
      const jsResources = resources.filter(r => r.name.includes('.js'))
      
      const totalBundleSize = jsResources.reduce((acc, resource) => acc + (resource.transferSize || 0), 0)
      
      this.sendMetric('performance', 'bundle-size', {
        total_size: totalBundleSize,
        chunk_count: jsResources.length,
        largest_chunk: Math.max(...jsResources.map(r => r.transferSize || 0)),
      })
    }
  }

  // User experience monitoring
  initializeUserExperienceMonitoring() {
    // Rage click detection
    this.monitorRageClicks()
    
    // Dead click detection
    this.monitorDeadClicks()
    
    // Form abandonment
    this.monitorFormAbandonment()
    
    // Page visibility changes
    this.monitorPageVisibility()
    
    // Network status changes
    this.monitorNetworkStatus()
  }

  monitorRageClicks() {
    let clickCount = 0
    let clickTimer = null
    let lastClickTarget = null

    document.addEventListener('click', (event) => {
      const target = event.target

      if (target === lastClickTarget) {
        clickCount++
        
        if (clickTimer) clearTimeout(clickTimer)
        
        clickTimer = setTimeout(() => {
          if (clickCount >= 3) {
            this.sendMetric('ux', 'rage-click', {
              element: this.getElementSelector(target),
              click_count: clickCount,
              page: window.location.pathname,
            })
          }
          clickCount = 0
          lastClickTarget = null
        }, 1000)
      } else {
        clickCount = 1
        lastClickTarget = target
      }
    })
  }

  monitorDeadClicks() {
    document.addEventListener('click', (event) => {
      const target = event.target
      
      // Check if click resulted in any DOM changes or navigation
      const initialHTML = document.body.innerHTML
      const initialURL = window.location.href
      
      setTimeout(() => {
        if (document.body.innerHTML === initialHTML && window.location.href === initialURL) {
          // No changes detected - potential dead click
          this.sendMetric('ux', 'dead-click', {
            element: this.getElementSelector(target),
            page: window.location.pathname,
          })
        }
      }, 100)
    })
  }

  monitorFormAbandonment() {
    const forms = new Map()
    
    document.addEventListener('focus', (event) => {
      if (event.target.tagName === 'INPUT' || event.target.tagName === 'TEXTAREA') {
        const form = event.target.closest('form')
        if (form && !forms.has(form)) {
          forms.set(form, {
            started_at: Date.now(),
            fields_interacted: new Set(),
            form_selector: this.getElementSelector(form),
          })
        }
        
        if (form) {
          forms.get(form).fields_interacted.add(event.target.name || event.target.id)
        }
      }
    })
    
    document.addEventListener('submit', (event) => {
      const form = event.target
      if (forms.has(form)) {
        forms.delete(form)
      }
    })
    
    // Check for abandoned forms on page unload
    window.addEventListener('beforeunload', () => {
      forms.forEach((data, form) => {
        const duration = Date.now() - data.started_at
        if (duration > 5000 && data.fields_interacted.size > 0) { // At least 5 seconds
          this.sendMetric('ux', 'form-abandonment', {
            form_selector: data.form_selector,
            duration,
            fields_interacted: data.fields_interacted.size,
            page: window.location.pathname,
          })
        }
      })
    })
  }

  monitorPageVisibility() {
    let visibilityStart = Date.now()
    
    document.addEventListener('visibilitychange', () => {
      if (document.hidden) {
        const visibleDuration = Date.now() - visibilityStart
        this.sendMetric('ux', 'page-visibility', {
          event: 'hidden',
          visible_duration: visibleDuration,
          page: window.location.pathname,
        })
      } else {
        visibilityStart = Date.now()
        this.sendMetric('ux', 'page-visibility', {
          event: 'visible',
          page: window.location.pathname,
        })
      }
    })
  }

  monitorNetworkStatus() {
    if ('connection' in navigator) {
      const connection = navigator.connection
      
      this.sendMetric('network', 'connection-info', {
        effective_type: connection.effectiveType,
        downlink: connection.downlink,
        rtt: connection.rtt,
        save_data: connection.saveData,
      })
      
      connection.addEventListener('change', () => {
        this.sendMetric('network', 'connection-change', {
          effective_type: connection.effectiveType,
          downlink: connection.downlink,
          rtt: connection.rtt,
          save_data: connection.saveData,
        })
      })
    }
    
    window.addEventListener('online', () => {
      this.sendMetric('network', 'status-change', { status: 'online' })
    })
    
    window.addEventListener('offline', () => {
      this.sendMetric('network', 'status-change', { status: 'offline' })
    })
  }

  // Custom metrics collection
  initializeCustomMetrics() {
    // API response time tracking
    this.monitorAPIPerformance()
    
    // Route change performance
    this.monitorRouteChanges()
    
    // Component render performance
    this.monitorComponentPerformance()
  }

  monitorAPIPerformance() {
    // Monkey patch fetch to monitor API calls
    const originalFetch = window.fetch
    
    window.fetch = async (url, options = {}) => {
      const startTime = performance.now()
      const method = options.method || 'GET'
      
      try {
        const response = await originalFetch(url, options)
        const duration = performance.now() - startTime
        
        this.sendMetric('api', 'request-performance', {
          url: typeof url === 'string' ? url : url.href,
          method,
          status: response.status,
          duration,
          success: response.ok,
        })
        
        // Alert on slow API calls
        if (duration > 5000) {
          this.sendAlert('slow-api-call', {
            url: typeof url === 'string' ? url : url.href,
            method,
            duration,
          })
        }
        
        return response
      } catch (error) {
        const duration = performance.now() - startTime
        
        this.sendMetric('api', 'request-error', {
          url: typeof url === 'string' ? url : url.href,
          method,
          duration,
          error: error.message,
        })
        
        throw error
      }
    }
  }

  monitorRouteChanges() {
    let routeChangeStart = performance.now()
    
    // Monitor history changes
    const originalPushState = history.pushState
    const originalReplaceState = history.replaceState
    
    history.pushState = function(...args) {
      const duration = performance.now() - routeChangeStart
      monitoring.sendMetric('navigation', 'route-change', {
        from: window.location.pathname,
        to: args[2],
        duration,
        method: 'pushState',
      })
      routeChangeStart = performance.now()
      return originalPushState.apply(this, args)
    }
    
    history.replaceState = function(...args) {
      const duration = performance.now() - routeChangeStart
      monitoring.sendMetric('navigation', 'route-change', {
        from: window.location.pathname,
        to: args[2],
        duration,
        method: 'replaceState',
      })
      routeChangeStart = performance.now()
      return originalReplaceState.apply(this, args)
    }
    
    window.addEventListener('popstate', () => {
      const duration = performance.now() - routeChangeStart
      monitoring.sendMetric('navigation', 'route-change', {
        to: window.location.pathname,
        duration,
        method: 'popstate',
      })
      routeChangeStart = performance.now()
    })
  }

  // Error tracking and reporting
  captureError(error, context = {}) {
    if (this.isInitialized) {
      Sentry.captureException(error, {
        tags: {
          component: context.component || 'unknown',
          action: context.action || 'unknown',
        },
        extra: context,
      })
    }
    
    this.sendMetric('error', 'exception', {
      message: error.message,
      stack: error.stack,
      name: error.name,
      context,
    })
  }

  // User identification
  setUser(user) {
    if (this.isInitialized) {
      Sentry.setUser({
        id: user.id,
        email: user.email,
        username: user.username,
      })
    }
  }

  // Custom metric sending
  sendMetric(category, name, data = {}) {
    const metric = {
      category,
      name,
      data: {
        ...data,
        timestamp: Date.now(),
        session_id: this.userSessionId,
        url: window.location.href,
        user_agent: navigator.userAgent,
      },
    }
    
    // Send to custom analytics endpoint
    this.sendToAnalytics(metric)
    
    // Also send to Sentry as breadcrumb
    if (this.isInitialized) {
      Sentry.addBreadcrumb({
        category,
        message: name,
        data,
        level: 'info',
      })
    }
  }

  sendAlert(type, data = {}) {
    const alert = {
      type,
      severity: 'warning',
      data: {
        ...data,
        timestamp: Date.now(),
        session_id: this.userSessionId,
        url: window.location.href,
      },
    }
    
    // Send to alerting system
    this.sendToAlerting(alert)
    
    // Also capture in Sentry
    if (this.isInitialized) {
      Sentry.captureMessage(`Alert: ${type}`, 'warning')
    }
  }

  async sendToAnalytics(metric) {
    try {
      await fetch('/api/metrics', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(metric),
      })
    } catch (error) {
      console.warn('[Monitoring] Failed to send metric:', error)
    }
  }

  async sendToAlerting(alert) {
    try {
      await fetch('/api/alerts', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(alert),
      })
    } catch (error) {
      console.warn('[Monitoring] Failed to send alert:', error)
    }
  }

  // Utility methods
  getElementSelector(element) {
    if (element.id) return `#${element.id}`
    if (element.className) return `.${element.className.split(' ')[0]}`
    return element.tagName.toLowerCase()
  }

  generateSessionId() {
    return `${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  // Cleanup
  destroy() {
    if (this.performanceObserver) {
      this.performanceObserver.disconnect()
    }
  }
}

// Create global monitoring instance
export const monitoring = new MonitoringManager()

// React error boundary with monitoring
import React from 'react'

export class MonitoredErrorBoundary extends React.Component {
  constructor(props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error }
  }

  componentDidCatch(error, errorInfo) {
    monitoring.captureError(error, {
      component: this.props.name || 'ErrorBoundary',
      errorInfo,
      props: this.props,
    })
    
    monitoring.errorBoundaryCount++
    
    if (monitoring.errorBoundaryCount > 5) {
      monitoring.sendAlert('multiple-error-boundaries', {
        count: monitoring.errorBoundaryCount,
      })
    }
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div className="error-boundary">
          <h2>Something went wrong</h2>
          <p>We've been notified of this error and are working to fix it.</p>
          <button onClick={() => window.location.reload()}>
            Reload Page
          </button>
        </div>
      )
    }

    return this.props.children
  }
}

// HOC for monitoring component performance
export const withPerformanceMonitoring = (WrappedComponent, componentName) => {
  return function MonitoredComponent(props) {
    const [renderStart] = React.useState(() => performance.now())
    
    React.useEffect(() => {
      const renderEnd = performance.now()
      const renderTime = renderEnd - renderStart
      
      monitoring.sendMetric('component', 'render-performance', {
        component: componentName || WrappedComponent.name,
        render_time: renderTime,
        props_count: Object.keys(props).length,
      })
      
      if (renderTime > 100) {
        monitoring.sendAlert('slow-component-render', {
          component: componentName || WrappedComponent.name,
          render_time: renderTime,
        })
      }
    }, [renderStart, props])
    
    return <WrappedComponent {...props} />
  }
}

// Hook for monitoring custom events
export const useMonitoring = () => {
  return {
    sendMetric: monitoring.sendMetric.bind(monitoring),
    sendAlert: monitoring.sendAlert.bind(monitoring),
    captureError: monitoring.captureError.bind(monitoring),
    setUser: monitoring.setUser.bind(monitoring),
  }
}
```

### Step 2: Real-time Dashboards and Alerting
**Timeline: 3-4 days**

Create comprehensive monitoring dashboards and alerting system:

```javascript
// components/monitoring/MonitoringDashboard.jsx
import React, { useState, useEffect } from 'react'
import { 
  ExclamationTriangleIcon, 
  CheckCircleIcon, 
  ClockIcon,
  ChartBarIcon,
  UsersIcon,
  ServerIcon,
} from '@heroicons/react/24/outline'

const MonitoringDashboard = () => {
  const [metrics, setMetrics] = useState({
    health: 'healthy',
    performance: {},
    errors: [],
    users: {},
    infrastructure: {},
  })

  const [alerts, setAlerts] = useState([])
  const [realTimeData, setRealTimeData] = useState({})

  useEffect(() => {
    // Fetch initial metrics
    fetchMetrics()
    
    // Set up real-time updates
    const ws = new WebSocket(process.env.REACT_APP_MONITORING_WS_URL)
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data)
      updateRealTimeData(data)
    }
    
    // Polling fallback
    const interval = setInterval(fetchMetrics, 30000)
    
    return () => {
      ws.close()
      clearInterval(interval)
    }
  }, [])

  const fetchMetrics = async () => {
    try {
      const response = await fetch('/api/monitoring/dashboard')
      const data = await response.json()
      setMetrics(data.metrics)
      setAlerts(data.alerts)
    } catch (error) {
      console.error('Failed to fetch metrics:', error)
    }
  }

  const updateRealTimeData = (data) => {
    setRealTimeData(prev => ({
      ...prev,
      [data.type]: data.value,
    }))
  }

  const getHealthStatus = () => {
    const { health } = metrics
    const statusConfig = {
      healthy: { color: 'green', icon: CheckCircleIcon, label: 'Healthy' },
      degraded: { color: 'yellow', icon: ExclamationTriangleIcon, label: 'Degraded' },
      unhealthy: { color: 'red', icon: ExclamationTriangleIcon, label: 'Unhealthy' },
    }
    
    return statusConfig[health] || statusConfig.unhealthy
  }

  const healthStatus = getHealthStatus()

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <h1 className="text-3xl font-bold text-gray-900">
              System Monitoring
            </h1>
            
            <div className="flex items-center space-x-3">
              <healthStatus.icon className={`h-6 w-6 text-${healthStatus.color}-500`} />
              <span className={`text-lg font-medium text-${healthStatus.color}-600`}>
                {healthStatus.label}
              </span>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Alerts Section */}
        {alerts.length > 0 && (
          <div className="mb-8">
            <h2 className="text-xl font-semibold text-gray-900 mb-4">
              Active Alerts
            </h2>
            <div className="space-y-4">
              {alerts.map((alert, index) => (
                <AlertCard key={index} alert={alert} />
              ))}
            </div>
          </div>
        )}

        {/* Metrics Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <MetricCard
            title="Response Time"
            value={`${metrics.performance.avgResponseTime || 0}ms`}
            change={realTimeData.responseTime}
            icon={ClockIcon}
            color="blue"
          />
          
          <MetricCard
            title="Error Rate"
            value={`${metrics.performance.errorRate || 0}%`}
            change={realTimeData.errorRate}
            icon={ExclamationTriangleIcon}
            color="red"
          />
          
          <MetricCard
            title="Active Users"
            value={metrics.users.active || 0}
            change={realTimeData.activeUsers}
            icon={UsersIcon}
            color="green"
          />
          
          <MetricCard
            title="Throughput"
            value={`${metrics.performance.requestsPerSecond || 0}/s`}
            change={realTimeData.throughput}
            icon={ChartBarIcon}
            color="purple"
          />
        </div>

        {/* Charts Section */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <PerformanceChart data={metrics.performance.timeSeries} />
          <ErrorChart data={metrics.errors} />
        </div>

        {/* Infrastructure Status */}
        <div className="mt-8">
          <InfrastructureStatus data={metrics.infrastructure} />
        </div>
      </main>
    </div>
  )
}

// Alert Card Component
const AlertCard = ({ alert }) => {
  const severityColors = {
    critical: 'red',
    warning: 'yellow',
    info: 'blue',
  }
  
  const color = severityColors[alert.severity] || 'gray'

  return (
    <div className={`border-l-4 border-${color}-400 bg-${color}-50 p-4`}>
      <div className="flex items-center">
        <ExclamationTriangleIcon className={`h-5 w-5 text-${color}-400 mr-3`} />
        <div className="flex-1">
          <h3 className={`text-sm font-medium text-${color}-800`}>
            {alert.title}
          </h3>
          <p className={`text-sm text-${color}-700 mt-1`}>
            {alert.message}
          </p>
          <p className={`text-xs text-${color}-600 mt-2`}>
            {new Date(alert.timestamp).toLocaleString()}
          </p>
        </div>
      </div>
    </div>
  )
}

// Metric Card Component
const MetricCard = ({ title, value, change, icon: Icon, color }) => {
  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <Icon className={`h-6 w-6 text-${color}-600`} />
          </div>
          <div className="ml-5 w-0 flex-1">
            <dl>
              <dt className="text-sm font-medium text-gray-500 truncate">
                {title}
              </dt>
              <dd className="text-lg font-medium text-gray-900">
                {value}
              </dd>
              {change && (
                <dd className={`text-sm ${change >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                  {change >= 0 ? '+' : ''}{change}
                </dd>
              )}
            </dl>
          </div>
        </div>
      </div>
    </div>
  )
}

// Performance Chart Component
const PerformanceChart = ({ data = [] }) => {
  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900">
          Performance Trends
        </h3>
        <div className="mt-5">
          {/* Chart implementation would go here */}
          <div className="h-64 flex items-center justify-center bg-gray-50 rounded">
            <p className="text-gray-500">Performance chart would be rendered here</p>
          </div>
        </div>
      </div>
    </div>
  )
}

// Error Chart Component
const ErrorChart = ({ data = [] }) => {
  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900">
          Error Trends
        </h3>
        <div className="mt-5">
          {/* Chart implementation would go here */}
          <div className="h-64 flex items-center justify-center bg-gray-50 rounded">
            <p className="text-gray-500">Error chart would be rendered here</p>
          </div>
        </div>
      </div>
    </div>
  )
}

// Infrastructure Status Component
const InfrastructureStatus = ({ data = {} }) => {
  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
          Infrastructure Status
        </h3>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <StatusItem
            title="Frontend CDN"
            status={data.cdn?.status || 'unknown'}
            uptime={data.cdn?.uptime || 0}
          />
          
          <StatusItem
            title="API Gateway"
            status={data.api?.status || 'unknown'}
            uptime={data.api?.uptime || 0}
          />
          
          <StatusItem
            title="Database"
            status={data.database?.status || 'unknown'}
            uptime={data.database?.uptime || 0}
          />
        </div>
      </div>
    </div>
  )
}

// Status Item Component
const StatusItem = ({ title, status, uptime }) => {
  const statusColors = {
    healthy: 'green',
    degraded: 'yellow',
    unhealthy: 'red',
    unknown: 'gray',
  }
  
  const color = statusColors[status] || 'gray'

  return (
    <div className="text-center">
      <div className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-${color}-100 text-${color}-800 mb-2`}>
        {status}
      </div>
      <h4 className="text-sm font-medium text-gray-900">{title}</h4>
      <p className="text-sm text-gray-500">
        {uptime}% uptime
      </p>
    </div>
  )
}

export default MonitoringDashboard

// services/alertingService.js
class AlertingService {
  constructor() {
    this.subscribers = new Map()
    this.alertHistory = []
    this.thresholds = {
      responseTime: 1000,
      errorRate: 5,
      memoryUsage: 90,
      diskUsage: 85,
    }
  }

  // Subscribe to alerts
  subscribe(callback) {
    const id = Date.now().toString()
    this.subscribers.set(id, callback)
    
    return () => {
      this.subscribers.delete(id)
    }
  }

  // Process incoming metric and check thresholds
  processMetric(metric) {
    const alerts = this.checkThresholds(metric)
    
    alerts.forEach(alert => {
      this.sendAlert(alert)
    })
  }

  checkThresholds(metric) {
    const alerts = []
    
    switch (metric.category) {
      case 'performance':
        if (metric.name === 'api-response-time' && metric.data.duration > this.thresholds.responseTime) {
          alerts.push({
            type: 'slow_response',
            severity: 'warning',
            title: 'Slow API Response',
            message: `API response time (${metric.data.duration}ms) exceeded threshold (${this.thresholds.responseTime}ms)`,
            timestamp: Date.now(),
            data: metric.data,
          })
        }
        break
        
      case 'error':
        if (metric.name === 'error-rate' && metric.data.rate > this.thresholds.errorRate) {
          alerts.push({
            type: 'high_error_rate',
            severity: 'critical',
            title: 'High Error Rate',
            message: `Error rate (${metric.data.rate}%) exceeded threshold (${this.thresholds.errorRate}%)`,
            timestamp: Date.now(),
            data: metric.data,
          })
        }
        break
        
      case 'system':
        if (metric.name === 'memory-usage' && metric.data.usage_percentage > this.thresholds.memoryUsage) {
          alerts.push({
            type: 'high_memory_usage',
            severity: 'warning',
            title: 'High Memory Usage',
            message: `Memory usage (${metric.data.usage_percentage}%) exceeded threshold (${this.thresholds.memoryUsage}%)`,
            timestamp: Date.now(),
            data: metric.data,
          })
        }
        break
    }
    
    return alerts
  }

  sendAlert(alert) {
    // Add to history
    this.alertHistory.unshift(alert)
    
    // Keep only last 1000 alerts
    if (this.alertHistory.length > 1000) {
      this.alertHistory = this.alertHistory.slice(0, 1000)
    }
    
    // Notify subscribers
    this.subscribers.forEach(callback => {
      try {
        callback(alert)
      } catch (error) {
        console.error('Error notifying alert subscriber:', error)
      }
    })
    
    // Send to external services
    this.sendToSlack(alert)
    this.sendToPagerDuty(alert)
    this.sendEmail(alert)
  }

  async sendToSlack(alert) {
    if (!process.env.REACT_APP_SLACK_WEBHOOK_URL) return
    
    const color = {
      critical: 'danger',
      warning: 'warning',
      info: 'good',
    }[alert.severity] || 'warning'
    
    const payload = {
      channel: '#alerts',
      username: 'SocialPredict Monitoring',
      attachments: [{
        color,
        title: alert.title,
        text: alert.message,
        timestamp: Math.floor(alert.timestamp / 1000),
        fields: Object.entries(alert.data || {}).map(([key, value]) => ({
          title: key,
          value: String(value),
          short: true,
        })),
      }],
    }
    
    try {
      await fetch(process.env.REACT_APP_SLACK_WEBHOOK_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })
    } catch (error) {
      console.error('Failed to send Slack alert:', error)
    }
  }

  async sendToPagerDuty(alert) {
    if (alert.severity !== 'critical' || !process.env.REACT_APP_PAGERDUTY_ROUTING_KEY) return
    
    const payload = {
      routing_key: process.env.REACT_APP_PAGERDUTY_ROUTING_KEY,
      event_action: 'trigger',
      payload: {
        summary: alert.title,
        source: 'socialpredict-frontend',
        severity: alert.severity,
        component: 'frontend',
        group: 'monitoring',
        class: alert.type,
        custom_details: alert.data,
      },
    }
    
    try {
      await fetch('https://events.pagerduty.com/v2/enqueue', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })
    } catch (error) {
      console.error('Failed to send PagerDuty alert:', error)
    }
  }

  async sendEmail(alert) {
    if (alert.severity === 'info') return
    
    try {
      await fetch('/api/alerts/email', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(alert),
      })
    } catch (error) {
      console.error('Failed to send email alert:', error)
    }
  }

  // Get alert history
  getAlertHistory(limit = 100) {
    return this.alertHistory.slice(0, limit)
  }

  // Update thresholds
  updateThresholds(newThresholds) {
    this.thresholds = { ...this.thresholds, ...newThresholds }
  }
}

export const alertingService = new AlertingService()
```

### Step 3: Logging and Log Management
**Timeline: 2-3 days**

Implement comprehensive logging system:

```javascript
// utils/logger.js
import { config } from './configManager'

class Logger {
  constructor() {
    this.logLevel = this.getLogLevel()
    this.context = {}
    this.buffer = []
    this.maxBufferSize = 1000
    this.flushInterval = 5000 // 5 seconds
    
    this.startPeriodicFlush()
  }

  getLogLevel() {
    const levels = {
      debug: 0,
      info: 1,
      warn: 2,
      error: 3,
    }
    
    const configLevel = config.get('LOG_LEVEL', 'info')
    return levels[configLevel] || levels.info
  }

  setContext(context) {
    this.context = { ...this.context, ...context }
  }

  clearContext() {
    this.context = {}
  }

  debug(message, data = {}) {
    this.log('debug', message, data)
  }

  info(message, data = {}) {
    this.log('info', message, data)
  }

  warn(message, data = {}) {
    this.log('warn', message, data)
  }

  error(message, data = {}) {
    this.log('error', message, data)
  }

  log(level, message, data = {}) {
    const levels = { debug: 0, info: 1, warn: 2, error: 3 }
    
    if (levels[level] < this.logLevel) {
      return
    }

    const logEntry = {
      timestamp: new Date().toISOString(),
      level,
      message,
      data: {
        ...this.context,
        ...data,
      },
      meta: {
        url: window.location.href,
        userAgent: navigator.userAgent,
        session: this.getSessionId(),
        version: process.env.REACT_APP_VERSION || 'unknown',
      },
    }

    // Console output in development
    if (config.isDevelopment()) {
      this.logToConsole(logEntry)
    }

    // Add to buffer for batch sending
    this.buffer.push(logEntry)

    // Flush if buffer is full or if error level
    if (this.buffer.length >= this.maxBufferSize || level === 'error') {
      this.flush()
    }
  }

  logToConsole(entry) {
    const { level, message, data } = entry
    const style = {
      debug: 'color: #6B7280',
      info: 'color: #3B82F6',
      warn: 'color: #F59E0B',
      error: 'color: #EF4444; font-weight: bold',
    }[level]

    console.log(
      `%c[${entry.timestamp}] ${level.toUpperCase()}: ${message}`,
      style,
      data
    )
  }

  startPeriodicFlush() {
    setInterval(() => {
      if (this.buffer.length > 0) {
        this.flush()
      }
    }, this.flushInterval)
  }

  async flush() {
    if (this.buffer.length === 0) return

    const logs = [...this.buffer]
    this.buffer = []

    try {
      await this.sendLogs(logs)
    } catch (error) {
      console.error('Failed to send logs:', error)
      // Re-add logs to buffer on failure (up to max size)
      this.buffer = [...logs.slice(-this.maxBufferSize / 2), ...this.buffer]
    }
  }

  async sendLogs(logs) {
    // Send to custom logging endpoint
    await fetch('/api/logs', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ logs }),
    })

    // Also send to external logging service if configured
    if (config.get('LOGTAIL_TOKEN')) {
      await this.sendToLogtail(logs)
    }
  }

  async sendToLogtail(logs) {
    try {
      await fetch('https://in.logtail.com/', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${config.get('LOGTAIL_TOKEN')}`,
        },
        body: JSON.stringify(logs),
      })
    } catch (error) {
      console.error('Failed to send logs to Logtail:', error)
    }
  }

  getSessionId() {
    let sessionId = sessionStorage.getItem('logging_session_id')
    if (!sessionId) {
      sessionId = `${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      sessionStorage.setItem('logging_session_id', sessionId)
    }
    return sessionId
  }

  // Special method for API requests
  logAPIRequest(method, url, duration, status, error = null) {
    const level = error ? 'error' : status >= 400 ? 'warn' : 'info'
    
    this.log(level, `API ${method} ${url}`, {
      method,
      url,
      duration,
      status,
      error: error?.message,
      category: 'api',
    })
  }

  // Special method for user actions  
  logUserAction(action, data = {}) {
    this.info(`User action: ${action}`, {
      action,
      ...data,
      category: 'user_action',
    })
  }

  // Special method for performance metrics
  logPerformance(metric, value, context = {}) {
    this.info(`Performance: ${metric}`, {
      metric,
      value,
      ...context,
      category: 'performance',
    })
  }

  // Special method for business events
  logBusinessEvent(event, data = {}) {
    this.info(`Business event: ${event}`, {
      event,
      ...data,
      category: 'business',
    })
  }
}

// Create global logger instance
export const logger = new Logger()

// React hook for logging
export const useLogger = () => {
  const contextualLogger = {
    debug: (message, data) => logger.debug(message, data),
    info: (message, data) => logger.info(message, data),
    warn: (message, data) => logger.warn(message, data),
    error: (message, data) => logger.error(message, data),
    logUserAction: (action, data) => logger.logUserAction(action, data),
    logPerformance: (metric, value, context) => logger.logPerformance(metric, value, context),
    logBusinessEvent: (event, data) => logger.logBusinessEvent(event, data),
  }

  return contextualLogger
}

// HOC for component logging
export const withLogging = (WrappedComponent, componentName) => {
  return function LoggedComponent(props) {
    const componentLogger = useLogger()

    React.useEffect(() => {
      logger.setContext({ component: componentName })
      componentLogger.debug(`Component ${componentName} mounted`)

      return () => {
        componentLogger.debug(`Component ${componentName} unmounted`)
        logger.clearContext()
      }
    }, [componentLogger])

    return <WrappedComponent {...props} />
  }
}

// Global error handler with logging
window.addEventListener('error', (event) => {
  logger.error('Global error caught', {
    message: event.error?.message || event.message,
    stack: event.error?.stack || 'No stack trace',
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno,
    category: 'global_error',
  })
})

window.addEventListener('unhandledrejection', (event) => {
  logger.error('Unhandled promise rejection', {
    reason: event.reason,
    category: 'unhandled_rejection',
  })
})
```

## Directory Structure
```
src/
├── utils/
│   ├── monitoring.js         # Main monitoring manager
│   ├── logger.js            # Logging system
│   └── alerting.js          # Alerting utilities
├── components/
│   ├── monitoring/
│   │   ├── MonitoringDashboard.jsx
│   │   ├── AlertCard.jsx
│   │   ├── MetricCard.jsx
│   │   ├── PerformanceChart.jsx
│   │   └── MonitoredErrorBoundary.jsx
│   └── common/
│       └── HealthCheck.jsx
├── services/
│   ├── monitoringService.js  # API communication
│   ├── alertingService.js    # Alert management
│   └── metricsService.js     # Metrics collection
├── hooks/
│   ├── useMonitoring.js      # Monitoring hook
│   ├── useLogger.js          # Logging hook
│   └── useHealthCheck.js     # Health check hook
└── config/
    ├── monitoring.js         # Monitoring configuration
    ├── logging.js           # Logging configuration
    └── alerts.js            # Alert configuration
```

## Benefits
- Real-time application monitoring
- Proactive issue detection
- Performance optimization insights
- User experience tracking
- Error tracking and debugging
- Infrastructure monitoring
- Automated alerting system
- Historical data analysis
- Compliance and auditing
- Incident response automation
- Business metrics visibility
- Service level monitoring

## Monitoring Features Implemented
- ✅ Application Performance Monitoring (APM)
- ✅ Error tracking and reporting
- ✅ Real-time dashboards
- ✅ Automated alerting system
- ✅ Comprehensive logging
- ✅ User experience monitoring
- ✅ Infrastructure monitoring
- ✅ Performance metrics
- ✅ Business metrics tracking
- ✅ Health checks
- ✅ Incident management
- ✅ Observability tools

## Key Metrics Monitored
- Response times and latency
- Error rates and exceptions
- User engagement metrics
- Core Web Vitals
- Memory and resource usage
- API performance
- Business KPIs
- Infrastructure health
- Security events
- Compliance metrics