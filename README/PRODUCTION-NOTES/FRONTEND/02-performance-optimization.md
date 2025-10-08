# Performance Optimization Implementation Plan

## Overview
Implement comprehensive performance optimizations to ensure fast loading times, smooth user interactions, and efficient resource utilization across all devices and network conditions.

## Current State Analysis
- Basic React application with minimal optimizations
- No code splitting or lazy loading
- Large bundle size with all dependencies loaded upfront
- No image optimization or lazy loading
- No caching strategies implemented
- Basic Vite configuration without advanced optimizations
- No performance monitoring or metrics collection

## Implementation Steps

### Step 1: Bundle Analysis and Code Splitting
**Timeline: 2-3 days**

Analyze current bundle size and implement strategic code splitting:

```javascript
// vite.config.mjs - Enhanced configuration
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { visualizer } from 'rollup-plugin-visualizer'
import { splitVendorChunkPlugin } from 'vite'

export default defineConfig({
  plugins: [
    react(),
    splitVendorChunkPlugin(),
    visualizer({
      filename: 'dist/stats.html',
      open: true,
      gzipSize: true,
      brotliSize: true,
    }),
  ],
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Vendor chunks
          react: ['react', 'react-dom'],
          router: ['react-router-dom'],
          charts: ['chart.js', 'react-chartjs-2', 'recharts', 'd3'],
          ui: ['@headlessui/react', '@heroicons/react'],
          // App chunks
          auth: ['./src/components/auth/AuthContext.jsx'],
          markets: ['./src/pages/Markets', './src/components/markets'],
          admin: ['./src/pages/admin', './src/components/admin'],
        },
      },
    },
    chunkSizeWarningLimit: 1000,
    sourcemap: false, // Disable in production
  },
  optimizeDeps: {
    include: ['react', 'react-dom', 'react-router-dom'],
    exclude: ['@vite/client', '@vite/env'],
  },
})

// Bundle analyzer script
// package.json
{
  "scripts": {
    "analyze": "npm run build && npx vite-bundle-analyzer dist",
    "build:analyze": "vite build --mode analyze"
  }
}
```

Implement route-based code splitting:

```javascript
// helpers/AppRoutes.jsx - Lazy loaded routes
import React, { Suspense } from 'react'
import { Routes, Route } from 'react-router-dom'
import LoadingSpinner from '../components/common/LoadingSpinner'
import ErrorBoundary from '../components/common/ErrorBoundary'

// Lazy load components
const HomePage = React.lazy(() => import('../pages/HomePage'))
const Markets = React.lazy(() => import('../pages/Markets'))
const MarketDetail = React.lazy(() => import('../pages/MarketDetail'))
const Profile = React.lazy(() => 
  import('../pages/Profile').then(module => ({
    default: module.Profile
  }))
)
const AdminDashboard = React.lazy(() => 
  import('../pages/admin/AdminDashboard')
)
const StatsPage = React.lazy(() => import('../pages/StatsPage'))

// Preload critical routes
const preloadRoutes = {
  markets: () => import('../pages/Markets'),
  profile: () => import('../pages/Profile'),
}

// Preload on user interaction
export const preloadRoute = (routeName) => {
  if (preloadRoutes[routeName]) {
    preloadRoutes[routeName]()
  }
}

const LazyWrapper = ({ children }) => (
  <ErrorBoundary>
    <Suspense fallback={<LoadingSpinner />}>
      {children}
    </Suspense>
  </ErrorBoundary>
)

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={
        <LazyWrapper>
          <HomePage />
        </LazyWrapper>
      } />
      <Route path="/markets" element={
        <LazyWrapper>
          <Markets />
        </LazyWrapper>
      } />
      <Route path="/markets/:id" element={
        <LazyWrapper>
          <MarketDetail />
        </LazyWrapper>
      } />
      <Route path="/profile" element={
        <LazyWrapper>
          <Profile />
        </LazyWrapper>
      } />
      <Route path="/admin/*" element={
        <LazyWrapper>
          <AdminDashboard />
        </LazyWrapper>
      } />
      <Route path="/stats" element={
        <LazyWrapper>
          <StatsPage />
        </LazyWrapper>
      } />
    </Routes>
  )
}

export default AppRoutes
```

### Step 2: React Performance Optimizations
**Timeline: 2-3 days**

Implement React-specific performance optimizations:

```javascript
// hooks/useOptimizedCallback.js
import { useCallback, useMemo, useRef } from 'react'

// Stable callback hook
export const useStableCallback = (callback) => {
  const callbackRef = useRef(callback)
  callbackRef.current = callback
  
  return useCallback((...args) => {
    return callbackRef.current(...args)
  }, [])
}

// Debounced callback hook
export const useDebouncedCallback = (callback, delay) => {
  const timeoutRef = useRef()
  
  return useCallback((...args) => {
    clearTimeout(timeoutRef.current)
    timeoutRef.current = setTimeout(() => callback(...args), delay)
  }, [callback, delay])
}

// Throttled callback hook
export const useThrottledCallback = (callback, delay) => {
  const lastRun = useRef(Date.now())
  
  return useCallback((...args) => {
    if (Date.now() - lastRun.current >= delay) {
      callback(...args)
      lastRun.current = Date.now()
    }
  }, [callback, delay])
}
```

Optimize components with memoization:

```javascript
// components/markets/MarketCard.jsx - Optimized component
import React, { memo, useMemo } from 'react'
import { useAppSelector } from '../../hooks/redux'
import { selectMarketById } from '../../store/slices/marketsSlice'

const MarketCard = memo(({ marketId, onBetClick, className }) => {
  const market = useAppSelector(state => selectMarketById(state, marketId))
  
  // Memoize expensive calculations
  const marketStats = useMemo(() => {
    if (!market) return null
    
    return {
      totalVolume: market.bets?.reduce((sum, bet) => sum + bet.amount, 0) || 0,
      participantCount: new Set(market.bets?.map(bet => bet.userId)).size || 0,
      timeRemaining: Math.max(0, new Date(market.closingDate) - new Date()),
      winningProbability: calculateWinningProbability(market),
    }
  }, [market])
  
  // Memoize handlers
  const handleBetClick = useStableCallback(() => {
    onBetClick(market.id)
  })
  
  if (!market) return null
  
  return (
    <div className={`market-card ${className}`}>
      <h3>{market.title}</h3>
      <div className="market-stats">
        <span>Volume: ${marketStats.totalVolume}</span>
        <span>Participants: {marketStats.participantCount}</span>
        <span>Time Left: {formatTimeRemaining(marketStats.timeRemaining)}</span>
      </div>
      <button onClick={handleBetClick}>
        Place Bet ({(marketStats.winningProbability * 100).toFixed(1)}%)
      </button>
    </div>
  )
}, (prevProps, nextProps) => {
  // Custom comparison for shallow equality
  return (
    prevProps.marketId === nextProps.marketId &&
    prevProps.className === nextProps.className &&
    prevProps.onBetClick === nextProps.onBetClick
  )
})

MarketCard.displayName = 'MarketCard'
export default MarketCard
```

Implement virtualization for large lists:

```javascript
// components/markets/VirtualizedMarketList.jsx
import React, { useMemo } from 'react'
import { FixedSizeList as List } from 'react-window'
import { useAppSelector } from '../../hooks/redux'
import { selectAllMarkets } from '../../store/slices/marketsSlice'
import MarketCard from './MarketCard'

const VirtualizedMarketList = ({ height = 600, itemHeight = 200 }) => {
  const markets = useAppSelector(selectAllMarkets)
  
  const itemCount = markets.length
  
  const ItemRenderer = useMemo(() => 
    ({ index, style }) => {
      const market = markets[index]
      return (
        <div style={style}>
          <MarketCard marketId={market.id} />
        </div>
      )
    }, [markets]
  )
  
  return (
    <List
      height={height}
      itemCount={itemCount}
      itemSize={itemHeight}
      overscanCount={5}
    >
      {ItemRenderer}
    </List>
  )
}

export default VirtualizedMarketList
```

### Step 3: Image and Asset Optimization
**Timeline: 2 days**

Implement comprehensive asset optimization:

```javascript
// components/common/OptimizedImage.jsx
import React, { useState, useRef, useEffect } from 'react'
import { useIntersection } from '../../hooks/useIntersection'

const OptimizedImage = ({
  src,
  alt,
  width,
  height,
  className,
  loading = 'lazy',
  placeholder = '/images/placeholder.svg',
  sizes,
  quality = 75,
  format = 'webp',
}) => {
  const [isLoaded, setIsLoaded] = useState(false)
  const [hasError, setHasError] = useState(false)
  const [currentSrc, setCurrentSrc] = useState(placeholder)
  const imgRef = useRef()
  
  const isIntersecting = useIntersection(imgRef, {
    threshold: 0.1,
    rootMargin: '50px',
  })
  
  // Generate responsive image URLs
  const generateSrcSet = (baseSrc) => {
    const breakpoints = [320, 640, 768, 1024, 1280, 1536]
    return breakpoints
      .map(bp => `${baseSrc}?w=${bp}&q=${quality}&f=${format} ${bp}w`)
      .join(', ')
  }
  
  useEffect(() => {
    if (!isIntersecting) return
    
    const img = new Image()
    img.onload = () => {
      setCurrentSrc(src)
      setIsLoaded(true)
    }
    img.onerror = () => {
      setHasError(true)
    }
    img.src = src
  }, [isIntersecting, src])
  
  if (hasError) {
    return (
      <div className={`error-placeholder ${className}`}>
        <span>Failed to load image</span>
      </div>
    )
  }
  
  return (
    <img
      ref={imgRef}
      src={currentSrc}
      srcSet={generateSrcSet(src)}
      sizes={sizes}
      alt={alt}
      width={width}
      height={height}
      loading={loading}
      className={`${className} ${isLoaded ? 'loaded' : 'loading'}`}
      style={{
        transition: 'opacity 0.3s ease',
        opacity: isLoaded ? 1 : 0.7,
      }}
    />
  )
}

export default OptimizedImage

// hooks/useIntersection.js
import { useState, useEffect } from 'react'

export const useIntersection = (elementRef, options) => {
  const [isIntersecting, setIsIntersecting] = useState(false)
  
  useEffect(() => {
    if (!elementRef.current) return
    
    const observer = new IntersectionObserver(([entry]) => {
      setIsIntersecting(entry.isIntersecting)
      if (entry.isIntersecting) {
        observer.disconnect()
      }
    }, options)
    
    observer.observe(elementRef.current)
    
    return () => observer.disconnect()
  }, [elementRef, options])
  
  return isIntersecting
}
```

### Step 4: Caching Strategies
**Timeline: 2-3 days**

Implement multi-level caching:

```javascript
// utils/cacheManager.js
class CacheManager {
  constructor() {
    this.memoryCache = new Map()
    this.maxMemorySize = 100 // Maximum items in memory
    this.ttl = 5 * 60 * 1000 // 5 minutes default TTL
  }
  
  // Memory cache with LRU eviction
  setMemory(key, value, ttl = this.ttl) {
    if (this.memoryCache.size >= this.maxMemorySize) {
      const firstKey = this.memoryCache.keys().next().value
      this.memoryCache.delete(firstKey)
    }
    
    this.memoryCache.set(key, {
      value,
      timestamp: Date.now(),
      ttl,
    })
  }
  
  getMemory(key) {
    const item = this.memoryCache.get(key)
    if (!item) return null
    
    if (Date.now() - item.timestamp > item.ttl) {
      this.memoryCache.delete(key)
      return null
    }
    
    // Move to end (LRU)
    this.memoryCache.delete(key)
    this.memoryCache.set(key, item)
    
    return item.value
  }
  
  // IndexedDB cache for persistent storage
  async setIndexedDB(key, value, ttl = this.ttl) {
    try {
      const db = await this.openDB()
      const transaction = db.transaction(['cache'], 'readwrite')
      const store = transaction.objectStore('cache')
      
      await store.put({
        key,
        value,
        timestamp: Date.now(),
        ttl,
      })
    } catch (error) {
      console.warn('IndexedDB cache write failed:', error)
    }
  }
  
  async getIndexedDB(key) {
    try {
      const db = await this.openDB()
      const transaction = db.transaction(['cache'], 'readonly')
      const store = transaction.objectStore('cache')
      const item = await store.get(key)
      
      if (!item) return null
      
      if (Date.now() - item.timestamp > item.ttl) {
        await this.deleteIndexedDB(key)
        return null
      }
      
      return item.value
    } catch (error) {
      console.warn('IndexedDB cache read failed:', error)
      return null
    }
  }
  
  async openDB() {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open('SocialPredictCache', 1)
      
      request.onerror = () => reject(request.error)
      request.onsuccess = () => resolve(request.result)
      
      request.onupgradeneeded = (event) => {
        const db = event.target.result
        if (!db.objectStoreNames.contains('cache')) {
          db.createObjectStore('cache', { keyPath: 'key' })
        }
      }
    })
  }
  
  // Smart cache strategy
  async get(key, fetchFn, options = {}) {
    const { memoryOnly = false, ttl = this.ttl } = options
    
    // Try memory cache first
    let value = this.getMemory(key)
    if (value) return value
    
    // Try IndexedDB if not memory-only
    if (!memoryOnly) {
      value = await this.getIndexedDB(key)
      if (value) {
        // Populate memory cache
        this.setMemory(key, value, ttl)
        return value
      }
    }
    
    // Fetch fresh data
    if (fetchFn) {
      try {
        value = await fetchFn()
        this.setMemory(key, value, ttl)
        if (!memoryOnly) {
          this.setIndexedDB(key, value, ttl)
        }
        return value
      } catch (error) {
        console.error('Cache fetch failed:', error)
        throw error
      }
    }
    
    return null
  }
  
  // Cache invalidation
  invalidate(pattern) {
    const regex = new RegExp(pattern)
    
    // Clear memory cache
    for (const key of this.memoryCache.keys()) {
      if (regex.test(key)) {
        this.memoryCache.delete(key)
      }
    }
    
    // Clear IndexedDB cache
    this.clearIndexedDBPattern(pattern)
  }
}

export const cacheManager = new CacheManager()

// services/apiCache.js
import { cacheManager } from '../utils/cacheManager'

export class APICache {
  static async getMarkets(params = {}) {
    const cacheKey = `markets:${JSON.stringify(params)}`
    
    return cacheManager.get(
      cacheKey,
      () => fetch(`/api/v0/markets?${new URLSearchParams(params)}`).then(r => r.json()),
      { ttl: 2 * 60 * 1000 } // 2 minutes for markets
    )
  }
  
  static async getMarket(id) {
    const cacheKey = `market:${id}`
    
    return cacheManager.get(
      cacheKey,
      () => fetch(`/api/v0/markets/${id}`).then(r => r.json()),
      { ttl: 30 * 1000 } // 30 seconds for individual market
    )
  }
  
  static async getUserProfile() {
    const cacheKey = 'user:profile'
    
    return cacheManager.get(
      cacheKey,
      () => fetch('/api/v0/users/profile').then(r => r.json()),
      { ttl: 5 * 60 * 1000 } // 5 minutes for user profile
    )
  }
  
  static invalidateMarkets() {
    cacheManager.invalidate('^markets:')
  }
  
  static invalidateMarket(id) {
    cacheManager.invalidate(`^market:${id}$`)
  }
}
```

### Step 5: Network Optimization
**Timeline: 2 days**

Implement network-level optimizations:

```javascript
// utils/networkOptimization.js
export class NetworkOptimizer {
  constructor() {
    this.connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection
    this.isOnline = navigator.onLine
    this.networkType = this.getNetworkType()
    
    this.setupNetworkListeners()
  }
  
  getNetworkType() {
    if (!this.connection) return 'unknown'
    
    const { effectiveType, downlink } = this.connection
    
    if (effectiveType === '4g' && downlink > 10) return 'fast'
    if (effectiveType === '4g' || effectiveType === '3g') return 'medium'
    return 'slow'
  }
  
  setupNetworkListeners() {
    window.addEventListener('online', () => {
      this.isOnline = true
      this.onNetworkChange('online')
    })
    
    window.addEventListener('offline', () => {
      this.isOnline = false
      this.onNetworkChange('offline')
    })
    
    if (this.connection) {
      this.connection.addEventListener('change', () => {
        this.networkType = this.getNetworkType()
        this.onNetworkChange('speed')
      })
    }
  }
  
  onNetworkChange(type) {
    // Emit custom events for network changes
    window.dispatchEvent(new CustomEvent('networkChange', {
      detail: { type, isOnline: this.isOnline, networkType: this.networkType }
    }))
  }
  
  // Adaptive loading based on network conditions
  getOptimalChunkSize() {
    switch (this.networkType) {
      case 'fast': return 50
      case 'medium': return 20
      case 'slow': return 10
      default: return 20
    }
  }
  
  getOptimalImageQuality() {
    switch (this.networkType) {
      case 'fast': return 85
      case 'medium': return 70
      case 'slow': return 50
      default: return 70
    }
  }
  
  shouldPreloadResources() {
    return this.networkType === 'fast' && this.isOnline
  }
}

export const networkOptimizer = new NetworkOptimizer()

// hooks/useNetworkOptimized.js
import { useState, useEffect } from 'react'
import { networkOptimizer } from '../utils/networkOptimization'

export const useNetworkOptimized = () => {
  const [networkState, setNetworkState] = useState({
    isOnline: networkOptimizer.isOnline,
    networkType: networkOptimizer.networkType,
  })
  
  useEffect(() => {
    const handleNetworkChange = (event) => {
      setNetworkState({
        isOnline: event.detail.isOnline,
        networkType: event.detail.networkType,
      })
    }
    
    window.addEventListener('networkChange', handleNetworkChange)
    return () => window.removeEventListener('networkChange', handleNetworkChange)
  }, [])
  
  return {
    ...networkState,
    chunkSize: networkOptimizer.getOptimalChunkSize(),
    imageQuality: networkOptimizer.getOptimalImageQuality(),
    shouldPreload: networkOptimizer.shouldPreloadResources(),
  }
}
```

### Step 6: Performance Monitoring
**Timeline: 2 days**

Implement comprehensive performance monitoring:

```javascript
// utils/performanceMonitor.js
export class PerformanceMonitor {
  constructor() {
    this.metrics = new Map()
    this.observers = new Map()
    this.initialized = false
    
    this.init()
  }
  
  init() {
    if (this.initialized) return
    
    // Web Vitals monitoring
    this.observeWebVitals()
    
    // Custom performance marks
    this.observeCustomMetrics()
    
    // Resource loading monitoring
    this.observeResourceLoading()
    
    // Long task monitoring
    this.observeLongTasks()
    
    this.initialized = true
  }
  
  observeWebVitals() {
    // Largest Contentful Paint
    if ('PerformanceObserver' in window) {
      const lcpObserver = new PerformanceObserver((list) => {
        const entries = list.getEntries()
        const lastEntry = entries[entries.length - 1]
        this.recordMetric('LCP', lastEntry.startTime)
      })
      lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] })
      this.observers.set('lcp', lcpObserver)
    }
    
    // First Input Delay
    const fidObserver = new PerformanceObserver((list) => {
      const entries = list.getEntries()
      entries.forEach(entry => {
        this.recordMetric('FID', entry.processingStart - entry.startTime)
      })
    })
    fidObserver.observe({ entryTypes: ['first-input'] })
    this.observers.set('fid', fidObserver)
    
    // Cumulative Layout Shift
    let clsValue = 0
    const clsObserver = new PerformanceObserver((list) => {
      const entries = list.getEntries()
      entries.forEach(entry => {
        if (!entry.hadRecentInput) {
          clsValue += entry.value
        }
      })
      this.recordMetric('CLS', clsValue)
    })
    clsObserver.observe({ entryTypes: ['layout-shift'] })
    this.observers.set('cls', clsObserver)
  }
  
  observeCustomMetrics() {
    // Time to Interactive
    this.measureTTI()
    
    // First Meaningful Paint
    this.measureFMP()
    
    // Bundle loading time
    this.measureBundleLoading()
  }
  
  observeResourceLoading() {
    const resourceObserver = new PerformanceObserver((list) => {
      const entries = list.getEntries()
      entries.forEach(entry => {
        if (entry.name.includes('chunk') || entry.name.includes('bundle')) {
          this.recordMetric('ChunkLoadTime', entry.duration, {
            name: entry.name,
            size: entry.transferSize,
          })
        }
      })
    })
    resourceObserver.observe({ entryTypes: ['resource'] })
    this.observers.set('resource', resourceObserver)
  }
  
  observeLongTasks() {
    if ('PerformanceObserver' in window) {
      const longTaskObserver = new PerformanceObserver((list) => {
        const entries = list.getEntries()
        entries.forEach(entry => {
          this.recordMetric('LongTask', entry.duration, {
            startTime: entry.startTime,
            attribution: entry.attribution,
          })
        })
      })
      
      try {
        longTaskObserver.observe({ entryTypes: ['longtask'] })
        this.observers.set('longtask', longTaskObserver)
      } catch (e) {
        console.warn('Long task monitoring not supported')
      }
    }
  }
  
  recordMetric(name, value, metadata = {}) {
    if (!this.metrics.has(name)) {
      this.metrics.set(name, [])
    }
    
    this.metrics.get(name).push({
      value,
      timestamp: Date.now(),
      url: window.location.pathname,
      userAgent: navigator.userAgent,
      ...metadata,
    })
    
    // Send to analytics if value exceeds thresholds
    this.checkThresholds(name, value)
  }
  
  checkThresholds(name, value) {
    const thresholds = {
      LCP: 2500, // 2.5s
      FID: 100, // 100ms
      CLS: 0.1, // 0.1
      LongTask: 50, // 50ms
      ChunkLoadTime: 3000, // 3s
    }
    
    if (thresholds[name] && value > thresholds[name]) {
      this.reportPerformanceIssue(name, value, thresholds[name])
    }
  }
  
  reportPerformanceIssue(metric, value, threshold) {
    // Report to analytics service
    console.warn(`Performance issue: ${metric} = ${value}ms (threshold: ${threshold}ms)`)
    
    // Send to monitoring service
    if (window.gtag) {
      window.gtag('event', 'performance_issue', {
        metric_name: metric,
        metric_value: value,
        threshold: threshold,
        page_path: window.location.pathname,
      })
    }
  }
  
  getMetrics(name) {
    return this.metrics.get(name) || []
  }
  
  getAllMetrics() {
    const result = {}
    for (const [name, values] of this.metrics) {
      result[name] = values
    }
    return result
  }
  
  // Custom timing for React operations
  startTiming(name) {
    performance.mark(`${name}-start`)
  }
  
  endTiming(name) {
    performance.mark(`${name}-end`)
    performance.measure(name, `${name}-start`, `${name}-end`)
    
    const measure = performance.getEntriesByName(name, 'measure')[0]
    if (measure) {
      this.recordMetric(name, measure.duration)
    }
  }
  
  // React component performance tracking
  trackComponentRender(componentName, renderTime) {
    this.recordMetric('ComponentRender', renderTime, { component: componentName })
  }
  
  // API call performance
  trackAPICall(endpoint, duration, success) {
    this.recordMetric('APICall', duration, { endpoint, success })
  }
}

export const performanceMonitor = new PerformanceMonitor()

// hooks/usePerformanceMonitor.js
import { useEffect } from 'react'
import { performanceMonitor } from '../utils/performanceMonitor'

export const usePerformanceMonitor = (componentName) => {
  useEffect(() => {
    const startTime = performance.now()
    
    return () => {
      const endTime = performance.now()
      performanceMonitor.trackComponentRender(componentName, endTime - startTime)
    }
  }, [componentName])
}

// HOC for component performance monitoring
export const withPerformanceMonitoring = (WrappedComponent) => {
  return function PerformanceMonitoredComponent(props) {
    const componentName = WrappedComponent.displayName || WrappedComponent.name
    usePerformanceMonitor(componentName)
    
    return <WrappedComponent {...props} />
  }
}
```

## Directory Structure
```
src/
├── utils/
│   ├── cacheManager.js          # Multi-level caching
│   ├── performanceMonitor.js    # Performance monitoring
│   ├── networkOptimization.js   # Network-aware optimizations
│   ├── bundleAnalyzer.js        # Bundle analysis utilities
│   └── lazyLoading.js           # Lazy loading utilities
├── hooks/
│   ├── useIntersection.js       # Intersection Observer hook
│   ├── useNetworkOptimized.js   # Network-aware hook
│   ├── usePerformanceMonitor.js # Performance monitoring hook
│   ├── useDebouncedCallback.js  # Debounced callbacks
│   └── useVirtualization.js     # Virtualization hook
├── components/
│   ├── common/
│   │   ├── OptimizedImage.jsx   # Optimized image component
│   │   ├── VirtualizedList.jsx  # Virtualized list component
│   │   └── LazyWrapper.jsx      # Lazy loading wrapper
│   └── performance/
│       ├── PerformanceProvider.jsx # Performance context
│       └── PerformanceReporter.jsx # Performance reporting
└── services/
    ├── apiCache.js              # API caching service
    └── preloadManager.js        # Resource preloading
```

## Performance Metrics to Track
- **Core Web Vitals**: LCP, FID, CLS
- **Custom Metrics**: TTI, FMP, Bundle Load Time
- **Component Metrics**: Render time, Re-render count
- **Network Metrics**: API response time, Cache hit rate
- **User Metrics**: Time to interaction, Error rate

## Benefits
- Faster initial page load
- Smoother user interactions
- Reduced bandwidth usage
- Better Core Web Vitals scores
- Improved user experience on slow networks
- Reduced server load through caching
- Better mobile performance
- Enhanced SEO rankings

## Performance Budget
- Initial bundle size: < 250KB (gzipped)
- Largest Contentful Paint: < 2.5s
- First Input Delay: < 100ms
- Cumulative Layout Shift: < 0.1
- Time to Interactive: < 3.5s
- Chunk load time: < 3s
- API response time: < 500ms