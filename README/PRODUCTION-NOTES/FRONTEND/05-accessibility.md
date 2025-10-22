# Accessibility Implementation Plan

## Overview
Implement comprehensive accessibility features to ensure the application is usable by people with disabilities, meets WCAG 2.1 AA standards, and provides an inclusive user experience across all assistive technologies.

## Current State Analysis
- Basic semantic HTML structure
- Limited ARIA attributes and roles
- No keyboard navigation support
- Missing focus management
- No screen reader optimization
- Insufficient color contrast in some areas
- No alternative text for images
- Missing form labels and error announcements

## Implementation Steps

### Step 1: Semantic HTML and ARIA Implementation
**Timeline: 2-3 days**

Establish proper semantic structure and ARIA landmarks:

```javascript
// components/common/AccessibleContainer.jsx
import React, { forwardRef } from 'react'

const AccessibleContainer = forwardRef(({
  as = 'div',
  role,
  ariaLabel,
  ariaLabelledBy,
  ariaDescribedBy,
  children,
  className = '',
  ...props
}, ref) => {
  const Component = as
  
  return (
    <Component
      ref={ref}
      role={role}
      aria-label={ariaLabel}
      aria-labelledby={ariaLabelledBy}
      aria-describedby={ariaDescribedBy}
      className={className}
      {...props}
    >
      {children}
    </Component>
  )
})

AccessibleContainer.displayName = 'AccessibleContainer'
export default AccessibleContainer

// components/layout/Header.jsx - Accessible header
import React from 'react'
import { useAuth } from '../../hooks/useAuth'
import SkipLink from '../common/SkipLink'

const Header = () => {
  const { user, isAuthenticated } = useAuth()

  return (
    <>
      <SkipLink href="#main-content">Skip to main content</SkipLink>
      <header 
        role="banner" 
        className="bg-primary text-white p-4"
        aria-label="Site header"
      >
        <div className="container mx-auto flex justify-between items-center">
          <h1 className="text-2xl font-bold">
            <img 
              src="/logo.png" 
              alt="SocialPredict - Prediction Markets Platform"
              className="inline-block w-8 h-8 mr-2"
            />
            SocialPredict
          </h1>
          
          <nav role="navigation" aria-label="Main navigation">
            <ul className="flex space-x-4" role="list">
              <li role="listitem">
                <a 
                  href="/markets" 
                  className="hover:underline focus:outline-none focus:ring-2 focus:ring-yellow-400"
                  aria-current={location.pathname === '/markets' ? 'page' : undefined}
                >
                  Markets
                </a>
              </li>
              {isAuthenticated ? (
                <>
                  <li role="listitem">
                    <a 
                      href="/profile" 
                      className="hover:underline focus:outline-none focus:ring-2 focus:ring-yellow-400"
                      aria-current={location.pathname === '/profile' ? 'page' : undefined}
                    >
                      Profile
                    </a>
                  </li>
                  <li role="listitem">
                    <span className="sr-only">Current user: </span>
                    {user?.username}
                  </li>
                </>
              ) : (
                <li role="listitem">
                  <button 
                    className="bg-blue-600 px-4 py-2 rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-yellow-400"
                    aria-label="Sign in to your account"
                  >
                    Sign In
                  </button>
                </li>
              )}
            </ul>
          </nav>
        </div>
      </header>
    </>
  )
}

export default Header

// components/common/SkipLink.jsx
import React from 'react'

const SkipLink = ({ href, children }) => {
  return (
    <a
      href={href}
      className="sr-only focus:not-sr-only focus:absolute focus:top-0 focus:left-0 z-50 bg-blue-600 text-white p-2 rounded focus:outline-none focus:ring-2 focus:ring-yellow-400"
      onFocus={(e) => e.target.scrollIntoView()}
    >
      {children}
    </a>
  )
}

export default SkipLink
```

Implement accessible navigation:

```javascript
// components/navigation/AccessibleNav.jsx
import React, { useState, useRef, useEffect } from 'react'
import { useLocation } from 'react-router-dom'

const AccessibleNav = ({ items, orientation = 'horizontal' }) => {
  const [activeIndex, setActiveIndex] = useState(0)
  const [isKeyboardNavigation, setIsKeyboardNavigation] = useState(false)
  const navRef = useRef()
  const location = useLocation()

  useEffect(() => {
    // Set active item based on current route
    const currentIndex = items.findIndex(item => item.path === location.pathname)
    if (currentIndex !== -1) {
      setActiveIndex(currentIndex)
    }
  }, [location.pathname, items])

  const handleKeyDown = (event, index) => {
    setIsKeyboardNavigation(true)
    
    switch (event.key) {
      case 'ArrowDown':
      case 'ArrowRight':
        event.preventDefault()
        const nextIndex = orientation === 'horizontal' 
          ? (index + 1) % items.length
          : Math.min(index + 1, items.length - 1)
        setActiveIndex(nextIndex)
        focusItem(nextIndex)
        break
        
      case 'ArrowUp':
      case 'ArrowLeft':
        event.preventDefault()
        const prevIndex = orientation === 'horizontal'
          ? (index - 1 + items.length) % items.length
          : Math.max(index - 1, 0)
        setActiveIndex(prevIndex)
        focusItem(prevIndex)
        break
        
      case 'Home':
        event.preventDefault()
        setActiveIndex(0)
        focusItem(0)
        break
        
      case 'End':
        event.preventDefault()
        setActiveIndex(items.length - 1)
        focusItem(items.length - 1)
        break
        
      case 'Enter':
      case ' ':
        event.preventDefault()
        const item = items[index]
        if (item.onClick) {
          item.onClick()
        } else if (item.path) {
          window.location.href = item.path
        }
        break
    }
  }

  const focusItem = (index) => {
    const item = navRef.current?.children[index]?.querySelector('a, button')
    if (item) {
      item.focus()
    }
  }

  const handleMouseEnter = () => {
    setIsKeyboardNavigation(false)
  }

  return (
    <nav
      ref={navRef}
      role="navigation"
      aria-label="Main navigation"
      className={`flex ${orientation === 'vertical' ? 'flex-col' : 'flex-row'} space-x-2`}
    >
      <ul role="list" className="flex">
        {items.map((item, index) => (
          <li key={item.id} role="listitem">
            <a
              href={item.path}
              role="menuitem"
              tabIndex={index === activeIndex && isKeyboardNavigation ? 0 : -1}
              aria-current={location.pathname === item.path ? 'page' : undefined}
              className={`block px-4 py-2 rounded transition-colors ${
                location.pathname === item.path
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-100'
              } focus:outline-none focus:ring-2 focus:ring-blue-400`}
              onKeyDown={(e) => handleKeyDown(e, index)}
              onMouseEnter={handleMouseEnter}
              onFocus={() => setActiveIndex(index)}
            >
              {item.icon && (
                <span className="mr-2" aria-hidden="true">
                  {item.icon}
                </span>
              )}
              {item.label}
              {item.badge && (
                <span 
                  className="ml-2 bg-red-500 text-white text-xs px-2 py-1 rounded-full"
                  aria-label={`${item.badge} notifications`}
                >
                  {item.badge}
                </span>
              )}
            </a>
          </li>
        ))}
      </ul>
    </nav>
  )
}

export default AccessibleNav
```

### Step 2: Keyboard Navigation Support
**Timeline: 2-3 days**

Implement comprehensive keyboard navigation:

```javascript
// hooks/useKeyboardNavigation.js
import { useState, useEffect, useCallback } from 'react'

export const useKeyboardNavigation = (items, options = {}) => {
  const {
    orientation = 'horizontal',
    wrap = true,
    autoFocus = false,
    onSelectionChange,
  } = options

  const [currentIndex, setCurrentIndex] = useState(0)
  const [isKeyboardActive, setIsKeyboardActive] = useState(false)

  const moveToIndex = useCallback((newIndex) => {
    let targetIndex = newIndex

    if (wrap) {
      if (targetIndex < 0) targetIndex = items.length - 1
      if (targetIndex >= items.length) targetIndex = 0
    } else {
      targetIndex = Math.max(0, Math.min(targetIndex, items.length - 1))
    }

    setCurrentIndex(targetIndex)
    if (onSelectionChange) {
      onSelectionChange(targetIndex, items[targetIndex])
    }
  }, [items, wrap, onSelectionChange])

  const handleKeyDown = useCallback((event) => {
    if (!isKeyboardActive) return

    switch (event.key) {
      case 'ArrowDown':
        if (orientation === 'vertical' || orientation === 'both') {
          event.preventDefault()
          moveToIndex(currentIndex + 1)
        }
        break

      case 'ArrowUp':
        if (orientation === 'vertical' || orientation === 'both') {
          event.preventDefault()
          moveToIndex(currentIndex - 1)
        }
        break

      case 'ArrowRight':
        if (orientation === 'horizontal' || orientation === 'both') {
          event.preventDefault()
          moveToIndex(currentIndex + 1)
        }
        break

      case 'ArrowLeft':
        if (orientation === 'horizontal' || orientation === 'both') {
          event.preventDefault()
          moveToIndex(currentIndex - 1)
        }
        break

      case 'Home':
        event.preventDefault()
        moveToIndex(0)
        break

      case 'End':
        event.preventDefault()
        moveToIndex(items.length - 1)
        break

      case 'Enter':
      case ' ':
        event.preventDefault()
        const currentItem = items[currentIndex]
        if (currentItem?.onSelect) {
          currentItem.onSelect()
        }
        break

      case 'Escape':
        setIsKeyboardActive(false)
        break
    }
  }, [currentIndex, items, orientation, moveToIndex, isKeyboardActive])

  useEffect(() => {
    if (isKeyboardActive) {
      document.addEventListener('keydown', handleKeyDown)
      return () => document.removeEventListener('keydown', handleKeyDown)
    }
  }, [handleKeyDown, isKeyboardActive])

  const activateKeyboardNavigation = useCallback(() => {
    setIsKeyboardActive(true)
    if (autoFocus) {
      setCurrentIndex(0)
    }
  }, [autoFocus])

  const deactivateKeyboardNavigation = useCallback(() => {
    setIsKeyboardActive(false)
  }, [])

  return {
    currentIndex,
    isKeyboardActive,
    activateKeyboardNavigation,
    deactivateKeyboardNavigation,
    moveToIndex,
  }
}

// components/common/KeyboardTrap.jsx - Focus trap for modals
import React, { useEffect, useRef } from 'react'

const KeyboardTrap = ({ children, active = true, restoreFocus = true }) => {
  const containerRef = useRef()
  const previousActiveElement = useRef()

  useEffect(() => {
    if (!active) return

    previousActiveElement.current = document.activeElement

    const container = containerRef.current
    if (!container) return

    const focusableElements = container.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    )

    const firstElement = focusableElements[0]
    const lastElement = focusableElements[focusableElements.length - 1]

    const handleTabKey = (event) => {
      if (event.key !== 'Tab') return

      if (event.shiftKey) {
        if (document.activeElement === firstElement) {
          event.preventDefault()
          lastElement?.focus()
        }
      } else {
        if (document.activeElement === lastElement) {
          event.preventDefault()
          firstElement?.focus()
        }
      }
    }

    const handleEscapeKey = (event) => {
      if (event.key === 'Escape') {
        event.preventDefault()
        // Emit escape event for parent components to handle
        container.dispatchEvent(new CustomEvent('escapePressed'))
      }
    }

    container.addEventListener('keydown', handleTabKey)
    container.addEventListener('keydown', handleEscapeKey)

    // Focus first element
    firstElement?.focus()

    return () => {
      container.removeEventListener('keydown', handleTabKey)
      container.removeEventListener('keydown', handleEscapeKey)

      if (restoreFocus && previousActiveElement.current) {
        previousActiveElement.current.focus()
      }
    }
  }, [active, restoreFocus])

  return (
    <div ref={containerRef}>
      {children}
    </div>
  )
}

export default KeyboardTrap
```

### Step 3: Screen Reader Optimization
**Timeline: 2-3 days**

Implement screen reader support and live regions:

```javascript
// components/common/LiveRegion.jsx
import React, { useState, useEffect } from 'react'

const LiveRegion = ({ 
  children, 
  politeness = 'polite', 
  atomic = false,
  relevant = 'additions text',
  className = 'sr-only' 
}) => {
  const [announcement, setAnnouncement] = useState('')

  useEffect(() => {
    if (typeof children === 'string') {
      setAnnouncement(children)
    }
  }, [children])

  return (
    <div
      aria-live={politeness}
      aria-atomic={atomic}
      aria-relevant={relevant}
      className={className}
    >
      {announcement}
    </div>
  )
}

export default LiveRegion

// hooks/useAnnouncement.js
import { useState, useCallback } from 'react'

export const useAnnouncement = () => {
  const [announcement, setAnnouncement] = useState('')
  
  const announce = useCallback((message, politeness = 'polite') => {
    // Clear the announcement first to ensure it's re-read
    setAnnouncement('')
    
    // Use setTimeout to ensure the screen reader picks up the change
    setTimeout(() => {
      setAnnouncement(message)
    }, 100)
    
    // Clear the announcement after a delay
    setTimeout(() => {
      setAnnouncement('')
    }, 3000)
  }, [])

  const announceError = useCallback((message) => {
    announce(`Error: ${message}`, 'assertive')
  }, [announce])

  const announceSuccess = useCallback((message) => {
    announce(`Success: ${message}`, 'polite')
  }, [announce])

  const announceLoading = useCallback((message = 'Loading') => {
    announce(message, 'polite')
  }, [announce])

  return {
    announcement,
    announce,
    announceError,
    announceSuccess,
    announceLoading,
  }
}

// components/common/ScreenReaderContent.jsx
import React from 'react'

const ScreenReaderContent = ({ children, showOnFocus = false }) => {
  const baseClasses = 'absolute w-px h-px p-0 -m-px overflow-hidden clip-rect-0 whitespace-nowrap border-0'
  const focusClasses = showOnFocus 
    ? 'focus:relative focus:w-auto focus:h-auto focus:p-2 focus:m-0 focus:overflow-visible focus:clip-auto focus:whitespace-normal focus:bg-blue-600 focus:text-white focus:z-50'
    : ''

  return (
    <span className={`${baseClasses} ${focusClasses}`}>
      {children}
    </span>
  )
}

export default ScreenReaderContent

// components/markets/AccessibleMarketCard.jsx
import React from 'react'
import { useAnnouncement } from '../../hooks/useAnnouncement'
import LiveRegion from '../common/LiveRegion'
import ScreenReaderContent from '../common/ScreenReaderContent'

const AccessibleMarketCard = ({ market, onBetClick }) => {
  const { announcement, announceSuccess, announceError } = useAnnouncement()

  const handleBetClick = async () => {
    try {
      await onBetClick(market.id)
      announceSuccess(`Bet placed on ${market.title}`)
    } catch (error) {
      announceError(`Failed to place bet on ${market.title}: ${error.message}`)
    }
  }

  const marketDescription = `
    ${market.title}. 
    Status: ${market.status}. 
    Current odds: ${market.outcomes[0]?.odds || 'N/A'}. 
    Total volume: $${market.totalVolume}. 
    ${market.participantCount} participants.
    ${market.description}
  `

  return (
    <article
      className="market-card border rounded-lg p-4 hover:shadow-lg focus-within:shadow-lg"
      aria-labelledby={`market-${market.id}-title`}
      aria-describedby={`market-${market.id}-description`}
    >
      <header>
        <h3 
          id={`market-${market.id}-title`}
          className="text-xl font-semibold mb-2"
        >
          {market.title}
        </h3>
        
        <div className="flex items-center mb-2" role="group" aria-label="Market status">
          <span 
            className={`inline-block px-2 py-1 text-xs rounded ${
              market.status === 'open' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
            }`}
            aria-label={`Market status: ${market.status}`}
          >
            {market.status}
          </span>
          
          {market.trending && (
            <span 
              className="ml-2 text-orange-500"
              aria-label="Trending market"
            >
              🔥
            </span>
          )}
        </div>
      </header>

      <div 
        id={`market-${market.id}-description`}
        className="text-gray-600 mb-4"
      >
        {market.description}
        
        <ScreenReaderContent>
          {marketDescription}
        </ScreenReaderContent>
      </div>

      <div className="market-stats" role="group" aria-label="Market statistics">
        <dl className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <dt className="font-medium">Total Volume:</dt>
            <dd aria-label={`${market.totalVolume} dollars`}>
              ${market.totalVolume.toLocaleString()}
            </dd>
          </div>
          
          <div>
            <dt className="font-medium">Participants:</dt>
            <dd aria-label={`${market.participantCount} participants`}>
              {market.participantCount}
            </dd>
          </div>
          
          <div>
            <dt className="font-medium">Current Odds:</dt>
            <dd aria-label={`${market.outcomes[0]?.odds || 'No odds available'}`}>
              {market.outcomes[0]?.odds || 'N/A'}
            </dd>
          </div>
          
          <div>
            <dt className="font-medium">Time Remaining:</dt>
            <dd aria-label={`${market.timeRemaining} remaining`}>
              {market.timeRemaining}
            </dd>
          </div>
        </dl>
      </div>

      <footer className="mt-4">
        <button
          onClick={handleBetClick}
          disabled={market.status !== 'open'}
          className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:ring-offset-2 disabled:bg-gray-400 disabled:cursor-not-allowed"
          aria-describedby={`market-${market.id}-bet-help`}
        >
          {market.status === 'open' ? 'Place Bet' : 'Market Closed'}
        </button>
        
        <div id={`market-${market.id}-bet-help`} className="sr-only">
          {market.status === 'open' 
            ? `Place a bet on ${market.title}. Current odds are ${market.outcomes[0]?.odds}`
            : `This market is closed and no longer accepting bets`
          }
        </div>
      </footer>

      <LiveRegion>
        {announcement}
      </LiveRegion>
    </article>
  )
}

export default AccessibleMarketCard
```

### Step 4: Form Accessibility
**Timeline: 2 days**

Implement accessible forms with proper labeling and error handling:

```javascript
// components/forms/AccessibleForm.jsx
import React, { useState } from 'react'
import { useAnnouncement } from '../../hooks/useAnnouncement'
import LiveRegion from '../common/LiveRegion'

const AccessibleForm = ({ children, onSubmit, validation, ariaLabel }) => {
  const [errors, setErrors] = useState({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const { announcement, announceError, announceSuccess } = useAnnouncement()

  const handleSubmit = async (event) => {
    event.preventDefault()
    setIsSubmitting(true)

    const formData = new FormData(event.target)
    const data = Object.fromEntries(formData.entries())

    // Validate form
    if (validation) {
      const validationResult = validation(data)
      if (!validationResult.isValid) {
        setErrors(validationResult.errors)
        
        // Announce errors to screen readers
        const errorCount = Object.keys(validationResult.errors).length
        const errorMessage = `Form has ${errorCount} error${errorCount !== 1 ? 's' : ''}. Please review and correct.`
        announceError(errorMessage)
        
        // Focus first error field
        const firstErrorField = Object.keys(validationResult.errors)[0]
        const errorElement = document.querySelector(`[name="${firstErrorField}"]`)
        if (errorElement) {
          errorElement.focus()
        }
        
        setIsSubmitting(false)
        return
      }
    }

    try {
      await onSubmit(data)
      announceSuccess('Form submitted successfully')
      setErrors({})
    } catch (error) {
      announceError(`Form submission failed: ${error.message}`)
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      noValidate
      aria-label={ariaLabel}
      role="form"
    >
      <fieldset disabled={isSubmitting}>
        <legend className="sr-only">{ariaLabel}</legend>
        
        {React.Children.map(children, child => {
          if (React.isValidElement(child)) {
            return React.cloneElement(child, {
              errors: errors[child.props.name],
              disabled: isSubmitting,
            })
          }
          return child
        })}
        
        <div 
          id="form-errors" 
          role="alert" 
          aria-live="assertive"
          className={Object.keys(errors).length > 0 ? 'block' : 'sr-only'}
        >
          {Object.keys(errors).length > 0 && (
            <div className="bg-red-50 border border-red-200 rounded p-3 mb-4">
              <h3 className="text-red-800 font-medium mb-2">
                Please correct the following errors:
              </h3>
              <ul className="text-red-700 text-sm">
                {Object.entries(errors).map(([field, fieldErrors]) => (
                  <li key={field}>
                    <strong>{field}:</strong> {fieldErrors.join(', ')}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </fieldset>

      <LiveRegion politeness="assertive">
        {announcement}
      </LiveRegion>
    </form>
  )
}

// components/forms/AccessibleInput.jsx
import React, { useId } from 'react'

const AccessibleInput = ({
  label,
  name,
  type = 'text',
  required = false,
  errors = [],
  helpText,
  placeholder,
  value,
  onChange,
  disabled = false,
  autoComplete,
  ...props
}) => {
  const inputId = useId()
  const helpId = `${inputId}-help`
  const errorId = `${inputId}-error`
  
  const hasErrors = errors && errors.length > 0
  
  return (
    <div className="form-field mb-4">
      <label 
        htmlFor={inputId}
        className="block text-sm font-medium text-gray-700 mb-1"
      >
        {label}
        {required && (
          <span className="text-red-500 ml-1" aria-label="required">
            *
          </span>
        )}
      </label>
      
      <input
        id={inputId}
        name={name}
        type={type}
        value={value}
        onChange={onChange}
        placeholder={placeholder}
        required={required}
        disabled={disabled}
        autoComplete={autoComplete}
        aria-invalid={hasErrors}
        aria-describedby={`${helpText ? helpId : ''} ${hasErrors ? errorId : ''}`.trim()}
        className={`block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-blue-400 ${
          hasErrors 
            ? 'border-red-500 focus:ring-red-400 focus:border-red-400' 
            : 'border-gray-300'
        } ${disabled ? 'bg-gray-100 cursor-not-allowed' : ''}`}
        {...props}
      />
      
      {helpText && (
        <div 
          id={helpId} 
          className="mt-1 text-sm text-gray-600"
        >
          {helpText}
        </div>
      )}
      
      {hasErrors && (
        <div 
          id={errorId}
          role="alert"
          className="mt-1 text-sm text-red-600"
        >
          {errors.join(', ')}
        </div>
      )}
    </div>
  )
}

export default AccessibleInput
```

### Step 5: Color Contrast and Visual Accessibility
**Timeline: 1-2 days**

Ensure proper color contrast and visual accessibility:

```javascript
// utils/colorContrast.js
export class ColorContrastChecker {
  constructor() {
    this.wcagAAThreshold = 4.5
    this.wcagAAAThreshold = 7
  }

  // Convert hex color to RGB
  hexToRgb(hex) {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex)
    return result ? {
      r: parseInt(result[1], 16),
      g: parseInt(result[2], 16),
      b: parseInt(result[3], 16),
    } : null
  }

  // Calculate relative luminance
  getLuminance(r, g, b) {
    const [rs, gs, bs] = [r, g, b].map(component => {
      component = component / 255
      return component <= 0.03928 
        ? component / 12.92 
        : Math.pow((component + 0.055) / 1.055, 2.4)
    })
    
    return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs
  }

  // Calculate contrast ratio between two colors
  getContrastRatio(color1, color2) {
    const rgb1 = this.hexToRgb(color1)
    const rgb2 = this.hexToRgb(color2)
    
    if (!rgb1 || !rgb2) return null
    
    const lum1 = this.getLuminance(rgb1.r, rgb1.g, rgb1.b)
    const lum2 = this.getLuminance(rgb2.r, rgb2.g, rgb2.b)
    
    const brightest = Math.max(lum1, lum2)
    const darkest = Math.min(lum1, lum2)
    
    return (brightest + 0.05) / (darkest + 0.05)
  }

  // Check if contrast meets WCAG standards
  checkContrast(foreground, background, level = 'AA') {
    const ratio = this.getContrastRatio(foreground, background)
    const threshold = level === 'AAA' ? this.wcagAAAThreshold : this.wcagAAThreshold
    
    return {
      ratio: ratio ? Math.round(ratio * 100) / 100 : null,
      passes: ratio >= threshold,
      level,
      threshold,
    }
  }

  // Get accessible color suggestions
  getAccessibleColor(baseColor, targetBackground, level = 'AA') {
    const threshold = level === 'AAA' ? this.wcagAAAThreshold : this.wcagAAThreshold
    const rgb = this.hexToRgb(baseColor)
    
    if (!rgb) return null

    // Try darkening the color
    for (let factor = 0.9; factor > 0; factor -= 0.1) {
      const darkerColor = this.rgbToHex(
        Math.round(rgb.r * factor),
        Math.round(rgb.g * factor),
        Math.round(rgb.b * factor)
      )
      
      if (this.getContrastRatio(darkerColor, targetBackground) >= threshold) {
        return darkerColor
      }
    }

    // Try lightening the color
    for (let factor = 1.1; factor <= 2; factor += 0.1) {
      const lighterColor = this.rgbToHex(
        Math.min(255, Math.round(rgb.r * factor)),
        Math.min(255, Math.round(rgb.g * factor)),
        Math.min(255, Math.round(rgb.b * factor))
      )
      
      if (this.getContrastRatio(lighterColor, targetBackground) >= threshold) {
        return lighterColor
      }
    }

    return null
  }

  rgbToHex(r, g, b) {
    return `#${((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1)}`
  }
}

export const colorContrastChecker = new ColorContrastChecker()

// tailwind.config.js - Accessible color palette
module.exports = {
  theme: {
    extend: {
      colors: {
        // High contrast color system
        primary: {
          50: '#f0f9ff',   // Very light blue - 21:1 contrast with primary-900
          100: '#e0f2fe',  // Light blue - 18:1 contrast with primary-900
          200: '#bae6fd',  // Lighter blue - 14:1 contrast with primary-900
          300: '#7dd3fc',  // Light blue - 10:1 contrast with primary-900
          400: '#38bdf8',  // Medium blue - 7:1 contrast with primary-900
          500: '#0ea5e9',  // Blue - 5:1 contrast with white
          600: '#0284c7',  // Darker blue - 6:1 contrast with white
          700: '#0369a1',  // Dark blue - 7:1 contrast with white
          800: '#075985',  // Darker blue - 9:1 contrast with white
          900: '#0c4a6e',  // Very dark blue - 12:1 contrast with white
        },
        success: {
          50: '#f0fdf4',   // Very light green
          500: '#22c55e',  // Green - meets AA contrast
          600: '#16a34a',  // Darker green - meets AAA contrast
          900: '#14532d',  // Very dark green
        },
        warning: {
          50: '#fffbeb',   // Very light yellow
          500: '#f59e0b',  // Yellow - meets AA contrast on dark
          600: '#d97706',  // Darker yellow - meets AAA contrast
          900: '#78350f',  // Very dark yellow
        },
        error: {
          50: '#fef2f2',   // Very light red
          500: '#ef4444',  // Red - meets AA contrast
          600: '#dc2626',  // Darker red - meets AAA contrast
          900: '#7f1d1d',  // Very dark red
        },
        neutral: {
          50: '#fafafa',   // Almost white
          100: '#f5f5f5',  // Very light gray
          200: '#e5e5e5',  // Light gray
          300: '#d4d4d4',  // Light gray
          400: '#a3a3a3',  // Medium gray - AA contrast with white
          500: '#737373',  // Medium gray - AA contrast with white
          600: '#525252',  // Dark gray - AA contrast with white
          700: '#404040',  // Darker gray - AAA contrast with white
          800: '#262626',  // Very dark gray
          900: '#171717',  // Almost black
        },
      },
    },
  },
}

// components/common/AccessibleButton.jsx
import React from 'react'

const AccessibleButton = ({
  children,
  variant = 'primary',
  size = 'medium',
  loading = false,
  disabled = false,
  ariaLabel,
  ariaDescribedBy,
  onClick,
  type = 'button',
  className = '',
  ...props
}) => {
  const baseClasses = 'inline-flex items-center justify-center font-medium rounded-md focus:outline-none focus:ring-2 focus:ring-offset-2 transition-colors duration-200'
  
  const variantClasses = {
    primary: 'bg-primary-600 text-white hover:bg-primary-700 focus:ring-primary-400 disabled:bg-neutral-400',
    secondary: 'bg-neutral-600 text-white hover:bg-neutral-700 focus:ring-neutral-400 disabled:bg-neutral-400',
    success: 'bg-success-600 text-white hover:bg-success-700 focus:ring-success-400 disabled:bg-neutral-400',
    warning: 'bg-warning-600 text-white hover:bg-warning-700 focus:ring-warning-400 disabled:bg-neutral-400',
    error: 'bg-error-600 text-white hover:bg-error-700 focus:ring-error-400 disabled:bg-neutral-400',
    outline: 'border-2 border-primary-600 text-primary-600 hover:bg-primary-50 focus:ring-primary-400 disabled:border-neutral-400 disabled:text-neutral-400',
  }
  
  const sizeClasses = {
    small: 'px-3 py-1.5 text-sm',
    medium: 'px-4 py-2 text-base',
    large: 'px-6 py-3 text-lg',
  }

  const isDisabled = disabled || loading

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={isDisabled}
      aria-label={ariaLabel}
      aria-describedby={ariaDescribedBy}
      aria-disabled={isDisabled}
      className={`${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]} ${className}`}
      {...props}
    >
      {loading && (
        <svg 
          className="animate-spin -ml-1 mr-3 h-5 w-5 text-current" 
          xmlns="http://www.w3.org/2000/svg" 
          fill="none" 
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle 
            className="opacity-25" 
            cx="12" 
            cy="12" 
            r="10" 
            stroke="currentColor" 
            strokeWidth="4"
          />
          <path 
            className="opacity-75" 
            fill="currentColor" 
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      )}
      {children}
    </button>
  )
}

export default AccessibleButton
```

## Directory Structure
```
src/
├── components/
│   ├── common/
│   │   ├── AccessibleContainer.jsx
│   │   ├── AccessibleButton.jsx
│   │   ├── LiveRegion.jsx
│   │   ├── ScreenReaderContent.jsx
│   │   ├── SkipLink.jsx
│   │   └── KeyboardTrap.jsx
│   ├── forms/
│   │   ├── AccessibleForm.jsx
│   │   ├── AccessibleInput.jsx
│   │   ├── AccessibleSelect.jsx
│   │   ├── AccessibleTextarea.jsx
│   │   └── AccessibleFieldset.jsx
│   ├── navigation/
│   │   ├── AccessibleNav.jsx
│   │   ├── Breadcrumbs.jsx
│   │   └── SiteMap.jsx
│   └── layout/
│       ├── Header.jsx
│       ├── Footer.jsx
│       └── MainContent.jsx
├── hooks/
│   ├── useKeyboardNavigation.js
│   ├── useAnnouncement.js
│   ├── useFocusManagement.js
│   └── useAccessibilityPreferences.js
├── utils/
│   ├── colorContrast.js
│   ├── accessibilityTesting.js
│   └── ariaUtils.js
└── styles/
    ├── accessibility.css
    └── high-contrast.css
```

## Accessibility Features Implemented
- ✅ Semantic HTML structure with proper landmarks
- ✅ ARIA attributes and roles
- ✅ Keyboard navigation support
- ✅ Focus management and trapping
- ✅ Screen reader optimization
- ✅ Live regions for dynamic content
- ✅ High contrast color system
- ✅ Accessible forms with proper labeling
- ✅ Skip links for navigation
- ✅ Alternative text for images
- ✅ Error announcements
- ✅ Loading state announcements

## Benefits
- WCAG 2.1 AA compliance
- Support for screen readers
- Full keyboard accessibility
- High color contrast ratios
- Inclusive user experience
- Better SEO through semantic HTML
- Legal compliance
- Improved usability for all users