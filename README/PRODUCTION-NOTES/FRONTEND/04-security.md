# Security Implementation Plan

## Overview
Implement comprehensive frontend security measures to protect against common web vulnerabilities, secure user authentication, and ensure data protection throughout the application lifecycle.

## Current State Analysis
- Basic JWT token storage in localStorage
- Minimal input validation on frontend
- No Content Security Policy (CSP) implementation
- Basic authentication flow without advanced security features
- No protection against common frontend attacks
- Limited secure headers implementation
- No security monitoring or incident detection

## Implementation Steps

### Step 1: Authentication Security Hardening
**Timeline: 2-3 days**

Implement secure authentication practices:

```javascript
// utils/secureStorage.js
class SecureStorage {
  constructor() {
    this.prefix = 'sp_'
    this.encryptionKey = this.getOrCreateEncryptionKey()
  }

  getOrCreateEncryptionKey() {
    let key = sessionStorage.getItem(`${this.prefix}ek`)
    if (!key) {
      key = this.generateRandomKey()
      sessionStorage.setItem(`${this.prefix}ek`, key)
    }
    return key
  }

  generateRandomKey() {
    const array = new Uint8Array(32)
    crypto.getRandomValues(array)
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('')
  }

  // Simple XOR encryption for basic obfuscation
  encrypt(data) {
    const jsonString = JSON.stringify(data)
    let encrypted = ''
    for (let i = 0; i < jsonString.length; i++) {
      const keyChar = this.encryptionKey[i % this.encryptionKey.length]
      encrypted += String.fromCharCode(
        jsonString.charCodeAt(i) ^ keyChar.charCodeAt(0)
      )
    }
    return btoa(encrypted)
  }

  decrypt(encryptedData) {
    try {
      const encrypted = atob(encryptedData)
      let decrypted = ''
      for (let i = 0; i < encrypted.length; i++) {
        const keyChar = this.encryptionKey[i % this.encryptionKey.length]
        decrypted += String.fromCharCode(
          encrypted.charCodeAt(i) ^ keyChar.charCodeAt(0)
        )
      }
      return JSON.parse(decrypted)
    } catch (error) {
      console.error('Decryption failed:', error)
      return null
    }
  }

  setItem(key, value) {
    const encrypted = this.encrypt(value)
    localStorage.setItem(`${this.prefix}${key}`, encrypted)
  }

  getItem(key) {
    const encrypted = localStorage.getItem(`${this.prefix}${key}`)
    return encrypted ? this.decrypt(encrypted) : null
  }

  removeItem(key) {
    localStorage.removeItem(`${this.prefix}${key}`)
  }

  clear() {
    const keys = Object.keys(localStorage).filter(key => 
      key.startsWith(this.prefix)
    )
    keys.forEach(key => localStorage.removeItem(key))
    sessionStorage.removeItem(`${this.prefix}ek`)
  }
}

export const secureStorage = new SecureStorage()

// services/authSecurity.js
import { secureStorage } from '../utils/secureStorage'

export class AuthSecurity {
  constructor() {
    this.tokenKey = 'auth_token'
    this.refreshTokenKey = 'refresh_token'
    this.sessionTimeoutKey = 'session_timeout'
    this.maxSessionDuration = 8 * 60 * 60 * 1000 // 8 hours
    this.inactivityTimeout = 30 * 60 * 1000 // 30 minutes
    this.lastActivityKey = 'last_activity'
    
    this.startSessionMonitoring()
  }

  // Secure token storage
  storeTokens(token, refreshToken) {
    const tokenData = {
      token,
      refreshToken,
      timestamp: Date.now(),
      expires: Date.now() + this.maxSessionDuration,
    }
    
    secureStorage.setItem(this.tokenKey, tokenData)
    this.updateLastActivity()
  }

  getToken() {
    const tokenData = secureStorage.getItem(this.tokenKey)
    if (!tokenData) return null

    // Check token expiration
    if (Date.now() > tokenData.expires) {
      this.clearTokens()
      return null
    }

    // Check inactivity timeout
    if (this.isSessionInactive()) {
      this.clearTokens()
      return null
    }

    return tokenData.token
  }

  getRefreshToken() {
    const tokenData = secureStorage.getItem(this.tokenKey)
    return tokenData?.refreshToken || null
  }

  clearTokens() {
    secureStorage.removeItem(this.tokenKey)
    secureStorage.removeItem(this.lastActivityKey)
    this.broadcastLogout()
  }

  updateLastActivity() {
    secureStorage.setItem(this.lastActivityKey, Date.now())
  }

  isSessionInactive() {
    const lastActivity = secureStorage.getItem(this.lastActivityKey)
    if (!lastActivity) return true
    
    return Date.now() - lastActivity > this.inactivityTimeout
  }

  // Session monitoring
  startSessionMonitoring() {
    // Monitor user activity
    const activityEvents = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart']
    
    activityEvents.forEach(event => {
      document.addEventListener(event, this.handleUserActivity.bind(this), true)
    })

    // Check session validity periodically
    setInterval(() => {
      if (this.getToken() === null) {
        this.handleSessionExpiry()
      }
    }, 60000) // Check every minute

    // Handle page visibility changes
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden) {
        this.updateLastActivity()
      }
    })
  }

  handleUserActivity() {
    this.updateLastActivity()
  }

  handleSessionExpiry() {
    // Emit custom event for session expiry
    window.dispatchEvent(new CustomEvent('sessionExpired', {
      detail: { reason: this.isSessionInactive() ? 'inactivity' : 'timeout' }
    }))
  }

  // Cross-tab logout synchronization
  broadcastLogout() {
    localStorage.setItem('logout-event', Date.now().toString())
    localStorage.removeItem('logout-event')
  }

  // Fingerprinting for additional security
  generateFingerprint() {
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    ctx.textBaseline = 'top'
    ctx.font = '14px Arial'
    ctx.fillText('Browser fingerprint', 2, 2)
    
    const fingerprint = {
      userAgent: navigator.userAgent,
      language: navigator.language,
      platform: navigator.platform,
      screen: `${screen.width}x${screen.height}`,
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
      canvas: canvas.toDataURL(),
    }
    
    return btoa(JSON.stringify(fingerprint))
  }

  validateFingerprint(storedFingerprint) {
    const currentFingerprint = this.generateFingerprint()
    return currentFingerprint === storedFingerprint
  }
}

export const authSecurity = new AuthSecurity()
```

### Step 2: Input Validation and Sanitization
**Timeline: 2 days**

Implement comprehensive input validation:

```javascript
// utils/validation.js
import DOMPurify from 'dompurify'

export class InputValidator {
  constructor() {
    this.patterns = {
      email: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
      username: /^[a-zA-Z0-9_]{3,20}$/,
      password: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/,
      amount: /^\d+(\.\d{1,2})?$/,
      marketTitle: /^[a-zA-Z0-9\s\-.,!?]{5,100}$/,
      url: /^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$/,
    }

    this.maxLengths = {
      username: 20,
      email: 254,
      password: 128,
      marketTitle: 100,
      marketDescription: 1000,
      comment: 500,
    }
  }

  validate(field, value, rules = []) {
    const errors = []

    // Check if field is required
    if (rules.includes('required') && (!value || value.trim() === '')) {
      errors.push(`${field} is required`)
      return errors
    }

    if (!value) return errors

    // Type-specific validation
    if (rules.includes('email') && !this.patterns.email.test(value)) {
      errors.push('Please enter a valid email address')
    }

    if (rules.includes('username') && !this.patterns.username.test(value)) {
      errors.push('Username must be 3-20 characters, letters, numbers, and underscores only')
    }

    if (rules.includes('password') && !this.patterns.password.test(value)) {
      errors.push('Password must be at least 8 characters with uppercase, lowercase, number, and special character')
    }

    if (rules.includes('amount')) {
      if (!this.patterns.amount.test(value)) {
        errors.push('Please enter a valid amount')
      } else {
        const numValue = parseFloat(value)
        if (numValue <= 0) {
          errors.push('Amount must be greater than 0')
        }
        if (numValue > 10000) {
          errors.push('Amount cannot exceed $10,000')
        }
      }
    }

    if (rules.includes('url') && !this.patterns.url.test(value)) {
      errors.push('Please enter a valid URL')
    }

    // Length validation
    const maxLength = this.maxLengths[field]
    if (maxLength && value.length > maxLength) {
      errors.push(`${field} cannot exceed ${maxLength} characters`)
    }

    // Custom validation rules
    if (rules.includes('noScript') && this.containsScript(value)) {
      errors.push('Script tags are not allowed')
    }

    if (rules.includes('noHtml') && this.containsHtml(value)) {
      errors.push('HTML tags are not allowed')
    }

    return errors
  }

  sanitize(value, options = {}) {
    if (typeof value !== 'string') return value

    const {
      allowHtml = false,
      stripHtml = false,
      escapeHtml = true,
      maxLength = null,
    } = options

    let sanitized = value

    // Remove or escape HTML
    if (stripHtml) {
      sanitized = sanitized.replace(/<[^>]*>/g, '')
    } else if (allowHtml) {
      sanitized = DOMPurify.sanitize(sanitized)
    } else if (escapeHtml) {
      sanitized = this.escapeHtml(sanitized)
    }

    // Trim whitespace
    sanitized = sanitized.trim()

    // Limit length
    if (maxLength && sanitized.length > maxLength) {
      sanitized = sanitized.substring(0, maxLength)
    }

    return sanitized
  }

  escapeHtml(text) {
    const div = document.createElement('div')
    div.textContent = text
    return div.innerHTML
  }

  containsScript(value) {
    const scriptPattern = /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi
    return scriptPattern.test(value)
  }

  containsHtml(value) {
    const htmlPattern = /<[^>]+>/g
    return htmlPattern.test(value)
  }

  validateForm(formData, schema) {
    const errors = {}
    let isValid = true

    Object.keys(schema).forEach(field => {
      const value = formData[field]
      const rules = schema[field]
      const fieldErrors = this.validate(field, value, rules)
      
      if (fieldErrors.length > 0) {
        errors[field] = fieldErrors
        isValid = false
      }
    })

    return { isValid, errors }
  }
}

export const inputValidator = new InputValidator()

// hooks/useFormValidation.js
import { useState, useCallback } from 'react'
import { inputValidator } from '../utils/validation'

export const useFormValidation = (schema, sanitizeOptions = {}) => {
  const [errors, setErrors] = useState({})
  const [touched, setTouched] = useState({})

  const validateField = useCallback((field, value) => {
    if (!schema[field]) return []
    
    const fieldErrors = inputValidator.validate(field, value, schema[field])
    setErrors(prev => ({
      ...prev,
      [field]: fieldErrors.length > 0 ? fieldErrors : undefined,
    }))
    
    return fieldErrors
  }, [schema])

  const validateForm = useCallback((formData) => {
    const validation = inputValidator.validateForm(formData, schema)
    setErrors(validation.errors)
    return validation
  }, [schema])

  const sanitizeValue = useCallback((field, value) => {
    const options = sanitizeOptions[field] || sanitizeOptions.default || {}
    return inputValidator.sanitize(value, options)
  }, [sanitizeOptions])

  const handleFieldBlur = useCallback((field) => {
    setTouched(prev => ({ ...prev, [field]: true }))
  }, [])

  const clearErrors = useCallback(() => {
    setErrors({})
    setTouched({})
  }, [])

  return {
    errors,
    touched,
    validateField,
    validateForm,
    sanitizeValue,
    handleFieldBlur,
    clearErrors,
  }
}
```

### Step 3: Content Security Policy (CSP)
**Timeline: 1-2 days**

Implement CSP and security headers:

```javascript
// public/index.html - Add CSP meta tag
<meta http-equiv="Content-Security-Policy" content="
  default-src 'self';
  script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.googletagmanager.com https://www.google-analytics.com;
  style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
  font-src 'self' https://fonts.gstatic.com;
  img-src 'self' data: https: blob:;
  connect-src 'self' https://api.socialpredict.com wss://api.socialpredict.com;
  media-src 'self' data: blob:;
  object-src 'none';
  base-uri 'self';
  form-action 'self';
  frame-ancestors 'none';
  upgrade-insecure-requests;
">

// utils/securityHeaders.js - For server-side implementation
export const generateSecurityHeaders = () => ({
  'Content-Security-Policy': [
    "default-src 'self'",
    "script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.googletagmanager.com",
    "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
    "font-src 'self' https://fonts.gstatic.com",
    "img-src 'self' data: https: blob:",
    "connect-src 'self' https://api.socialpredict.com wss://api.socialpredict.com",
    "media-src 'self' data: blob:",
    "object-src 'none'",
    "base-uri 'self'",
    "form-action 'self'",
    "frame-ancestors 'none'",
    "upgrade-insecure-requests",
  ].join('; '),
  'X-Content-Type-Options': 'nosniff',
  'X-Frame-Options': 'DENY',
  'X-XSS-Protection': '1; mode=block',
  'Referrer-Policy': 'strict-origin-when-cross-origin',
  'Permissions-Policy': [
    'camera=()',
    'microphone=()',
    'geolocation=()',
    'payment=()',
    'usb=()',
  ].join(', '),
  'Strict-Transport-Security': 'max-age=31536000; includeSubDomains; preload',
})

// utils/cspReporting.js - CSP violation reporting
export class CSPReporter {
  constructor() {
    this.reportEndpoint = '/api/v0/security/csp-violations'
    this.setupViolationReporting()
  }

  setupViolationReporting() {
    document.addEventListener('securitypolicyviolation', (event) => {
      this.reportViolation({
        blockedURI: event.blockedURI,
        columnNumber: event.columnNumber,
        disposition: event.disposition,
        documentURI: event.documentURI,
        effectiveDirective: event.effectiveDirective,
        lineNumber: event.lineNumber,
        originalPolicy: event.originalPolicy,
        referrer: event.referrer,
        sample: event.sample,
        sourceFile: event.sourceFile,
        statusCode: event.statusCode,
        violatedDirective: event.violatedDirective,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent,
      })
    })
  }

  async reportViolation(violation) {
    try {
      await fetch(this.reportEndpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(violation),
      })
    } catch (error) {
      console.error('Failed to report CSP violation:', error)
    }
  }
}

export const cspReporter = new CSPReporter()
```

### Step 4: XSS Protection
**Timeline: 2 days**

Implement comprehensive XSS protection:

```javascript
// utils/xssProtection.js
import DOMPurify from 'dompurify'

export class XSSProtection {
  constructor() {
    this.config = {
      ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'p', 'br', 'ul', 'ol', 'li'],
      ALLOWED_ATTR: ['class'],
      KEEP_CONTENT: true,
      RETURN_DOM: false,
      RETURN_DOM_FRAGMENT: false,
      RETURN_DOM_IMPORT: false,
      SANITIZE_DOM: true,
      WHOLE_DOCUMENT: false,
      FORBID_TAGS: ['script', 'object', 'embed', 'link', 'style', 'img', 'svg'],
      FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover', 'style'],
    }
    
    this.setupDOMPurify()
  }

  setupDOMPurify() {
    // Add custom hooks
    DOMPurify.addHook('beforeSanitizeElements', (node) => {
      // Block data URLs in src attributes
      if (node.getAttribute && node.getAttribute('src')) {
        const src = node.getAttribute('src')
        if (src.startsWith('data:') || src.startsWith('javascript:')) {
          node.removeAttribute('src')
        }
      }
    })

    DOMPurify.addHook('beforeSanitizeAttributes', (node) => {
      // Remove dangerous attributes
      const dangerousAttrs = ['onclick', 'onload', 'onerror', 'onmouseover', 'onfocus']
      dangerousAttrs.forEach(attr => {
        if (node.hasAttribute(attr)) {
          node.removeAttribute(attr)
        }
      })
    })
  }

  sanitizeHTML(html, options = {}) {
    const config = { ...this.config, ...options }
    return DOMPurify.sanitize(html, config)
  }

  sanitizeText(text) {
    // For plain text, escape HTML entities
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#x27;')
      .replace(/\//g, '&#x2F;')
  }

  validateUserInput(input, context = 'general') {
    const patterns = {
      general: [
        /<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi,
        /javascript:/gi,
        /vbscript:/gi,
        /data:text\/html/gi,
        /on\w+\s*=/gi,
      ],
      url: [
        /javascript:/gi,
        /vbscript:/gi,
        /data:(?!image\/)/gi,
      ],
      css: [
        /expression\s*\(/gi,
        /javascript:/gi,
        /vbscript:/gi,
        /@import/gi,
        /binding:/gi,
      ],
    }

    const contextPatterns = patterns[context] || patterns.general
    
    for (const pattern of contextPatterns) {
      if (pattern.test(input)) {
        return false
      }
    }
    
    return true
  }

  createSafeHTML(html) {
    const sanitized = this.sanitizeHTML(html)
    return { __html: sanitized }
  }

  // Safe innerHTML replacement
  safeInnerHTML(element, html) {
    const sanitized = this.sanitizeHTML(html)
    element.innerHTML = sanitized
  }

  // Safe attribute setting
  safeSetAttribute(element, attribute, value) {
    const dangerousAttrs = ['onclick', 'onload', 'onerror', 'onmouseover', 'onfocus', 'href']
    
    if (dangerousAttrs.includes(attribute.toLowerCase())) {
      if (attribute.toLowerCase() === 'href' && !this.validateUserInput(value, 'url')) {
        return false
      }
      if (attribute.toLowerCase().startsWith('on')) {
        return false
      }
    }
    
    element.setAttribute(attribute, value)
    return true
  }
}

export const xssProtection = new XSSProtection()

// components/common/SafeHTML.jsx - Safe HTML rendering component
import React from 'react'
import { xssProtection } from '../../utils/xssProtection'

const SafeHTML = ({ 
  html, 
  allowedTags = [], 
  allowedAttributes = [], 
  className = '',
  tag = 'div' 
}) => {
  const config = {
    ALLOWED_TAGS: allowedTags.length > 0 ? allowedTags : undefined,
    ALLOWED_ATTR: allowedAttributes.length > 0 ? allowedAttributes : undefined,
  }

  const sanitizedHTML = xssProtection.sanitizeHTML(html, config)
  const TagName = tag

  return (
    <TagName 
      className={className}
      dangerouslySetInnerHTML={{ __html: sanitizedHTML }}
    />
  )
}

export default SafeHTML
```

### Step 5: Secure API Communication
**Timeline: 2-3 days**

Implement secure API communication:

```javascript
// services/secureAPI.js
import { authSecurity } from './authSecurity'

export class SecureAPIClient {
  constructor() {
    this.baseURL = process.env.REACT_APP_API_URL || '/api/v0'
    this.requestInterceptors = []
    this.responseInterceptors = []
    this.maxRetries = 3
    this.retryDelay = 1000
    
    this.setupInterceptors()
  }

  setupInterceptors() {
    // Request interceptor for authentication
    this.addRequestInterceptor((config) => {
      const token = authSecurity.getToken()
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      
      // Add fingerprint for additional security
      config.headers['X-Client-Fingerprint'] = authSecurity.generateFingerprint()
      
      // Add CSRF token if available
      const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content')
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken
      }
      
      return config
    })

    // Response interceptor for token refresh
    this.addResponseInterceptor(
      (response) => response,
      async (error) => {
        const originalRequest = error.config
        
        if (error.response?.status === 401 && !originalRequest._retry) {
          originalRequest._retry = true
          
          const refreshToken = authSecurity.getRefreshToken()
          if (refreshToken) {
            try {
              const newToken = await this.refreshAccessToken(refreshToken)
              if (newToken) {
                originalRequest.headers.Authorization = `Bearer ${newToken}`
                return this.request(originalRequest)
              }
            } catch (refreshError) {
              authSecurity.clearTokens()
              window.location.href = '/login'
            }
          } else {
            authSecurity.clearTokens()
            window.location.href = '/login'
          }
        }
        
        return Promise.reject(error)
      }
    )
  }

  addRequestInterceptor(interceptor) {
    this.requestInterceptors.push(interceptor)
  }

  addResponseInterceptor(onFulfilled, onRejected) {
    this.responseInterceptors.push({ onFulfilled, onRejected })
  }

  async request(config) {
    // Apply request interceptors
    let processedConfig = { ...config }
    for (const interceptor of this.requestInterceptors) {
      processedConfig = await interceptor(processedConfig)
    }

    // Add default headers
    processedConfig.headers = {
      'Content-Type': 'application/json',
      'X-Requested-With': 'XMLHttpRequest',
      ...processedConfig.headers,
    }

    let attempt = 0
    while (attempt < this.maxRetries) {
      try {
        const response = await fetch(`${this.baseURL}${processedConfig.url}`, {
          method: processedConfig.method || 'GET',
          headers: processedConfig.headers,
          body: processedConfig.data ? JSON.stringify(processedConfig.data) : undefined,
          credentials: 'same-origin',
          ...processedConfig,
        })

        // Apply response interceptors
        let processedResponse = response
        for (const interceptor of this.responseInterceptors) {
          try {
            processedResponse = await interceptor.onFulfilled(processedResponse)
          } catch (error) {
            if (interceptor.onRejected) {
              processedResponse = await interceptor.onRejected(error)
            } else {
              throw error
            }
          }
        }

        // Parse JSON response
        if (processedResponse.headers.get('content-type')?.includes('application/json')) {
          const data = await processedResponse.json()
          return { ...processedResponse, data }
        }

        return processedResponse
        
      } catch (error) {
        attempt++
        
        if (attempt >= this.maxRetries) {
          throw error
        }
        
        // Exponential backoff
        await this.delay(this.retryDelay * Math.pow(2, attempt - 1))
      }
    }
  }

  async refreshAccessToken(refreshToken) {
    try {
      const response = await fetch(`${this.baseURL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refreshToken }),
      })

      if (response.ok) {
        const data = await response.json()
        authSecurity.storeTokens(data.token, data.refreshToken)
        return data.token
      }
      
      return null
    } catch (error) {
      console.error('Token refresh failed:', error)
      return null
    }
  }

  delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  // HTTP methods
  get(url, config = {}) {
    return this.request({ ...config, method: 'GET', url })
  }

  post(url, data, config = {}) {
    return this.request({ ...config, method: 'POST', url, data })
  }

  put(url, data, config = {}) {
    return this.request({ ...config, method: 'PUT', url, data })
  }

  patch(url, data, config = {}) {
    return this.request({ ...config, method: 'PATCH', url, data })
  }

  delete(url, config = {}) {
    return this.request({ ...config, method: 'DELETE', url })
  }
}

export const secureAPI = new SecureAPIClient()

// utils/requestSigning.js - Request signing for critical operations
import CryptoJS from 'crypto-js'

export class RequestSigner {
  constructor() {
    this.secretKey = process.env.REACT_APP_SIGNING_KEY || 'default-key'
  }

  signRequest(method, url, body, timestamp) {
    const message = `${method}|${url}|${body || ''}|${timestamp}`
    return CryptoJS.HmacSHA256(message, this.secretKey).toString()
  }

  signCriticalRequest(config) {
    const timestamp = Date.now()
    const signature = this.signRequest(
      config.method,
      config.url,
      config.data ? JSON.stringify(config.data) : '',
      timestamp
    )

    return {
      ...config,
      headers: {
        ...config.headers,
        'X-Timestamp': timestamp,
        'X-Signature': signature,
      },
    }
  }
}

export const requestSigner = new RequestSigner()
```

### Step 6: Security Monitoring
**Timeline: 2 days**

Implement security monitoring and incident detection:

```javascript
// utils/securityMonitor.js
export class SecurityMonitor {
  constructor() {
    this.incidents = []
    this.thresholds = {
      loginAttempts: 5,
      apiCalls: 100,
      formSubmissions: 20,
      timeWindow: 15 * 60 * 1000, // 15 minutes
    }
    
    this.counters = new Map()
    this.setupMonitoring()
  }

  setupMonitoring() {
    // Monitor failed login attempts
    window.addEventListener('loginFailed', (event) => {
      this.recordIncident('login_failed', {
        username: event.detail.username,
        reason: event.detail.reason,
        ip: event.detail.ip,
      })
    })

    // Monitor suspicious API calls
    window.addEventListener('apiError', (event) => {
      if (event.detail.status === 403 || event.detail.status === 401) {
        this.recordIncident('unauthorized_access', {
          endpoint: event.detail.endpoint,
          status: event.detail.status,
        })
      }
    })

    // Monitor DOM manipulation attempts
    this.setupDOMMonitoring()
    
    // Monitor console access
    this.setupConsoleMonitoring()
    
    // Monitor form submissions
    this.setupFormMonitoring()
  }

  setupDOMMonitoring() {
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.type === 'childList') {
          mutation.addedNodes.forEach((node) => {
            if (node.nodeType === Node.ELEMENT_NODE) {
              // Check for suspicious script injections
              if (node.tagName === 'SCRIPT' || 
                  node.innerHTML?.includes('<script') ||
                  node.innerHTML?.includes('javascript:')) {
                this.recordIncident('dom_manipulation', {
                  element: node.tagName,
                  content: node.innerHTML?.substring(0, 100),
                })
              }
            }
          })
        }
      })
    })

    observer.observe(document.body, {
      childList: true,
      subtree: true,
    })
  }

  setupConsoleMonitoring() {
    const originalConsole = { ...console }
    
    // Monitor console.log overrides (potential debugging attempts)
    Object.keys(originalConsole).forEach(method => {
      const original = originalConsole[method]
      console[method] = (...args) => {
        // Check for suspicious console usage patterns
        const message = args.join(' ')
        if (message.includes('password') || 
            message.includes('token') || 
            message.includes('secret')) {
          this.recordIncident('console_access', {
            method,
            message: message.substring(0, 100),
          })
        }
        
        return original.apply(console, args)
      }
    })
  }

  setupFormMonitoring() {
    document.addEventListener('submit', (event) => {
      const form = event.target
      const formData = new FormData(form)
      
      // Check for suspicious form data
      for (const [key, value] of formData.entries()) {
        if (typeof value === 'string') {
          if (value.includes('<script') || 
              value.includes('javascript:') ||
              value.includes('data:text/html')) {
            this.recordIncident('form_injection', {
              field: key,
              value: value.substring(0, 100),
              form: form.action || window.location.pathname,
            })
          }
        }
      }
      
      this.incrementCounter('form_submissions')
    })
  }

  recordIncident(type, details) {
    const incident = {
      id: this.generateId(),
      type,
      details,
      timestamp: new Date().toISOString(),
      url: window.location.href,
      userAgent: navigator.userAgent,
      fingerprint: this.getBrowserFingerprint(),
    }

    this.incidents.push(incident)
    this.reportIncident(incident)
    
    // Check if incident triggers threshold
    this.checkThresholds(type)
  }

  incrementCounter(type) {
    const now = Date.now()
    const key = `${type}_${Math.floor(now / this.thresholds.timeWindow)}`
    
    this.counters.set(key, (this.counters.get(key) || 0) + 1)
    
    // Clean old counters
    for (const [counterKey, value] of this.counters.entries()) {
      const timestamp = parseInt(counterKey.split('_').pop()) * this.thresholds.timeWindow
      if (now - timestamp > this.thresholds.timeWindow) {
        this.counters.delete(counterKey)
      }
    }
  }

  checkThresholds(type) {
    const recentIncidents = this.incidents.filter(incident => 
      incident.type === type &&
      Date.now() - new Date(incident.timestamp).getTime() < this.thresholds.timeWindow
    )

    const threshold = this.thresholds[type] || 10
    if (recentIncidents.length >= threshold) {
      this.triggerSecurityAlert(type, recentIncidents.length)
    }
  }

  triggerSecurityAlert(type, count) {
    const alert = {
      type: 'security_alert',
      alertType: type,
      count,
      timestamp: new Date().toISOString(),
      severity: this.getSeverity(type, count),
    }

    this.reportIncident(alert)
    
    // Take automatic protective actions
    this.takeProtectiveAction(type, count)
  }

  takeProtectiveAction(type, count) {
    switch (type) {
      case 'login_failed':
        if (count >= this.thresholds.loginAttempts) {
          // Temporarily disable login
          this.disableLogin(5 * 60 * 1000) // 5 minutes
        }
        break
        
      case 'dom_manipulation':
        // Clear potentially injected content
        this.sanitizePage()
        break
        
      case 'form_injection':
        // Disable form submissions temporarily
        this.disableForms(2 * 60 * 1000) // 2 minutes
        break
    }
  }

  async reportIncident(incident) {
    try {
      await fetch('/api/v0/security/incidents', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(incident),
      })
    } catch (error) {
      console.error('Failed to report security incident:', error)
    }
  }

  disableLogin(duration) {
    const loginButton = document.querySelector('[data-testid="login-button"]')
    if (loginButton) {
      loginButton.disabled = true
      setTimeout(() => {
        loginButton.disabled = false
      }, duration)
    }
  }

  disableForms(duration) {
    const forms = document.querySelectorAll('form')
    forms.forEach(form => {
      form.style.pointerEvents = 'none'
      setTimeout(() => {
        form.style.pointerEvents = ''
      }, duration)
    })
  }

  sanitizePage() {
    // Remove potentially dangerous elements
    const dangerousElements = document.querySelectorAll('script[src*="data:"], script[src*="javascript:"]')
    dangerousElements.forEach(el => el.remove())
  }

  getSeverity(type, count) {
    const severityMap = {
      login_failed: count >= 10 ? 'high' : 'medium',
      dom_manipulation: 'high',
      form_injection: 'high',
      unauthorized_access: 'medium',
      console_access: 'low',
    }
    
    return severityMap[type] || 'low'
  }

  generateId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2)
  }

  getBrowserFingerprint() {
    return btoa(JSON.stringify({
      userAgent: navigator.userAgent,
      language: navigator.language,
      platform: navigator.platform,
      screen: `${screen.width}x${screen.height}`,
    }))
  }

  getIncidents(type = null) {
    return type ? 
      this.incidents.filter(incident => incident.type === type) :
      this.incidents
  }
}

export const securityMonitor = new SecurityMonitor()
```

## Directory Structure
```
src/
├── utils/
│   ├── secureStorage.js         # Secure token storage
│   ├── validation.js            # Input validation
│   ├── xssProtection.js         # XSS protection
│   ├── securityHeaders.js       # Security headers
│   ├── cspReporting.js          # CSP violation reporting
│   ├── securityMonitor.js       # Security monitoring
│   └── requestSigning.js        # Request signing
├── services/
│   ├── authSecurity.js          # Authentication security
│   ├── secureAPI.js             # Secure API client
│   └── securityReporting.js     # Security incident reporting
├── hooks/
│   ├── useFormValidation.js     # Form validation hook
│   ├── useSecureAuth.js         # Secure authentication hook
│   └── useSecurityMonitor.js    # Security monitoring hook
├── components/
│   ├── common/
│   │   ├── SafeHTML.jsx         # Safe HTML rendering
│   │   ├── SecureForm.jsx       # Secure form component
│   │   └── SecurityAlert.jsx    # Security alert component
│   └── security/
│       ├── CSPViolationHandler.jsx
│       └── SecurityProvider.jsx
└── middleware/
    ├── securityMiddleware.js    # Security middleware
    └── rateLimiting.js          # Client-side rate limiting
```

## Security Checklist
- ✅ Secure token storage with encryption
- ✅ Session management with timeout and inactivity detection
- ✅ Comprehensive input validation and sanitization
- ✅ XSS protection with DOMPurify
- ✅ Content Security Policy implementation
- ✅ Secure API communication with retry and signing
- ✅ Security monitoring and incident detection
- ✅ CSRF protection
- ✅ Secure headers implementation
- ✅ Browser fingerprinting for additional security

## Benefits
- Protection against XSS attacks
- Secure authentication and session management
- Input validation and sanitization
- CSP violation monitoring
- Incident detection and response
- Secure API communication
- Protection against CSRF attacks
- Browser security hardening
- Security monitoring and alerting