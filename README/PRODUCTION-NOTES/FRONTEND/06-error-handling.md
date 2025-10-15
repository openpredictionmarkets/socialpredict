# Error Handling Implementation Plan

## Overview
Implement comprehensive error handling strategies to gracefully manage errors, provide meaningful user feedback, and maintain application stability across all components and user interactions.

## Current State Analysis
- Basic try-catch blocks in some components
- Limited error boundaries implementation
- No centralized error reporting
- Minimal user-friendly error messages
- No error recovery mechanisms
- No error logging or monitoring
- Basic API error handling without retry logic

## Implementation Steps

### Step 1: Global Error Boundary System
**Timeline: 2-3 days**

Implement comprehensive error boundaries with recovery mechanisms:

```javascript
// components/errors/GlobalErrorBoundary.jsx
import React from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import ErrorFallback from './ErrorFallback'
import { errorReporter } from '../../services/errorReporter'

const GlobalErrorBoundary = ({ children }) => {
  const handleError = (error, errorInfo) => {
    // Log error to console in development
    if (process.env.NODE_ENV === 'development') {
      console.error('Global Error Boundary caught an error:', error, errorInfo)
    }

    // Report error to monitoring service
    errorReporter.captureException(error, {
      errorBoundary: 'GlobalErrorBoundary',
      componentStack: errorInfo.componentStack,
      errorBoundaryStack: errorInfo.errorBoundaryStack,
      eventId: errorInfo.eventId,
      extra: {
        timestamp: new Date().toISOString(),
        url: window.location.href,
        userAgent: navigator.userAgent,
      },
    })
  }

  const handleReset = () => {
    // Reset application state if needed
    window.location.reload()
  }

  return (
    <ErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={handleError}
      onReset={handleReset}
      resetKeys={[window.location.pathname]}
    >
      {children}
    </ErrorBoundary>
  )
}

export default GlobalErrorBoundary

// components/errors/ErrorFallback.jsx
import React, { useState } from 'react'
import { ChevronDownIcon, ChevronRightIcon } from '@heroicons/react/24/outline'

const ErrorFallback = ({ error, resetErrorBoundary, resetKeys }) => {
  const [showDetails, setShowDetails] = useState(false)
  const [reportSent, setReportSent] = useState(false)

  const handleSendReport = async () => {
    try {
      await fetch('/api/v0/errors/report', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          error: error.message,
          stack: error.stack,
          timestamp: new Date().toISOString(),
          url: window.location.href,
          userAgent: navigator.userAgent,
        }),
      })
      setReportSent(true)
    } catch (reportError) {
      console.error('Failed to send error report:', reportError)
    }
  }

  const getErrorLevel = (error) => {
    if (error.name === 'ChunkLoadError' || error.message.includes('Loading chunk')) {
      return 'recoverable'
    }
    if (error.message.includes('Network Error') || error.message.includes('fetch')) {
      return 'network'
    }
    return 'critical'
  }

  const getErrorMessage = (error) => {
    const level = getErrorLevel(error)
    
    switch (level) {
      case 'recoverable':
        return 'We detected an issue loading part of the application. Please try refreshing the page.'
      case 'network':
        return 'We\'re having trouble connecting to our servers. Please check your internet connection and try again.'
      case 'critical':
      default:
        return 'Something went wrong. We\'ve been notified and are working to fix this issue.'
    }
  }

  const getRecoveryActions = (error) => {
    const level = getErrorLevel(error)
    
    const actions = [
      {
        label: 'Try Again',
        action: resetErrorBoundary,
        primary: true,
      },
    ]

    if (level === 'recoverable' || level === 'network') {
      actions.push({
        label: 'Refresh Page',
        action: () => window.location.reload(),
        primary: false,
      })
    }

    actions.push({
      label: 'Go Home',
      action: () => window.location.href = '/',
      primary: false,
    })

    return actions
  }

  const errorLevel = getErrorLevel(error)
  const errorMessage = getErrorMessage(error)
  const recoveryActions = getRecoveryActions(error)

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <div className={`mx-auto h-12 w-12 flex items-center justify-center rounded-full ${
            errorLevel === 'critical' ? 'bg-red-100' : 'bg-yellow-100'
          }`}>
            <svg
              className={`h-6 w-6 ${
                errorLevel === 'critical' ? 'text-red-600' : 'text-yellow-600'
              }`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.464 0L4.732 15.5c-.77.833.192 2.5 1.732 2.5z"
              />
            </svg>
          </div>
          
          <h2 className="mt-6 text-3xl font-extrabold text-gray-900">
            Oops! Something went wrong
          </h2>
          
          <p className="mt-2 text-sm text-gray-600">
            {errorMessage}
          </p>
        </div>

        <div className="space-y-3">
          {recoveryActions.map((action, index) => (
            <button
              key={index}
              onClick={action.action}
              className={`w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white focus:outline-none focus:ring-2 focus:ring-offset-2 ${
                action.primary
                  ? 'bg-indigo-600 hover:bg-indigo-700 focus:ring-indigo-500'
                  : 'bg-gray-600 hover:bg-gray-700 focus:ring-gray-500'
              }`}
            >
              {action.label}
            </button>
          ))}
        </div>

        <div className="border-t border-gray-200 pt-4 space-y-2">
          {!reportSent ? (
            <button
              onClick={handleSendReport}
              className="w-full text-sm text-indigo-600 hover:text-indigo-500 underline"
            >
              Send Error Report
            </button>
          ) : (
            <p className="text-sm text-green-600 text-center">
              Error report sent. Thank you!
            </p>
          )}

          <button
            onClick={() => setShowDetails(!showDetails)}
            className="w-full flex items-center justify-center text-sm text-gray-500 hover:text-gray-700"
          >
            {showDetails ? (
              <ChevronDownIcon className="w-4 h-4 mr-1" />
            ) : (
              <ChevronRightIcon className="w-4 h-4 mr-1" />
            )}
            Technical Details
          </button>

          {showDetails && (
            <div className="mt-2 p-3 bg-gray-100 rounded text-xs text-gray-700 overflow-auto">
              <div className="mb-2">
                <strong>Error:</strong> {error.name}
              </div>
              <div className="mb-2">
                <strong>Message:</strong> {error.message}
              </div>
              <div>
                <strong>Stack:</strong>
                <pre className="mt-1 whitespace-pre-wrap text-xs">
                  {error.stack}
                </pre>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default ErrorFallback

// components/errors/AsyncErrorBoundary.jsx
import React, { useState, useEffect } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import LoadingSpinner from '../common/LoadingSpinner'
import ErrorFallback from './ErrorFallback'

const AsyncErrorBoundary = ({ 
  children, 
  fallback = null,
  onError,
  resetKeys = [],
  isolate = false,
}) => {
  const [hasError, setHasError] = useState(false)
  const [retryCount, setRetryCount] = useState(0)
  const maxRetries = 3

  const handleError = (error, errorInfo) => {
    setHasError(true)
    
    if (onError) {
      onError(error, errorInfo)
    }

    // Auto-retry for network errors
    if ((error.message.includes('Network') || error.message.includes('fetch')) && retryCount < maxRetries) {
      setTimeout(() => {
        setRetryCount(prev => prev + 1)
        setHasError(false)
      }, 1000 * Math.pow(2, retryCount)) // Exponential backoff
    }
  }

  const handleReset = () => {
    setHasError(false)
    setRetryCount(0)
  }

  // Reset when resetKeys change
  useEffect(() => {
    if (hasError) {
      handleReset()
    }
  }, resetKeys)

  const ErrorFallbackComponent = fallback || (isolate ? IsolatedErrorFallback : ErrorFallback)

  return (
    <ErrorBoundary
      FallbackComponent={ErrorFallbackComponent}
      onError={handleError}
      onReset={handleReset}
      resetKeys={[...resetKeys, retryCount]}
    >
      {children}
    </ErrorBoundary>
  )
}

const IsolatedErrorFallback = ({ error, resetErrorBoundary }) => (
  <div className="bg-red-50 border border-red-200 rounded-md p-4 my-4">
    <div className="flex">
      <div className="flex-shrink-0">
        <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
          <path
            fillRule="evenodd"
            d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
            clipRule="evenodd"
          />
        </svg>
      </div>
      <div className="ml-3">
        <h3 className="text-sm font-medium text-red-800">
          Component Error
        </h3>
        <div className="mt-2 text-sm text-red-700">
          <p>This section encountered an error and couldn't load.</p>
        </div>
        <div className="mt-4">
          <button
            onClick={resetErrorBoundary}
            className="bg-red-100 px-3 py-2 rounded-md text-sm font-medium text-red-800 hover:bg-red-200"
          >
            Try Again
          </button>
        </div>
      </div>
    </div>
  </div>
)

export default AsyncErrorBoundary
```

### Step 2: API Error Handling
**Timeline: 2-3 days**

Implement comprehensive API error handling with retry logic:

```javascript
// services/errorHandler.js
export class APIErrorHandler {
  constructor() {
    this.retryableStatuses = [408, 429, 500, 502, 503, 504]
    this.maxRetries = 3
    this.baseDelay = 1000
    this.maxDelay = 10000
  }

  isRetryableError(error) {
    if (error.name === 'NetworkError' || error.message.includes('fetch')) {
      return true
    }
    
    if (error.response) {
      return this.retryableStatuses.includes(error.response.status)
    }
    
    return false
  }

  getRetryDelay(attempt) {
    const delay = this.baseDelay * Math.pow(2, attempt)
    const jitter = Math.random() * 0.1 * delay
    return Math.min(delay + jitter, this.maxDelay)
  }

  async executeWithRetry(apiCall, options = {}) {
    const { maxRetries = this.maxRetries, onRetry, context } = options
    let lastError

    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        return await apiCall()
      } catch (error) {
        lastError = error
        
        if (attempt === maxRetries || !this.isRetryableError(error)) {
          throw this.enhanceError(error, context)
        }

        const delay = this.getRetryDelay(attempt)
        
        if (onRetry) {
          onRetry(error, attempt + 1, delay)
        }

        await this.delay(delay)
      }
    }

    throw this.enhanceError(lastError, context)
  }

  enhanceError(error, context = {}) {
    const enhancedError = new Error()
    enhancedError.name = error.name || 'APIError'
    enhancedError.message = this.getUserFriendlyMessage(error)
    enhancedError.originalError = error
    enhancedError.context = context
    enhancedError.timestamp = new Date().toISOString()
    enhancedError.userAgent = navigator.userAgent
    enhancedError.url = window.location.href

    // Add response information if available
    if (error.response) {
      enhancedError.status = error.response.status
      enhancedError.statusText = error.response.statusText
      enhancedError.responseData = error.response.data
    }

    return enhancedError
  }

  getUserFriendlyMessage(error) {
    // Network errors
    if (error.name === 'NetworkError' || error.message.includes('fetch')) {
      return 'Unable to connect to the server. Please check your internet connection and try again.'
    }

    // HTTP status errors
    if (error.response) {
      const status = error.response.status
      
      switch (status) {
        case 400:
          return 'Invalid request. Please check your input and try again.'
        case 401:
          return 'Your session has expired. Please log in again.'
        case 403:
          return 'You don\'t have permission to perform this action.'
        case 404:
          return 'The requested resource was not found.'
        case 408:
          return 'Request timed out. Please try again.'
        case 409:
          return 'There was a conflict with your request. Please refresh and try again.'
        case 429:
          return 'Too many requests. Please wait a moment and try again.'
        case 500:
          return 'Server error. We\'ve been notified and are working to fix this.'
        case 502:
        case 503:
        case 504:
          return 'Service temporarily unavailable. Please try again in a few moments.'
        default:
          return 'Something went wrong. Please try again.'
      }
    }

    // Specific error types
    if (error.message.includes('timeout')) {
      return 'Request timed out. Please try again.'
    }

    if (error.message.includes('abort')) {
      return 'Request was cancelled.'
    }

    return 'An unexpected error occurred. Please try again.'
  }

  delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
  }
}

export const apiErrorHandler = new APIErrorHandler()

// hooks/useAPIError.js
import { useState } from 'react'
import { apiErrorHandler } from '../services/errorHandler'
import { useNotification } from './useNotification'

export const useAPIError = () => {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)
  const { showNotification } = useNotification()

  const executeAPI = async (apiCall, options = {}) => {
    const {
      showErrorNotification = true,
      showRetryNotification = false,
      loadingState = true,
      context = {},
    } = options

    if (loadingState) {
      setLoading(true)
    }
    setError(null)

    try {
      const result = await apiErrorHandler.executeWithRetry(apiCall, {
        ...options,
        context,
        onRetry: (error, attempt, delay) => {
          if (showRetryNotification) {
            showNotification({
              type: 'info',
              message: `Retrying request (attempt ${attempt})...`,
              duration: 2000,
            })
          }
          
          if (options.onRetry) {
            options.onRetry(error, attempt, delay)
          }
        },
      })

      return result
    } catch (error) {
      setError(error)
      
      if (showErrorNotification) {
        showNotification({
          type: 'error',
          message: error.message,
          duration: 5000,
          action: {
            label: 'Retry',
            onClick: () => executeAPI(apiCall, options),
          },
        })
      }

      throw error
    } finally {
      if (loadingState) {
        setLoading(false)
      }
    }
  }

  const clearError = () => setError(null)

  return {
    loading,
    error,
    executeAPI,
    clearError,
  }
}

// services/apiClient.js - Enhanced API client with error handling
import { apiErrorHandler } from './errorHandler'

class APIClient {
  constructor() {
    this.baseURL = process.env.REACT_APP_API_URL || '/api/v0'
    this.defaultHeaders = {
      'Content-Type': 'application/json',
    }
  }

  async request(endpoint, options = {}) {
    const {
      method = 'GET',
      data,
      headers = {},
      timeout = 30000,
      retryOptions = {},
      context = {},
    } = options

    const apiCall = async () => {
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), timeout)

      try {
        const response = await fetch(`${this.baseURL}${endpoint}`, {
          method,
          headers: {
            ...this.defaultHeaders,
            ...headers,
          },
          body: data ? JSON.stringify(data) : undefined,
          signal: controller.signal,
        })

        clearTimeout(timeoutId)

        if (!response.ok) {
          throw new APIError(response)
        }

        const contentType = response.headers.get('content-type')
        if (contentType && contentType.includes('application/json')) {
          return await response.json()
        }

        return await response.text()
      } catch (error) {
        clearTimeout(timeoutId)
        
        if (error.name === 'AbortError') {
          throw new Error('Request timeout')
        }
        
        throw error
      }
    }

    return apiErrorHandler.executeWithRetry(apiCall, {
      ...retryOptions,
      context: {
        endpoint,
        method,
        ...context,
      },
    })
  }

  get(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'GET' })
  }

  post(endpoint, data, options = {}) {
    return this.request(endpoint, { ...options, method: 'POST', data })
  }

  put(endpoint, data, options = {}) {
    return this.request(endpoint, { ...options, method: 'PUT', data })
  }

  delete(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'DELETE' })
  }
}

class APIError extends Error {
  constructor(response) {
    super(`HTTP ${response.status}: ${response.statusText}`)
    this.name = 'APIError'
    this.response = response
    this.status = response.status
    this.statusText = response.statusText
  }
}

export const apiClient = new APIClient()
```

### Step 3: User-Friendly Error Messages
**Timeline: 1-2 days**

Implement user-friendly error messaging system:

```javascript
// components/errors/ErrorMessage.jsx
import React from 'react'
import { XCircleIcon, ExclamationTriangleIcon, InformationCircleIcon } from '@heroicons/react/24/outline'

const ErrorMessage = ({ 
  error, 
  type = 'error', 
  title, 
  showRetry = true, 
  onRetry,
  className = '',
  dismissible = false,
  onDismiss,
}) => {
  const getIcon = () => {
    switch (type) {
      case 'warning':
        return <ExclamationTriangleIcon className="h-5 w-5 text-yellow-400" />
      case 'info':
        return <InformationCircleIcon className="h-5 w-5 text-blue-400" />
      case 'error':
      default:
        return <XCircleIcon className="h-5 w-5 text-red-400" />
    }
  }

  const getColorClasses = () => {
    switch (type) {
      case 'warning':
        return 'bg-yellow-50 border-yellow-200'
      case 'info':
        return 'bg-blue-50 border-blue-200'
      case 'error':
      default:
        return 'bg-red-50 border-red-200'
    }
  }

  const getTextColorClasses = () => {
    switch (type) {
      case 'warning':
        return 'text-yellow-800'
      case 'info':
        return 'text-blue-800'
      case 'error':
      default:
        return 'text-red-800'
    }
  }

  const getMessage = () => {
    if (typeof error === 'string') {
      return error
    }
    
    if (error?.message) {
      return error.message
    }
    
    return 'An unexpected error occurred'
  }

  const getErrorId = () => {
    if (error?.context?.errorId) {
      return error.context.errorId
    }
    
    if (error?.timestamp) {
      return `Error ID: ${new Date(error.timestamp).getTime()}`
    }
    
    return null
  }

  return (
    <div className={`rounded-md border p-4 ${getColorClasses()} ${className}`}>
      <div className="flex">
        <div className="flex-shrink-0">
          {getIcon()}
        </div>
        <div className="ml-3 flex-1">
          {title && (
            <h3 className={`text-sm font-medium ${getTextColorClasses()}`}>
              {title}
            </h3>
          )}
          <div className={`${title ? 'mt-2' : ''} text-sm ${getTextColorClasses()}`}>
            <p>{getMessage()}</p>
            
            {getErrorId() && (
              <p className="mt-1 text-xs opacity-75">
                {getErrorId()}
              </p>
            )}
          </div>
          
          {(showRetry && onRetry) && (
            <div className="mt-4">
              <button
                onClick={onRetry}
                className={`text-sm font-medium underline hover:no-underline ${getTextColorClasses()}`}
              >
                Try Again
              </button>
            </div>
          )}
        </div>
        
        {dismissible && onDismiss && (
          <div className="ml-auto pl-3">
            <div className="-mx-1.5 -my-1.5">
              <button
                onClick={onDismiss}
                className={`inline-flex rounded-md p-1.5 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-red-50 focus:ring-red-600 ${getTextColorClasses()}`}
              >
                <span className="sr-only">Dismiss</span>
                <XCircleIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default ErrorMessage

// components/errors/ErrorToast.jsx
import React, { useEffect, useState } from 'react'
import { Transition } from '@headlessui/react'
import ErrorMessage from './ErrorMessage'

const ErrorToast = ({ 
  error, 
  onClose, 
  duration = 5000,
  position = 'top-right',
}) => {
  const [show, setShow] = useState(true)

  useEffect(() => {
    const timer = setTimeout(() => {
      setShow(false)
      setTimeout(onClose, 300) // Wait for animation to complete
    }, duration)

    return () => clearTimeout(timer)
  }, [duration, onClose])

  const getPositionClasses = () => {
    switch (position) {
      case 'top-left':
        return 'top-0 left-0'
      case 'top-right':
        return 'top-0 right-0'
      case 'bottom-left':
        return 'bottom-0 left-0'
      case 'bottom-right':
        return 'bottom-0 right-0'
      default:
        return 'top-0 right-0'
    }
  }

  return (
    <div className={`fixed ${getPositionClasses()} z-50 m-4`}>
      <Transition
        show={show}
        enter="transform ease-out duration-300 transition"
        enterFrom="translate-y-2 opacity-0 sm:translate-y-0 sm:translate-x-2"
        enterTo="translate-y-0 opacity-100 sm:translate-x-0"
        leave="transition ease-in duration-100"
        leaveFrom="opacity-100"
        leaveTo="opacity-0"
      >
        <div className="max-w-sm w-full">
          <ErrorMessage
            error={error}
            dismissible
            onDismiss={() => setShow(false)}
            showRetry={false}
          />
        </div>
      </Transition>
    </div>
  )
}

export default ErrorToast
```

### Step 4: Error Reporting and Monitoring
**Timeline: 2 days**

Implement error reporting and monitoring:

```javascript
// services/errorReporter.js
class ErrorReporter {
  constructor() {
    this.reportEndpoint = '/api/v0/errors/report'
    this.batchSize = 10
    this.batchTimeout = 5000
    this.errorQueue = []
    this.batchTimer = null
    this.userId = null
    this.sessionId = this.generateSessionId()
    
    this.setupGlobalErrorHandlers()
  }

  generateSessionId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2)
  }

  setUserId(userId) {
    this.userId = userId
  }

  setupGlobalErrorHandlers() {
    // Catch unhandled JavaScript errors
    window.addEventListener('error', (event) => {
      this.captureException(event.error || new Error(event.message), {
        type: 'javascript_error',
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
      })
    })

    // Catch unhandled promise rejections
    window.addEventListener('unhandledrejection', (event) => {
      this.captureException(event.reason, {
        type: 'unhandled_promise_rejection',
      })
    })

    // Catch resource loading errors
    window.addEventListener('error', (event) => {
      if (event.target !== window) {
        this.captureException(new Error(`Resource load error: ${event.target.src || event.target.href}`), {
          type: 'resource_error',
          element: event.target.tagName,
          source: event.target.src || event.target.href,
        })
      }
    }, true)
  }

  captureException(error, context = {}) {
    const errorData = {
      id: this.generateErrorId(),
      message: error.message,
      name: error.name,
      stack: error.stack,
      timestamp: new Date().toISOString(),
      url: window.location.href,
      userAgent: navigator.userAgent,
      userId: this.userId,
      sessionId: this.sessionId,
      context,
      breadcrumbs: this.getBreadcrumbs(),
      environment: {
        viewport: `${window.innerWidth}x${window.innerHeight}`,
        screen: `${window.screen.width}x${window.screen.height}`,
        language: navigator.language,
        platform: navigator.platform,
        cookieEnabled: navigator.cookieEnabled,
        onLine: navigator.onLine,
      },
    }

    this.addToQueue(errorData)
    
    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      console.error('Error captured:', errorData)
    }
  }

  addToQueue(errorData) {
    this.errorQueue.push(errorData)
    
    if (this.errorQueue.length >= this.batchSize) {
      this.flushQueue()
    } else if (!this.batchTimer) {
      this.batchTimer = setTimeout(() => {
        this.flushQueue()
      }, this.batchTimeout)
    }
  }

  async flushQueue() {
    if (this.errorQueue.length === 0) return

    const errors = [...this.errorQueue]
    this.errorQueue = []
    
    if (this.batchTimer) {
      clearTimeout(this.batchTimer)
      this.batchTimer = null
    }

    try {
      await fetch(this.reportEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ errors }),
      })
    } catch (reportError) {
      console.error('Failed to report errors:', reportError)
      // Re-add errors to queue for retry
      this.errorQueue.unshift(...errors)
    }
  }

  generateErrorId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2)
  }

  getBreadcrumbs() {
    // Implementation would track user actions leading to the error
    // This is a simplified version
    return [
      {
        timestamp: new Date().toISOString(),
        category: 'navigation',
        message: `User on page: ${window.location.pathname}`,
      },
    ]
  }

  // Public API
  captureMessage(message, level = 'info', context = {}) {
    this.captureException(new Error(message), {
      level,
      type: 'manual_message',
      ...context,
    })
  }

  addBreadcrumb(breadcrumb) {
    // Add to breadcrumb trail
    this.breadcrumbs = this.breadcrumbs || []
    this.breadcrumbs.push({
      timestamp: new Date().toISOString(),
      ...breadcrumb,
    })
    
    // Keep only last 10 breadcrumbs
    if (this.breadcrumbs.length > 10) {
      this.breadcrumbs = this.breadcrumbs.slice(-10)
    }
  }

  setContext(key, value) {
    this.contextData = this.contextData || {}
    this.contextData[key] = value
  }
}

export const errorReporter = new ErrorReporter()

// hooks/useErrorReporting.js
import { useEffect } from 'react'
import { errorReporter } from '../services/errorReporter'
import { useAuth } from './useAuth'

export const useErrorReporting = () => {
  const { user } = useAuth()

  useEffect(() => {
    if (user) {
      errorReporter.setUserId(user.id)
    }
  }, [user])

  const reportError = (error, context = {}) => {
    errorReporter.captureException(error, context)
  }

  const reportMessage = (message, level = 'info', context = {}) => {
    errorReporter.captureMessage(message, level, context)
  }

  const addBreadcrumb = (breadcrumb) => {
    errorReporter.addBreadcrumb(breadcrumb)
  }

  return {
    reportError,
    reportMessage,
    addBreadcrumb,
  }
}
```

### Step 5: Recovery Mechanisms
**Timeline: 2 days**

Implement error recovery mechanisms:

```javascript
// hooks/useErrorRecovery.js
import { useState, useCallback } from 'react'
import { useLocalStorage } from './useLocalStorage'

export const useErrorRecovery = () => {
  const [recoveryAttempts, setRecoveryAttempts] = useLocalStorage('errorRecoveryAttempts', {})
  const [isRecovering, setIsRecovering] = useState(false)

  const shouldAttemptRecovery = useCallback((errorType) => {
    const attempts = recoveryAttempts[errorType] || 0
    return attempts < 3 // Max 3 recovery attempts per error type
  }, [recoveryAttempts])

  const attemptRecovery = useCallback(async (errorType, recoveryStrategy) => {
    if (!shouldAttemptRecovery(errorType)) {
      return false
    }

    setIsRecovering(true)
    
    try {
      await recoveryStrategy()
      
      // Reset attempts on successful recovery
      setRecoveryAttempts(prev => ({
        ...prev,
        [errorType]: 0,
      }))
      
      return true
    } catch (error) {
      // Increment attempts on failed recovery
      setRecoveryAttempts(prev => ({
        ...prev,
        [errorType]: (prev[errorType] || 0) + 1,
      }))
      
      return false
    } finally {
      setIsRecovering(false)
    }
  }, [shouldAttemptRecovery, setRecoveryAttempts])

  const clearRecoveryAttempts = useCallback((errorType) => {
    setRecoveryAttempts(prev => {
      const updated = { ...prev }
      delete updated[errorType]
      return updated
    })
  }, [setRecoveryAttempts])

  return {
    isRecovering,
    shouldAttemptRecovery,
    attemptRecovery,
    clearRecoveryAttempts,
    recoveryAttempts,
  }
}

// components/errors/ErrorRecovery.jsx
import React, { useEffect } from 'react'
import { useErrorRecovery } from '../../hooks/useErrorRecovery'
import { useErrorReporting } from '../../hooks/useErrorReporting'

const ErrorRecovery = ({ error, children, fallback, recoveryStrategies = {} }) => {
  const { attemptRecovery, isRecovering } = useErrorRecovery()
  const { reportError } = useErrorReporting()

  useEffect(() => {
    if (error) {
      reportError(error, { component: 'ErrorRecovery' })
      
      // Attempt automatic recovery
      const errorType = error.name || 'UnknownError'
      const strategy = recoveryStrategies[errorType]
      
      if (strategy) {
        attemptRecovery(errorType, strategy)
      }
    }
  }, [error, attemptRecovery, recoveryStrategies, reportError])

  if (error) {
    if (isRecovering) {
      return (
        <div className="flex items-center justify-center p-8">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900 mx-auto"></div>
            <p className="mt-2 text-sm text-gray-600">Attempting to recover...</p>
          </div>
        </div>
      )
    }
    
    return fallback || <div>Something went wrong</div>
  }

  return children
}

export default ErrorRecovery

// utils/recoveryStrategies.js
export const recoveryStrategies = {
  ChunkLoadError: async () => {
    // Strategy: Reload the page to fetch fresh chunks
    window.location.reload()
  },
  
  NetworkError: async () => {
    // Strategy: Wait and retry
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    // Check if network is back
    if (navigator.onLine) {
      // Attempt to make a simple request
      await fetch('/api/v0/health')
    } else {
      throw new Error('Network still unavailable')
    }
  },
  
  TokenExpiredError: async () => {
    // Strategy: Try to refresh the token
    const refreshToken = localStorage.getItem('refreshToken')
    if (refreshToken) {
      const response = await fetch('/api/v0/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken }),
      })
      
      if (response.ok) {
        const data = await response.json()
        localStorage.setItem('token', data.token)
        localStorage.setItem('refreshToken', data.refreshToken)
      } else {
        // Redirect to login
        window.location.href = '/login'
      }
    } else {
      window.location.href = '/login'
    }
  },
  
  StateCorruptionError: async () => {
    // Strategy: Clear corrupted state and reload
    localStorage.clear()
    sessionStorage.clear()
    window.location.reload()
  },
}
```

## Directory Structure
```
src/
├── components/
│   ├── errors/
│   │   ├── GlobalErrorBoundary.jsx
│   │   ├── AsyncErrorBoundary.jsx
│   │   ├── ErrorFallback.jsx
│   │   ├── ErrorMessage.jsx
│   │   ├── ErrorToast.jsx
│   │   └── ErrorRecovery.jsx
│   └── common/
│       ├── LoadingSpinner.jsx
│       └── RetryButton.jsx
├── hooks/
│   ├── useAPIError.js
│   ├── useErrorRecovery.js
│   ├── useErrorReporting.js
│   └── useNotification.js
├── services/
│   ├── errorHandler.js
│   ├── errorReporter.js
│   └── apiClient.js
├── utils/
│   ├── recoveryStrategies.js
│   └── errorClassification.js
└── types/
    └── errors.js
```

## Error Handling Features
- ✅ Global error boundaries with recovery
- ✅ API error handling with retry logic
- ✅ User-friendly error messages
- ✅ Error reporting and monitoring
- ✅ Automatic recovery mechanisms
- ✅ Error classification and routing
- ✅ Breadcrumb tracking
- ✅ Error context capture
- ✅ Toast notifications for errors
- ✅ Graceful degradation

## Benefits
- Improved user experience during errors
- Automatic error recovery
- Comprehensive error tracking
- Better debugging capabilities
- Reduced user frustration
- Increased application stability
- Better error visibility for developers
- Proactive error resolution