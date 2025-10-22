# Progressive Web App (PWA) Implementation Plan

## Overview
Transform the application into a Progressive Web App to provide native app-like experiences, offline functionality, push notifications, and enhanced user engagement across all devices.

## Current State Analysis
- Standard web application without PWA features
- No service worker implementation
- No offline functionality
- No push notifications
- No app-like installation experience
- Limited mobile optimization
- No background sync capabilities
- No native device integration

## Implementation Steps

### Step 1: Service Worker Setup and Caching Strategy
**Timeline: 3-4 days**

Implement comprehensive service worker with advanced caching strategies:

```javascript
// public/sw.js - Service Worker
const CACHE_NAME = 'socialpredict-v1.0.0'
const STATIC_CACHE = 'static-v1.0.0'
const DYNAMIC_CACHE = 'dynamic-v1.0.0'
const API_CACHE = 'api-v1.0.0'

// Define cache strategies for different resource types
const CACHE_STRATEGIES = {
  // Static assets - Cache First
  static: [
    '/',
    '/static/js/',
    '/static/css/',
    '/static/media/',
    '/manifest.json',
    '/favicon.ico',
  ],
  
  // API endpoints - Network First with fallback
  api: [
    '/api/markets',
    '/api/users',
    '/api/auth',
  ],
  
  // Dynamic content - Stale While Revalidate
  dynamic: [
    '/markets/',
    '/profile/',
    '/admin/',
  ],
}

// Install event - Pre-cache static assets
self.addEventListener('install', (event) => {
  console.log('[SW] Installing service worker')
  
  event.waitUntil(
    Promise.all([
      // Cache static assets
      caches.open(STATIC_CACHE).then((cache) => {
        return cache.addAll([
          '/',
          '/static/js/main.js',
          '/static/css/main.css',
          '/manifest.json',
          '/offline.html',
          '/favicon.ico',
        ])
      }),
      
      // Skip waiting to activate immediately
      self.skipWaiting(),
    ])
  )
})

// Activate event - Clean up old caches
self.addEventListener('activate', (event) => {
  console.log('[SW] Activating service worker')
  
  event.waitUntil(
    Promise.all([
      // Clean up old caches
      caches.keys().then((cacheNames) => {
        return Promise.all(
          cacheNames.map((cacheName) => {
            if (cacheName !== CACHE_NAME && 
                cacheName !== STATIC_CACHE && 
                cacheName !== DYNAMIC_CACHE &&
                cacheName !== API_CACHE) {
              console.log('[SW] Deleting old cache:', cacheName)
              return caches.delete(cacheName)
            }
          })
        )
      }),
      
      // Take control of all clients
      self.clients.claim(),
    ])
  )
})

// Fetch event - Implement caching strategies
self.addEventListener('fetch', (event) => {
  const { request } = event
  const url = new URL(request.url)
  
  // Skip non-GET requests
  if (request.method !== 'GET') {
    return
  }
  
  // Handle different request types
  if (url.pathname.startsWith('/api/')) {
    // API requests - Network First
    event.respondWith(handleAPIRequest(request))
  } else if (isStaticAsset(url.pathname)) {
    // Static assets - Cache First
    event.respondWith(handleStaticRequest(request))
  } else {
    // Dynamic content - Stale While Revalidate
    event.respondWith(handleDynamicRequest(request))
  }
})

// Cache First strategy for static assets
async function handleStaticRequest(request) {
  try {
    const cachedResponse = await caches.match(request)
    if (cachedResponse) {
      return cachedResponse
    }
    
    const networkResponse = await fetch(request)
    const cache = await caches.open(STATIC_CACHE)
    cache.put(request, networkResponse.clone())
    
    return networkResponse
  } catch (error) {
    console.error('[SW] Static request failed:', error)
    return new Response('Offline', { status: 503 })
  }
}

// Network First strategy for API requests
async function handleAPIRequest(request) {
  try {
    const networkResponse = await fetch(request)
    
    // Cache successful responses
    if (networkResponse.ok) {
      const cache = await caches.open(API_CACHE)
      cache.put(request, networkResponse.clone())
    }
    
    return networkResponse
  } catch (error) {
    console.log('[SW] Network failed, trying cache:', request.url)
    
    const cachedResponse = await caches.match(request)
    if (cachedResponse) {
      return cachedResponse
    }
    
    // Return offline response for critical API endpoints
    if (request.url.includes('/api/auth')) {
      return new Response(JSON.stringify({ 
        error: 'Offline',
        message: 'Authentication requires internet connection' 
      }), {
        status: 503,
        headers: { 'Content-Type': 'application/json' }
      })
    }
    
    return new Response(JSON.stringify({ 
      error: 'Offline',
      message: 'This feature requires internet connection' 
    }), {
      status: 503,
      headers: { 'Content-Type': 'application/json' }
    })
  }
}

// Stale While Revalidate strategy for dynamic content
async function handleDynamicRequest(request) {
  const cache = await caches.open(DYNAMIC_CACHE)
  const cachedResponse = await cache.match(request)
  
  const fetchPromise = fetch(request).then((networkResponse) => {
    if (networkResponse.ok) {
      cache.put(request, networkResponse.clone())
    }
    return networkResponse
  }).catch(() => {
    // Return offline page for navigation requests
    if (request.mode === 'navigate') {
      return caches.match('/offline.html')
    }
    throw new Error('Network failed and no cache available')
  })
  
  // Return cached response immediately if available
  return cachedResponse || fetchPromise
}

// Background sync for offline actions
self.addEventListener('sync', (event) => {
  console.log('[SW] Background sync triggered:', event.tag)
  
  if (event.tag === 'background-sync-bets') {
    event.waitUntil(syncOfflineBets())
  } else if (event.tag === 'background-sync-profile') {
    event.waitUntil(syncProfileUpdates())
  }
})

// Sync offline bet data
async function syncOfflineBets() {
  try {
    const offlineBets = await getStoredOfflineData('offline-bets')
    
    for (const bet of offlineBets) {
      try {
        const response = await fetch('/api/bets', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${bet.token}`,
          },
          body: JSON.stringify(bet.data),
        })
        
        if (response.ok) {
          // Remove successfully synced bet
          await removeOfflineData('offline-bets', bet.id)
          
          // Notify client of successful sync
          notifyClients({
            type: 'SYNC_SUCCESS',
            data: { type: 'bet', id: bet.id },
          })
        }
      } catch (error) {
        console.error('[SW] Failed to sync bet:', error)
      }
    }
  } catch (error) {
    console.error('[SW] Background sync failed:', error)
  }
}

// Push notification handling
self.addEventListener('push', (event) => {
  console.log('[SW] Push received:', event.data?.text())
  
  if (!event.data) return
  
  const data = event.data.json()
  const options = {
    body: data.body,
    icon: '/icons/icon-192x192.png',
    badge: '/icons/badge-72x72.png',
    image: data.image,
    tag: data.tag || 'socialpredict-notification',
    data: data.data,
    actions: data.actions || [],
    requireInteraction: data.urgent || false,
    silent: false,
    vibrate: [200, 100, 200],
  }
  
  event.waitUntil(
    self.registration.showNotification(data.title, options)
  )
})

// Notification click handling
self.addEventListener('notificationclick', (event) => {
  console.log('[SW] Notification clicked:', event.notification.data)
  
  event.notification.close()
  
  const data = event.notification.data
  let url = '/'
  
  // Handle different notification types
  if (data.type === 'market-update') {
    url = `/markets/${data.marketId}`
  } else if (data.type === 'bet-result') {
    url = `/profile/bets/${data.betId}`
  } else if (data.type === 'market-closing') {
    url = `/markets/${data.marketId}`
  }
  
  event.waitUntil(
    clients.matchAll({ type: 'window' }).then((clientList) => {
      // Check if app is already open
      for (const client of clientList) {
        if (client.url === url && 'focus' in client) {
          return client.focus()
        }
      }
      
      // Open new window if app is not open
      if (clients.openWindow) {
        return clients.openWindow(url)
      }
    })
  )
})

// Utility functions
function isStaticAsset(pathname) {
  return pathname.startsWith('/static/') || 
         pathname.includes('.js') || 
         pathname.includes('.css') || 
         pathname.includes('.png') || 
         pathname.includes('.jpg') || 
         pathname.includes('.svg') ||
         pathname === '/manifest.json' ||
         pathname === '/favicon.ico'
}

async function getStoredOfflineData(storeName) {
  // Implementation would use IndexedDB
  return []
}

async function removeOfflineData(storeName, id) {
  // Implementation would use IndexedDB
}

function notifyClients(message) {
  self.clients.matchAll().then((clients) => {
    clients.forEach((client) => {
      client.postMessage(message)
    })
  })
}

// vite.config.mjs - PWA Plugin Configuration
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({
  plugins: [
    react(),
    VitePWA({
      registerType: 'autoUpdate',
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,json,vue,txt,woff2}'],
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/api\.socialpredict\.com\/api\/markets/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-markets',
              expiration: {
                maxEntries: 100,
                maxAgeSeconds: 60 * 60 * 24, // 24 hours
              },
              cacheKeyWillBeUsed: async ({ request }) => {
                return `${request.url}?${Date.now()}`
              },
            },
          },
          {
            urlPattern: /^https:\/\/api\.socialpredict\.com\/api\/users/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-users',
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 60 * 60, // 1 hour
              },
            },
          },
          {
            urlPattern: /\.(?:png|jpg|jpeg|svg|gif)$/,
            handler: 'CacheFirst',
            options: {
              cacheName: 'images',
              expiration: {
                maxEntries: 1000,
                maxAgeSeconds: 60 * 60 * 24 * 30, // 30 days
              },
            },
          },
        ],
      },
      includeAssets: ['favicon.ico', 'apple-touch-icon.png', 'masked-icon.svg'],
      manifest: {
        name: 'SocialPredict',
        short_name: 'SocialPredict',
        description: 'Social prediction markets platform',
        theme_color: '#4F46E5',
        background_color: '#FFFFFF',
        display: 'standalone',
        orientation: 'portrait-primary',
        start_url: '/',
        scope: '/',
        categories: ['finance', 'social', 'entertainment'],
        lang: 'en',
        dir: 'ltr',
        icons: [
          {
            src: 'icons/icon-72x72.png',
            sizes: '72x72',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: 'icons/icon-96x96.png',
            sizes: '96x96',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: 'icons/icon-128x128.png',
            sizes: '128x128',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: 'icons/icon-144x144.png',
            sizes: '144x144',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: 'icons/icon-152x152.png',
            sizes: '152x152',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: 'icons/icon-192x192.png',
            sizes: '192x192',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: 'icons/icon-384x384.png',
            sizes: '384x384',
            type: 'image/png',
            purpose: 'maskable any'
          },
          {
            src: 'icons/icon-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable any'
          }
        ],
        shortcuts: [
          {
            name: 'Browse Markets',
            short_name: 'Markets',
            description: 'Browse available prediction markets',
            url: '/markets',
            icons: [{ src: 'icons/market-96x96.png', sizes: '96x96' }]
          },
          {
            name: 'My Profile',
            short_name: 'Profile',
            description: 'View your profile and betting history',
            url: '/profile',
            icons: [{ src: 'icons/profile-96x96.png', sizes: '96x96' }]
          }
        ]
      }
    })
  ],
})
```

### Step 2: Offline Functionality and Data Synchronization
**Timeline: 4-5 days**

Implement comprehensive offline capabilities:

```javascript
// hooks/useOfflineStorage.js
import { useState, useEffect, useCallback } from 'react'

const OFFLINE_STORAGE_KEY = 'socialpredict-offline'

export const useOfflineStorage = () => {
  const [isOnline, setIsOnline] = useState(navigator.onLine)
  const [offlineQueue, setOfflineQueue] = useState([])

  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true)
      syncOfflineData()
    }

    const handleOffline = () => {
      setIsOnline(false)
    }

    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    // Load offline queue from localStorage
    loadOfflineQueue()

    return () => {
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  const loadOfflineQueue = useCallback(() => {
    const stored = localStorage.getItem(OFFLINE_STORAGE_KEY)
    if (stored) {
      try {
        setOfflineQueue(JSON.parse(stored))
      } catch (error) {
        console.error('Failed to load offline queue:', error)
      }
    }
  }, [])

  const addToOfflineQueue = useCallback((action) => {
    const queueItem = {
      id: Date.now().toString(),
      timestamp: new Date().toISOString(),
      action,
    }

    setOfflineQueue(prev => {
      const updated = [...prev, queueItem]
      localStorage.setItem(OFFLINE_STORAGE_KEY, JSON.stringify(updated))
      return updated
    })

    // Register background sync if available
    if ('serviceWorker' in navigator && 'sync' in window.ServiceWorkerRegistration.prototype) {
      navigator.serviceWorker.ready.then(registration => {
        registration.sync.register(`background-sync-${action.type}`)
      })
    }

    return queueItem.id
  }, [])

  const removeFromOfflineQueue = useCallback((id) => {
    setOfflineQueue(prev => {
      const updated = prev.filter(item => item.id !== id)
      localStorage.setItem(OFFLINE_STORAGE_KEY, JSON.stringify(updated))
      return updated
    })
  }, [])

  const syncOfflineData = useCallback(async () => {
    if (!isOnline || offlineQueue.length === 0) return

    for (const item of offlineQueue) {
      try {
        const success = await syncAction(item.action)
        if (success) {
          removeFromOfflineQueue(item.id)
        }
      } catch (error) {
        console.error('Failed to sync offline action:', error)
      }
    }
  }, [isOnline, offlineQueue, removeFromOfflineQueue])

  const syncAction = async (action) => {
    switch (action.type) {
      case 'place-bet':
        return await syncBet(action.data)
      case 'update-profile':
        return await syncProfile(action.data)
      case 'create-market':
        return await syncMarket(action.data)
      default:
        console.warn('Unknown action type:', action.type)
        return false
    }
  }

  return {
    isOnline,
    offlineQueue,
    addToOfflineQueue,
    removeFromOfflineQueue,
    syncOfflineData,
  }
}

// Individual sync functions
const syncBet = async (data) => {
  try {
    const response = await fetch('/api/bets', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.token}`,
      },
      body: JSON.stringify(data.bet),
    })
    return response.ok
  } catch (error) {
    console.error('Failed to sync bet:', error)
    return false
  }
}

const syncProfile = async (data) => {
  try {
    const response = await fetch('/api/users/profile', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.token}`,
      },
      body: JSON.stringify(data.profile),
    })
    return response.ok
  } catch (error) {
    console.error('Failed to sync profile:', error)
    return false
  }
}

const syncMarket = async (data) => {
  try {
    const response = await fetch('/api/markets', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.token}`,
      },
      body: JSON.stringify(data.market),
    })
    return response.ok
  } catch (error) {
    console.error('Failed to sync market:', error)
    return false
  }
}

// components/common/OfflineIndicator.jsx
import React from 'react'
import { useOfflineStorage } from '../../hooks/useOfflineStorage'
import { WifiIcon, ExclamationTriangleIcon } from '@heroicons/react/24/outline'

const OfflineIndicator = () => {
  const { isOnline, offlineQueue } = useOfflineStorage()

  if (isOnline && offlineQueue.length === 0) {
    return null
  }

  return (
    <div className={`fixed top-0 left-0 right-0 z-50 px-4 py-2 text-center text-sm font-medium ${
      isOnline ? 'bg-yellow-500 text-yellow-900' : 'bg-red-500 text-white'
    }`}>
      <div className="flex items-center justify-center space-x-2">
        {isOnline ? (
          <>
            <ExclamationTriangleIcon className="h-4 w-4" />
            <span>
              {offlineQueue.length} action{offlineQueue.length !== 1 ? 's' : ''} pending sync
            </span>
          </>
        ) : (
          <>
            <WifiIcon className="h-4 w-4" />
            <span>You're offline. Actions will sync when connection is restored.</span>
          </>
        )}
      </div>
    </div>
  )
}

export default OfflineIndicator

// components/common/OfflineFallback.jsx
import React from 'react'
import { WifiIcon, ArrowPathIcon } from '@heroicons/react/24/outline'

const OfflineFallback = ({ 
  title = "You're Offline",
  message = "This content requires an internet connection.",
  showRetry = true,
  onRetry,
}) => {
  const handleRetry = () => {
    if (onRetry) {
      onRetry()
    } else {
      window.location.reload()
    }
  }

  return (
    <div className="flex flex-col items-center justify-center min-h-64 p-8 text-center">
      <div className="w-16 h-16 mb-4 rounded-full bg-gray-100 flex items-center justify-center">
        <WifiIcon className="h-8 w-8 text-gray-400" />
      </div>
      
      <h3 className="text-lg font-semibold text-gray-900 mb-2">
        {title}
      </h3>
      
      <p className="text-gray-600 mb-6 max-w-md">
        {message}
      </p>
      
      {showRetry && (
        <button
          onClick={handleRetry}
          className="inline-flex items-center px-4 py-2 bg-indigo-600 text-white rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          <ArrowPathIcon className="h-4 w-4 mr-2" />
          Try Again
        </button>
      )}
    </div>
  )
}

export default OfflineFallback
```

### Step 3: Push Notifications Implementation
**Timeline: 3 days**

Implement comprehensive push notification system:

```javascript
// hooks/usePushNotifications.js
import { useState, useEffect, useCallback } from 'react'
import { useAuth } from './useAuth'

const VAPID_PUBLIC_KEY = process.env.REACT_APP_VAPID_PUBLIC_KEY

export const usePushNotifications = () => {
  const [permission, setPermission] = useState(Notification.permission)
  const [subscription, setSubscription] = useState(null)
  const [isSupported, setIsSupported] = useState(false)
  const { user, token } = useAuth()

  useEffect(() => {
    // Check if push notifications are supported
    setIsSupported(
      'serviceWorker' in navigator &&
      'PushManager' in window &&
      'Notification' in window
    )

    if (isSupported) {
      // Get existing subscription
      getExistingSubscription()
    }
  }, [isSupported])

  const getExistingSubscription = useCallback(async () => {
    try {
      const registration = await navigator.serviceWorker.ready
      const existingSubscription = await registration.pushManager.getSubscription()
      setSubscription(existingSubscription)
    } catch (error) {
      console.error('Failed to get existing subscription:', error)
    }
  }, [])

  const requestPermission = useCallback(async () => {
    if (!isSupported) {
      throw new Error('Push notifications not supported')
    }

    const result = await Notification.requestPermission()
    setPermission(result)
    
    if (result === 'granted') {
      await subscribe()
    }
    
    return result
  }, [isSupported])

  const subscribe = useCallback(async () => {
    if (!isSupported || permission !== 'granted') {
      throw new Error('Permission not granted for push notifications')
    }

    try {
      const registration = await navigator.serviceWorker.ready
      
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY),
      })

      setSubscription(subscription)

      // Send subscription to backend
      await sendSubscriptionToBackend(subscription)

      return subscription
    } catch (error) {
      console.error('Failed to subscribe to push notifications:', error)
      throw error
    }
  }, [isSupported, permission, user, token])

  const unsubscribe = useCallback(async () => {
    if (!subscription) return

    try {
      await subscription.unsubscribe()
      setSubscription(null)

      // Remove subscription from backend
      await removeSubscriptionFromBackend(subscription)
    } catch (error) {
      console.error('Failed to unsubscribe from push notifications:', error)
      throw error
    }
  }, [subscription])

  const sendSubscriptionToBackend = async (subscription) => {
    try {
      await fetch('/api/push/subscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          subscription,
          userId: user?.id,
          preferences: {
            marketUpdates: true,
            betResults: true,
            marketClosing: true,
            newMarkets: false,
          },
        }),
      })
    } catch (error) {
      console.error('Failed to send subscription to backend:', error)
    }
  }

  const removeSubscriptionFromBackend = async (subscription) => {
    try {
      await fetch('/api/push/unsubscribe', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          subscription,
          userId: user?.id,
        }),
      })
    } catch (error) {
      console.error('Failed to remove subscription from backend:', error)
    }
  }

  const updatePreferences = useCallback(async (preferences) => {
    if (!subscription) return

    try {
      await fetch('/api/push/preferences', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          subscription,
          userId: user?.id,
          preferences,
        }),
      })
    } catch (error) {
      console.error('Failed to update notification preferences:', error)
      throw error
    }
  }, [subscription, user, token])

  return {
    isSupported,
    permission,
    subscription,
    requestPermission,
    subscribe,
    unsubscribe,
    updatePreferences,
  }
}

// Utility function to convert VAPID key
function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - base64String.length % 4) % 4)
  const base64 = (base64String + padding)
    .replace(/-/g, '+')
    .replace(/_/g, '/')

  const rawData = window.atob(base64)
  const outputArray = new Uint8Array(rawData.length)

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i)
  }
  return outputArray
}

// components/notifications/NotificationSettings.jsx
import React, { useState, useEffect } from 'react'
import { usePushNotifications } from '../../hooks/usePushNotifications'
import { Switch } from '@headlessui/react'
import { BellIcon, BellSlashIcon } from '@heroicons/react/24/outline'

const NotificationSettings = () => {
  const {
    isSupported,
    permission,
    subscription,
    requestPermission,
    unsubscribe,
    updatePreferences,
  } = usePushNotifications()

  const [preferences, setPreferences] = useState({
    marketUpdates: true,
    betResults: true,
    marketClosing: true,
    newMarkets: false,
  })

  const [loading, setLoading] = useState(false)

  const handleEnableNotifications = async () => {
    setLoading(true)
    try {
      const result = await requestPermission()
      if (result === 'granted') {
        // Notifications enabled successfully
      }
    } catch (error) {
      console.error('Failed to enable notifications:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleDisableNotifications = async () => {
    setLoading(true)
    try {
      await unsubscribe()
    } catch (error) {
      console.error('Failed to disable notifications:', error)
    } finally {
      setLoading(false)
    }
  }

  const handlePreferenceChange = async (key, value) => {
    const newPreferences = { ...preferences, [key]: value }
    setPreferences(newPreferences)

    if (subscription) {
      try {
        await updatePreferences(newPreferences)
      } catch (error) {
        console.error('Failed to update preferences:', error)
        // Revert change on error
        setPreferences(preferences)
      }
    }
  }

  if (!isSupported) {
    return (
      <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
        <div className="flex">
          <BellSlashIcon className="h-5 w-5 text-yellow-400" />
          <div className="ml-3">
            <h3 className="text-sm font-medium text-yellow-800">
              Push notifications not supported
            </h3>
            <p className="mt-1 text-sm text-yellow-700">
              Your browser doesn't support push notifications.
            </p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <BellIcon className="h-8 w-8 text-indigo-600" />
              <div className="ml-4">
                <h3 className="text-lg font-medium text-gray-900">
                  Push Notifications
                </h3>
                <p className="text-sm text-gray-600">
                  Get notified about market updates and bet results
                </p>
              </div>
            </div>
            
            <div className="flex items-center space-x-4">
              {permission === 'granted' && subscription ? (
                <button
                  onClick={handleDisableNotifications}
                  disabled={loading}
                  className="bg-red-600 text-white px-4 py-2 rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50"
                >
                  {loading ? 'Disabling...' : 'Disable'}
                </button>
              ) : (
                <button
                  onClick={handleEnableNotifications}
                  disabled={loading || permission === 'denied'}
                  className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50"
                >
                  {loading ? 'Enabling...' : 'Enable'}
                </button>
              )}
            </div>
          </div>

          {permission === 'denied' && (
            <div className="mt-4 bg-red-50 border border-red-200 rounded-md p-4">
              <p className="text-sm text-red-800">
                Notifications are blocked. Please enable them in your browser settings.
              </p>
            </div>
          )}
        </div>
      </div>

      {permission === 'granted' && subscription && (
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">
              Notification Preferences
            </h3>
            
            <div className="space-y-4">
              {Object.entries({
                marketUpdates: 'Market price updates',
                betResults: 'Bet results',
                marketClosing: 'Markets closing soon',
                newMarkets: 'New markets available',
              }).map(([key, label]) => (
                <div key={key} className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-700">
                    {label}
                  </span>
                  <Switch
                    checked={preferences[key]}
                    onChange={(value) => handlePreferenceChange(key, value)}
                    className={`${
                      preferences[key] ? 'bg-indigo-600' : 'bg-gray-200'
                    } relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2`}
                  >
                    <span
                      className={`${
                        preferences[key] ? 'translate-x-6' : 'translate-x-1'
                      } inline-block h-4 w-4 transform rounded-full bg-white transition-transform`}
                    />
                  </Switch>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default NotificationSettings
```

### Step 4: App Installation and Native Features
**Timeline: 2 days**

Implement app installation prompts and native device features:

```javascript
// hooks/useInstallPrompt.js
import { useState, useEffect } from 'react'

export const useInstallPrompt = () => {
  const [deferredPrompt, setDeferredPrompt] = useState(null)
  const [showInstallPrompt, setShowInstallPrompt] = useState(false)
  const [isInstalled, setIsInstalled] = useState(false)

  useEffect(() => {
    // Check if app is already installed
    if (window.matchMedia('(display-mode: standalone)').matches) {
      setIsInstalled(true)
    }

    const handleBeforeInstallPrompt = (e) => {
      // Prevent Chrome 67 and earlier from automatically showing the prompt
      e.preventDefault()
      
      // Stash the event so it can be triggered later
      setDeferredPrompt(e)
      
      // Show custom install prompt after a delay
      setTimeout(() => {
        setShowInstallPrompt(true)
      }, 30000) // Show after 30 seconds
    }

    const handleAppInstalled = () => {
      setIsInstalled(true)
      setDeferredPrompt(null)
      setShowInstallPrompt(false)
      
      // Track installation
      if (window.gtag) {
        window.gtag('event', 'pwa_install', {
          event_category: 'engagement',
        })
      }
    }

    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
    window.addEventListener('appinstalled', handleAppInstalled)

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt)
      window.removeEventListener('appinstalled', handleAppInstalled)
    }
  }, [])

  const installApp = async () => {
    if (!deferredPrompt) return false

    try {
      // Show the install prompt
      deferredPrompt.prompt()
      
      // Wait for the user to respond to the prompt
      const { outcome } = await deferredPrompt.userChoice
      
      if (outcome === 'accepted') {
        setDeferredPrompt(null)
        setShowInstallPrompt(false)
        return true
      }
      
      return false
    } catch (error) {
      console.error('Failed to install app:', error)
      return false
    }
  }

  const dismissPrompt = () => {
    setShowInstallPrompt(false)
    localStorage.setItem('install-prompt-dismissed', Date.now().toString())
  }

  return {
    canInstall: !!deferredPrompt,
    showInstallPrompt,
    isInstalled,
    installApp,
    dismissPrompt,
  }
}

// components/pwa/InstallPrompt.jsx
import React from 'react'
import { useInstallPrompt } from '../../hooks/useInstallPrompt'
import { ArrowDownTrayIcon, XMarkIcon } from '@heroicons/react/24/outline'

const InstallPrompt = () => {
  const { showInstallPrompt, installApp, dismissPrompt } = useInstallPrompt()

  if (!showInstallPrompt) return null

  const handleInstall = async () => {
    const success = await installApp()
    if (success) {
      // Installation successful
    }
  }

  return (
    <div className="fixed bottom-4 left-4 right-4 bg-white border border-gray-200 rounded-lg shadow-lg p-4 z-50">
      <div className="flex items-start justify-between">
        <div className="flex items-center space-x-3">
          <div className="w-12 h-12 bg-indigo-100 rounded-lg flex items-center justify-center">
            <ArrowDownTrayIcon className="h-6 w-6 text-indigo-600" />
          </div>
          
          <div className="flex-1">
            <h3 className="text-sm font-semibold text-gray-900">
              Install SocialPredict
            </h3>
            <p className="text-xs text-gray-600 mt-1">
              Add to your home screen for quick access and offline use
            </p>
          </div>
        </div>
        
        <button
          onClick={dismissPrompt}
          className="text-gray-400 hover:text-gray-600"
        >
          <XMarkIcon className="h-5 w-5" />
        </button>
      </div>
      
      <div className="flex space-x-3 mt-4">
        <button
          onClick={handleInstall}
          className="flex-1 bg-indigo-600 text-white text-sm font-medium py-2 px-4 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
        >
          Install
        </button>
        
        <button
          onClick={dismissPrompt}
          className="flex-1 bg-gray-100 text-gray-700 text-sm font-medium py-2 px-4 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
        >
          Not now
        </button>
      </div>
    </div>
  )
}

export default InstallPrompt

// components/pwa/PWAUpdatePrompt.jsx
import React, { useState, useEffect } from 'react'
import { ArrowPathIcon, XMarkIcon } from '@heroicons/react/24/outline'

const PWAUpdatePrompt = () => {
  const [showUpdatePrompt, setShowUpdatePrompt] = useState(false)
  const [waitingWorker, setWaitingWorker] = useState(null)

  useEffect(() => {
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.addEventListener('controllerchange', () => {
        window.location.reload()
      })

      navigator.serviceWorker.ready.then((registration) => {
        registration.addEventListener('updatefound', () => {
          const newWorker = registration.installing
          
          newWorker.addEventListener('statechange', () => {
            if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
              setWaitingWorker(newWorker)
              setShowUpdatePrompt(true)
            }
          })
        })
      })
    }
  }, [])

  const handleUpdate = () => {
    if (waitingWorker) {
      waitingWorker.postMessage({ type: 'SKIP_WAITING' })
      setShowUpdatePrompt(false)
    }
  }

  const handleDismiss = () => {
    setShowUpdatePrompt(false)
  }

  if (!showUpdatePrompt) return null

  return (
    <div className="fixed top-4 left-4 right-4 bg-blue-600 text-white rounded-lg shadow-lg p-4 z-50">
      <div className="flex items-start justify-between">
        <div className="flex items-center space-x-3">
          <ArrowPathIcon className="h-6 w-6" />
          <div>
            <h3 className="text-sm font-semibold">
              App Update Available
            </h3>
            <p className="text-xs opacity-90 mt-1">
              A new version is ready to install
            </p>
          </div>
        </div>
        
        <button
          onClick={handleDismiss}
          className="text-white opacity-75 hover:opacity-100"
        >
          <XMarkIcon className="h-5 w-5" />
        </button>
      </div>
      
      <div className="flex space-x-3 mt-4">
        <button
          onClick={handleUpdate}
          className="bg-white text-blue-600 text-sm font-medium py-2 px-4 rounded-md hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-white"
        >
          Update Now
        </button>
        
        <button
          onClick={handleDismiss}
          className="border border-white text-white text-sm font-medium py-2 px-4 rounded-md hover:bg-white hover:bg-opacity-10 focus:outline-none focus:ring-2 focus:ring-white"
        >
          Later
        </button>
      </div>
    </div>
  )
}

export default PWAUpdatePrompt
```

## Directory Structure
```
src/
├── components/
│   ├── pwa/
│   │   ├── InstallPrompt.jsx
│   │   ├── PWAUpdatePrompt.jsx
│   │   └── OfflineIndicator.jsx
│   ├── notifications/
│   │   ├── NotificationSettings.jsx
│   │   └── NotificationPermissionBanner.jsx
│   └── common/
│       └── OfflineFallback.jsx
├── hooks/
│   ├── useInstallPrompt.js
│   ├── usePushNotifications.js
│   ├── useOfflineStorage.js
│   └── useNetworkStatus.js
├── services/
│   ├── notificationService.js
│   ├── backgroundSync.js
│   └── cacheManager.js
├── utils/
│   ├── pwaUtils.js
│   └── offlineUtils.js
└── workers/
    └── sw.js              # Service Worker
public/
├── manifest.json          # Web App Manifest
├── offline.html          # Offline fallback page
└── icons/               # PWA icons (various sizes)
    ├── icon-72x72.png
    ├── icon-96x96.png
    ├── icon-128x128.png
    ├── icon-144x144.png
    ├── icon-152x152.png
    ├── icon-192x192.png
    ├── icon-384x384.png
    └── icon-512x512.png
```

## Benefits
- Native app-like experience
- Offline functionality with data synchronization
- Push notifications for user engagement
- Faster loading with intelligent caching
- Reduced bandwidth usage
- Better user retention
- Enhanced mobile experience
- Background sync capabilities
- App store distribution potential
- Cross-platform compatibility

## PWA Features Implemented
- ✅ Service Worker with caching strategies
- ✅ Web App Manifest
- ✅ Offline functionality
- ✅ Background sync
- ✅ Push notifications
- ✅ App installation prompts
- ✅ Update notifications
- ✅ Network status detection
- ✅ Offline data queue
- ✅ Icon generation
- ✅ Splash screens
- ✅ App shortcuts