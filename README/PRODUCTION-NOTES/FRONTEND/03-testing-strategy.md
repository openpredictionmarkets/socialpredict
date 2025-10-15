# Testing Strategy Implementation Plan

## Overview
Implement a comprehensive testing strategy covering unit tests, integration tests, end-to-end tests, and performance testing to ensure code quality, reliability, and maintainability.

## Current State Analysis
- Basic testing setup with Jest and React Testing Library
- Minimal test coverage
- No integration or E2E testing
- No performance testing
- No visual regression testing
- No accessibility testing
- No load testing infrastructure

## Implementation Steps

### Step 1: Enhanced Unit Testing Setup
**Timeline: 2-3 days**

Upgrade the testing infrastructure with comprehensive tooling:

```javascript
// jest.config.js - Enhanced Jest configuration
module.exports = {
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/src/tests/setup.js'],
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
    '\\.(jpg|jpeg|png|gif|svg)$': '<rootDir>/src/tests/__mocks__/fileMock.js',
  },
  collectCoverageFrom: [
    'src/**/*.{js,jsx}',
    '!src/index.js',
    '!src/tests/**',
    '!src/**/*.stories.{js,jsx}',
    '!src/**/*.test.{js,jsx}',
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70,
    },
  },
  testMatch: [
    '<rootDir>/src/**/__tests__/**/*.{js,jsx}',
    '<rootDir>/src/**/*.{test,spec}.{js,jsx}',
  ],
  moduleFileExtensions: ['js', 'jsx', 'json'],
  transform: {
    '^.+\\.(js|jsx)$': 'babel-jest',
  },
  testTimeout: 10000,
  verbose: true,
  errorOnDeprecated: true,
}

// src/tests/setup.js - Test setup and global utilities
import '@testing-library/jest-dom'
import { configure } from '@testing-library/react'
import { server } from './mocks/server'

// Configure React Testing Library
configure({ testIdAttribute: 'data-testid' })

// Mock IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  unobserve() {}
}

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  constructor() {}
  disconnect() {}
  observe() {}
  unobserve() {}
}

// Mock Canvas for chart tests
HTMLCanvasElement.prototype.getContext = () => ({
  fillRect: () => {},
  clearRect: () => {},
  getImageData: () => ({ data: new Array(4) }),
  putImageData: () => {},
  createImageData: () => [],
  setTransform: () => {},
  drawImage: () => {},
  save: () => {},
  fillText: () => {},
  restore: () => {},
  beginPath: () => {},
  moveTo: () => {},
  lineTo: () => {},
  closePath: () => {},
  stroke: () => {},
  translate: () => {},
  scale: () => {},
  rotate: () => {},
  arc: () => {},
  fill: () => {},
  measureText: () => ({ width: 0 }),
  transform: () => {},
  rect: () => {},
  clip: () => {},
})

// Start MSW server for API mocking
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())

// Global test utilities
global.testUtils = {
  delay: (ms) => new Promise(resolve => setTimeout(resolve, ms)),
  mockLocalStorage: () => {
    const store = {}
    return {
      getItem: jest.fn(key => store[key] || null),
      setItem: jest.fn((key, value) => { store[key] = value }),
      removeItem: jest.fn(key => { delete store[key] }),
      clear: jest.fn(() => { Object.keys(store).forEach(key => delete store[key]) }),
    }
  },
}
```

Create comprehensive test utilities:

```javascript
// src/tests/utils/testUtils.js
import React from 'react'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router-dom'
import { configureStore } from '@reduxjs/toolkit'
import { authSlice } from '../../store/slices/authSlice'
import { marketsSlice } from '../../store/slices/marketsSlice'
import { uiSlice } from '../../store/slices/uiSlice'

// Test store factory
export const createTestStore = (initialState = {}) => {
  return configureStore({
    reducer: {
      auth: authSlice.reducer,
      markets: marketsSlice.reducer,
      ui: uiSlice.reducer,
    },
    preloadedState: initialState,
    middleware: (getDefaultMiddleware) =>
      getDefaultMiddleware({
        serializableCheck: false,
      }),
  })
}

// Custom render function with providers
export const renderWithProviders = (
  ui,
  {
    initialState = {},
    store = createTestStore(initialState),
    route = '/',
    ...renderOptions
  } = {}
) => {
  const Wrapper = ({ children }) => (
    <Provider store={store}>
      <BrowserRouter>
        <div data-testid="app-wrapper">{children}</div>
      </BrowserRouter>
    </Provider>
  )

  window.history.pushState({}, 'Test page', route)

  return {
    user: userEvent.setup(),
    store,
    ...render(ui, { wrapper: Wrapper, ...renderOptions }),
  }
}

// Test data factories
export const createMockUser = (overrides = {}) => ({
  id: '1',
  username: 'testuser',
  email: 'test@example.com',
  isAdmin: false,
  balance: 1000,
  createdAt: '2023-01-01T00:00:00Z',
  ...overrides,
})

export const createMockMarket = (overrides = {}) => ({
  id: '1',
  title: 'Test Market',
  description: 'A test market for testing',
  category: 'sports',
  status: 'open',
  closingDate: '2024-12-31T23:59:59Z',
  totalVolume: 5000,
  participantCount: 25,
  outcomes: [
    { id: '1', name: 'Yes', odds: 1.5, probability: 0.6 },
    { id: '2', name: 'No', odds: 2.5, probability: 0.4 },
  ],
  bets: [],
  createdAt: '2023-01-01T00:00:00Z',
  ...overrides,
})

export const createMockBet = (overrides = {}) => ({
  id: '1',
  marketId: '1',
  userId: '1',
  outcomeId: '1',
  amount: 100,
  odds: 1.5,
  status: 'active',
  placedAt: '2023-01-01T00:00:00Z',
  ...overrides,
})

// Custom matchers
export const customMatchers = {
  toBeInTheDOM: (received) => {
    const pass = document.body.contains(received)
    return {
      pass,
      message: () => `Expected element ${pass ? 'not ' : ''}to be in the DOM`,
    }
  },
  toHaveLoadingState: (received) => {
    const hasSpinner = received.querySelector('[data-testid="loading-spinner"]')
    const hasLoadingClass = received.classList.contains('loading')
    const pass = hasSpinner || hasLoadingClass
    
    return {
      pass,
      message: () => `Expected element ${pass ? 'not ' : ''}to have loading state`,
    }
  },
}

// Async testing utilities
export const waitForLoadingToFinish = () =>
  waitFor(() => {
    expect(screen.queryByTestId('loading-spinner')).not.toBeInTheDocument()
  })

export const waitForElementToBeRemoved = (element) =>
  waitFor(() => expect(element).not.toBeInTheDocument())

// Event simulation utilities
export const simulateNetworkDelay = (ms = 1000) =>
  new Promise(resolve => setTimeout(resolve, ms))

export const simulateNetworkError = () => {
  const error = new Error('Network error')
  error.name = 'NetworkError'
  throw error
}

// Mock API responses
export const createMockApiResponse = (data, status = 200) => ({
  ok: status >= 200 && status < 300,
  status,
  json: () => Promise.resolve(data),
  text: () => Promise.resolve(JSON.stringify(data)),
})
```

### Step 2: Component Unit Tests
**Timeline: 3-4 days**

Implement comprehensive component testing:

```javascript
// src/components/markets/MarketCard.test.jsx
import React from 'react'
import { screen, fireEvent, waitFor } from '@testing-library/react'
import { renderWithProviders, createMockMarket } from '../../tests/utils/testUtils'
import MarketCard from './MarketCard'

describe('MarketCard', () => {
  const mockMarket = createMockMarket()
  const mockOnBetClick = jest.fn()

  beforeEach(() => {
    jest.clearAllMocks()
  })

  it('renders market information correctly', () => {
    renderWithProviders(
      <MarketCard 
        marketId={mockMarket.id} 
        onBetClick={mockOnBetClick} 
      />,
      {
        initialState: {
          markets: {
            entities: { [mockMarket.id]: mockMarket },
            ids: [mockMarket.id],
          },
        },
      }
    )

    expect(screen.getByText(mockMarket.title)).toBeInTheDocument()
    expect(screen.getByText(mockMarket.description)).toBeInTheDocument()
    expect(screen.getByText(`Volume: $${mockMarket.totalVolume}`)).toBeInTheDocument()
    expect(screen.getByText(`Participants: ${mockMarket.participantCount}`)).toBeInTheDocument()
  })

  it('calculates and displays winning probability', () => {
    renderWithProviders(
      <MarketCard marketId={mockMarket.id} onBetClick={mockOnBetClick} />,
      {
        initialState: {
          markets: {
            entities: { [mockMarket.id]: mockMarket },
            ids: [mockMarket.id],
          },
        },
      }
    )

    expect(screen.getByText('Place Bet (60.0%)')).toBeInTheDocument()
  })

  it('calls onBetClick when bet button is clicked', async () => {
    const { user } = renderWithProviders(
      <MarketCard marketId={mockMarket.id} onBetClick={mockOnBetClick} />,
      {
        initialState: {
          markets: {
            entities: { [mockMarket.id]: mockMarket },
            ids: [mockMarket.id],
          },
        },
      }
    )

    const betButton = screen.getByText('Place Bet (60.0%)')
    await user.click(betButton)

    expect(mockOnBetClick).toHaveBeenCalledWith(mockMarket.id)
    expect(mockOnBetClick).toHaveBeenCalledTimes(1)
  })

  it('displays market status correctly', () => {
    const closedMarket = createMockMarket({ status: 'closed' })
    
    renderWithProviders(
      <MarketCard marketId={closedMarket.id} onBetClick={mockOnBetClick} />,
      {
        initialState: {
          markets: {
            entities: { [closedMarket.id]: closedMarket },
            ids: [closedMarket.id],
          },
        },
      }
    )

    expect(screen.getByText('Market Closed')).toBeInTheDocument()
    expect(screen.queryByText('Place Bet')).not.toBeInTheDocument()
  })

  it('handles missing market gracefully', () => {
    renderWithProviders(
      <MarketCard marketId="nonexistent" onBetClick={mockOnBetClick} />,
      { initialState: { markets: { entities: {}, ids: [] } } }
    )

    expect(screen.queryByText(mockMarket.title)).not.toBeInTheDocument()
  })

  it('formats time remaining correctly', () => {
    const futureDate = new Date(Date.now() + 24 * 60 * 60 * 1000) // 24 hours from now
    const marketWithTimeRemaining = createMockMarket({ 
      closingDate: futureDate.toISOString() 
    })

    renderWithProviders(
      <MarketCard marketId={marketWithTimeRemaining.id} onBetClick={mockOnBetClick} />,
      {
        initialState: {
          markets: {
            entities: { [marketWithTimeRemaining.id]: marketWithTimeRemaining },
            ids: [marketWithTimeRemaining.id],
          },
        },
      }
    )

    expect(screen.getByText(/Time Left: 1d/)).toBeInTheDocument()
  })

  it('updates when market data changes', () => {
    const { store } = renderWithProviders(
      <MarketCard marketId={mockMarket.id} onBetClick={mockOnBetClick} />,
      {
        initialState: {
          markets: {
            entities: { [mockMarket.id]: mockMarket },
            ids: [mockMarket.id],
          },
        },
      }
    )

    // Update market data
    const updatedMarket = { ...mockMarket, totalVolume: 10000 }
    store.dispatch({
      type: 'markets/updateMarketInList',
      payload: { id: mockMarket.id, changes: { totalVolume: 10000 } }
    })

    expect(screen.getByText('Volume: $10000')).toBeInTheDocument()
  })
})

// src/hooks/useAuth.test.js
import { renderHook, act } from '@testing-library/react'
import { Provider } from 'react-redux'
import { useAuth } from '../useAuth'
import { createTestStore, createMockUser } from '../../tests/utils/testUtils'

describe('useAuth', () => {
  let store

  beforeEach(() => {
    store = createTestStore()
  })

  const wrapper = ({ children }) => (
    <Provider store={store}>{children}</Provider>
  )

  it('returns initial auth state', () => {
    const { result } = renderHook(() => useAuth(), { wrapper })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBe(null)
    expect(result.current.loading).toBe(false)
  })

  it('handles login successfully', async () => {
    const mockUser = createMockUser()
    
    // Mock successful API response
    global.fetch = jest.fn().mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({
        user: mockUser,
        token: 'mock-token',
        refreshToken: 'mock-refresh-token',
      }),
    })

    const { result } = renderHook(() => useAuth(), { wrapper })

    await act(async () => {
      await result.current.login({ username: 'test', password: 'password' })
    })

    expect(result.current.isAuthenticated).toBe(true)
    expect(result.current.user).toEqual(mockUser)
    expect(result.current.token).toBe('mock-token')
  })

  it('handles login failure', async () => {
    global.fetch = jest.fn().mockRejectedValueOnce({
      response: { data: { message: 'Invalid credentials' } }
    })

    const { result } = renderHook(() => useAuth(), { wrapper })

    await act(async () => {
      try {
        await result.current.login({ username: 'test', password: 'wrong' })
      } catch (error) {
        // Expected to fail
      }
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.error).toBeTruthy()
  })

  it('handles logout', async () => {
    // Start with authenticated state
    store = createTestStore({
      auth: {
        isAuthenticated: true,
        user: createMockUser(),
        token: 'mock-token',
      },
    })

    const { result } = renderHook(() => useAuth(), { wrapper })

    await act(async () => {
      await result.current.logout()
    })

    expect(result.current.isAuthenticated).toBe(false)
    expect(result.current.user).toBe(null)
    expect(result.current.token).toBe(null)
  })
})
```

### Step 3: Integration Testing
**Timeline: 3-4 days**

Implement integration tests for complete user flows:

```javascript
// src/tests/integration/marketFlow.test.js
import React from 'react'
import { screen, waitFor } from '@testing-library/react'
import { server } from '../mocks/server'
import { rest } from 'msw'
import { renderWithProviders, createMockMarket, createMockUser } from '../utils/testUtils'
import App from '../../App'

describe('Market Flow Integration', () => {
  const mockUser = createMockUser({ balance: 1000 })
  const mockMarkets = [
    createMockMarket({ id: '1', title: 'Test Market 1' }),
    createMockMarket({ id: '2', title: 'Test Market 2' }),
  ]

  beforeEach(() => {
    // Set up API mocks
    server.use(
      rest.get('/api/v0/markets', (req, res, ctx) => {
        return res(ctx.json({ data: mockMarkets, meta: { total: 2 } }))
      }),
      rest.get('/api/v0/markets/:id', (req, res, ctx) => {
        const market = mockMarkets.find(m => m.id === req.params.id)
        return res(ctx.json({ data: market }))
      }),
      rest.post('/api/v0/auth/login', (req, res, ctx) => {
        return res(ctx.json({
          user: mockUser,
          token: 'mock-token',
          refreshToken: 'mock-refresh-token',
        }))
      }),
      rest.post('/api/v0/markets/:id/bets', (req, res, ctx) => {
        return res(ctx.json({
          id: '1',
          marketId: req.params.id,
          userId: mockUser.id,
          amount: 100,
          status: 'active',
        }))
      })
    )
  })

  it('completes full betting flow', async () => {
    const { user } = renderWithProviders(<App />, {
      route: '/markets',
    })

    // Wait for markets to load
    await waitFor(() => {
      expect(screen.getByText('Test Market 1')).toBeInTheDocument()
    })

    // Click on first market
    await user.click(screen.getByText('Test Market 1'))

    // Wait for market detail page
    await waitFor(() => {
      expect(screen.getByText('Place Bet')).toBeInTheDocument()
    })

    // Click place bet button
    await user.click(screen.getByText('Place Bet'))

    // Login modal should appear
    await waitFor(() => {
      expect(screen.getByText('Login')).toBeInTheDocument()
    })

    // Fill login form
    await user.type(screen.getByLabelText('Username'), 'testuser')
    await user.type(screen.getByLabelText('Password'), 'password')
    await user.click(screen.getByRole('button', { name: 'Login' }))

    // Wait for login to complete and bet modal to appear
    await waitFor(() => {
      expect(screen.getByText('Place Your Bet')).toBeInTheDocument()
    })

    // Fill bet form
    await user.type(screen.getByLabelText('Bet Amount'), '100')
    await user.click(screen.getByText('Yes'))
    await user.click(screen.getByRole('button', { name: 'Confirm Bet' }))

    // Wait for success message
    await waitFor(() => {
      expect(screen.getByText('Bet placed successfully!')).toBeInTheDocument()
    })

    // Check user balance updated
    expect(screen.getByText('Balance: $900')).toBeInTheDocument()
  })

  it('handles market creation flow', async () => {
    server.use(
      rest.post('/api/v0/markets', (req, res, ctx) => {
        return res(ctx.json({
          id: '3',
          title: 'New Test Market',
          status: 'open',
          createdAt: new Date().toISOString(),
        }))
      })
    )

    const { user } = renderWithProviders(<App />, {
      route: '/markets',
      initialState: {
        auth: {
          isAuthenticated: true,
          user: { ...mockUser, isAdmin: true },
          token: 'mock-token',
        },
      },
    })

    // Click create market button
    await user.click(screen.getByText('Create Market'))

    // Fill market form
    await user.type(screen.getByLabelText('Market Title'), 'New Test Market')
    await user.type(screen.getByLabelText('Description'), 'A new test market')
    await user.selectOptions(screen.getByLabelText('Category'), 'sports')
    await user.type(screen.getByLabelText('Closing Date'), '2024-12-31')

    // Submit form
    await user.click(screen.getByRole('button', { name: 'Create Market' }))

    // Wait for success and redirect
    await waitFor(() => {
      expect(screen.getByText('Market created successfully!')).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getByText('New Test Market')).toBeInTheDocument()
    })
  })

  it('handles error states gracefully', async () => {
    // Mock API error
    server.use(
      rest.get('/api/v0/markets', (req, res, ctx) => {
        return res(ctx.status(500), ctx.json({ message: 'Server error' }))
      })
    )

    renderWithProviders(<App />, { route: '/markets' })

    // Wait for error message
    await waitFor(() => {
      expect(screen.getByText('Failed to load markets')).toBeInTheDocument()
    })

    // Check retry button exists
    expect(screen.getByText('Retry')).toBeInTheDocument()
  })
})

// src/tests/integration/authFlow.test.js
import React from 'react'
import { screen, waitFor } from '@testing-library/react'
import { server } from '../mocks/server'
import { rest } from 'msw'
import { renderWithProviders, createMockUser } from '../utils/testUtils'
import App from '../../App'

describe('Authentication Flow Integration', () => {
  const mockUser = createMockUser()

  it('handles complete authentication cycle', async () => {
    server.use(
      rest.post('/api/v0/auth/login', (req, res, ctx) => {
        return res(ctx.json({
          user: mockUser,
          token: 'mock-token',
          refreshToken: 'mock-refresh-token',
        }))
      }),
      rest.post('/api/v0/auth/refresh', (req, res, ctx) => {
        return res(ctx.json({
          token: 'new-mock-token',
          refreshToken: 'new-mock-refresh-token',
        }))
      }),
      rest.post('/api/v0/auth/logout', (req, res, ctx) => {
        return res(ctx.status(200))
      })
    )

    const { user } = renderWithProviders(<App />)

    // Click login button
    await user.click(screen.getByText('Login'))

    // Fill login form
    await user.type(screen.getByLabelText('Username'), mockUser.username)
    await user.type(screen.getByLabelText('Password'), 'password')
    await user.click(screen.getByRole('button', { name: 'Login' }))

    // Wait for successful login
    await waitFor(() => {
      expect(screen.getByText(`Welcome, ${mockUser.username}`)).toBeInTheDocument()
    })

    // Check user menu appears
    expect(screen.getByText('Profile')).toBeInTheDocument()
    expect(screen.getByText('Logout')).toBeInTheDocument()

    // Test logout
    await user.click(screen.getByText('Logout'))

    await waitFor(() => {
      expect(screen.getByText('Login')).toBeInTheDocument()
      expect(screen.queryByText(`Welcome, ${mockUser.username}`)).not.toBeInTheDocument()
    })
  })

  it('handles token refresh automatically', async () => {
    let tokenRefreshed = false

    server.use(
      rest.get('/api/v0/markets', (req, res, ctx) => {
        const authHeader = req.headers.get('authorization')
        
        if (!tokenRefreshed && authHeader === 'Bearer expired-token') {
          tokenRefreshed = true
          return res(ctx.status(401), ctx.json({ message: 'Token expired' }))
        }
        
        return res(ctx.json({ data: [] }))
      }),
      rest.post('/api/v0/auth/refresh', (req, res, ctx) => {
        return res(ctx.json({
          token: 'new-mock-token',
          refreshToken: 'new-mock-refresh-token',
        }))
      })
    )

    renderWithProviders(<App />, {
      route: '/markets',
      initialState: {
        auth: {
          isAuthenticated: true,
          user: mockUser,
          token: 'expired-token',
          refreshToken: 'mock-refresh-token',
        },
      },
    })

    // Wait for automatic token refresh
    await waitFor(() => {
      expect(tokenRefreshed).toBe(true)
    })

    // Verify user stays logged in
    expect(screen.queryByText('Login')).not.toBeInTheDocument()
  })
})
```

### Step 4: End-to-End Testing with Playwright
**Timeline: 3-4 days**

Set up comprehensive E2E testing:

```javascript
// playwright.config.js
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './src/tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [
    ['html'],
    ['json', { outputFile: 'test-results/results.json' }],
    ['junit', { outputFile: 'test-results/results.xml' }],
  ],
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    port: 3000,
    reuseExistingServer: !process.env.CI,
  },
})

// src/tests/e2e/marketBetting.spec.js
import { test, expect } from '@playwright/test'

test.describe('Market Betting Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Set up test data
    await page.route('/api/v0/markets', async (route) => {
      await route.fulfill({
        json: {
          data: [
            {
              id: '1',
              title: 'Test Market',
              description: 'E2E Test Market',
              status: 'open',
              totalVolume: 1000,
              participantCount: 10,
            },
          ],
        },
      })
    })

    await page.route('/api/v0/auth/login', async (route) => {
      await route.fulfill({
        json: {
          user: { id: '1', username: 'testuser', balance: 500 },
          token: 'test-token',
        },
      })
    })
  })

  test('should complete betting flow successfully', async ({ page }) => {
    await page.goto('/markets')

    // Wait for markets to load
    await expect(page.locator('text=Test Market')).toBeVisible()

    // Click on market
    await page.click('text=Test Market')

    // Wait for market detail page
    await expect(page.locator('text=Place Bet')).toBeVisible()

    // Click place bet
    await page.click('text=Place Bet')

    // Should show login modal
    await expect(page.locator('text=Login')).toBeVisible()

    // Fill login form
    await page.fill('[data-testid="username-input"]', 'testuser')
    await page.fill('[data-testid="password-input"]', 'password')
    await page.click('[data-testid="login-button"]')

    // Wait for bet modal
    await expect(page.locator('text=Place Your Bet')).toBeVisible()

    // Fill bet amount
    await page.fill('[data-testid="bet-amount-input"]', '100')

    // Select outcome
    await page.click('[data-testid="outcome-yes"]')

    // Confirm bet
    await page.click('[data-testid="confirm-bet-button"]')

    // Wait for success message
    await expect(page.locator('text=Bet placed successfully')).toBeVisible()
  })

  test('should handle mobile betting flow', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/markets')

    // Test mobile navigation
    await page.click('[data-testid="mobile-menu-button"]')
    await expect(page.locator('[data-testid="mobile-sidebar"]')).toBeVisible()

    // Continue with betting flow...
    await page.click('text=Test Market')
    await expect(page.locator('text=Place Bet')).toBeVisible()
  })

  test('should handle network errors gracefully', async ({ page }) => {
    // Simulate network failure
    await page.route('/api/v0/markets', (route) => route.abort())

    await page.goto('/markets')

    // Should show error message
    await expect(page.locator('text=Failed to load markets')).toBeVisible()
    await expect(page.locator('text=Retry')).toBeVisible()

    // Test retry functionality
    await page.route('/api/v0/markets', async (route) => {
      await route.fulfill({
        json: { data: [{ id: '1', title: 'Test Market' }] },
      })
    })

    await page.click('text=Retry')
    await expect(page.locator('text=Test Market')).toBeVisible()
  })
})

// src/tests/e2e/accessibility.spec.js
import { test, expect } from '@playwright/test'
import AxeBuilder from '@axe-core/playwright'

test.describe('Accessibility Tests', () => {
  test('should not have accessibility violations on home page', async ({ page }) => {
    await page.goto('/')

    const accessibilityScanResults = await new AxeBuilder({ page }).analyze()
    expect(accessibilityScanResults.violations).toEqual([])
  })

  test('should not have accessibility violations on markets page', async ({ page }) => {
    await page.goto('/markets')

    const accessibilityScanResults = await new AxeBuilder({ page }).analyze()
    expect(accessibilityScanResults.violations).toEqual([])
  })

  test('should support keyboard navigation', async ({ page }) => {
    await page.goto('/markets')

    // Test tab navigation
    await page.keyboard.press('Tab')
    const focusedElement = await page.locator(':focus')
    await expect(focusedElement).toBeVisible()

    // Test skip link
    await page.keyboard.press('Tab')
    const skipLink = page.locator('[data-testid="skip-to-content"]')
    if (await skipLink.isVisible()) {
      await expect(skipLink).toBeFocused()
    }
  })
})
```

### Step 5: Performance Testing
**Timeline: 2-3 days**

Implement performance testing suite:

```javascript
// src/tests/performance/lighthouse.test.js
import lighthouse from 'lighthouse'
import chromeLauncher from 'chrome-launcher'

describe('Lighthouse Performance Tests', () => {
  let chrome

  beforeAll(async () => {
    chrome = await chromeLauncher.launch({ chromeFlags: ['--headless'] })
  })

  afterAll(async () => {
    await chrome.kill()
  })

  const runLighthouse = async (url) => {
    const options = {
      logLevel: 'info',
      output: 'json',
      onlyCategories: ['performance', 'accessibility', 'best-practices', 'seo'],
      port: chrome.port,
    }

    const runnerResult = await lighthouse(url, options)
    return runnerResult.lhr
  }

  test('home page performance', async () => {
    const result = await runLighthouse('http://localhost:3000/')

    expect(result.categories.performance.score).toBeGreaterThan(0.8)
    expect(result.categories.accessibility.score).toBeGreaterThan(0.9)
    expect(result.categories['best-practices'].score).toBeGreaterThan(0.9)
    expect(result.categories.seo.score).toBeGreaterThan(0.8)

    // Check specific metrics
    const metrics = result.audits
    expect(metrics['first-contentful-paint'].numericValue).toBeLessThan(2000)
    expect(metrics['largest-contentful-paint'].numericValue).toBeLessThan(2500)
    expect(metrics['cumulative-layout-shift'].numericValue).toBeLessThan(0.1)
  }, 30000)

  test('markets page performance', async () => {
    const result = await runLighthouse('http://localhost:3000/markets')

    expect(result.categories.performance.score).toBeGreaterThan(0.75)
    expect(result.audits['first-contentful-paint'].numericValue).toBeLessThan(2500)
    expect(result.audits['time-to-interactive'].numericValue).toBeLessThan(3500)
  }, 30000)
})

// src/tests/performance/loadTesting.test.js
import { test } from '@playwright/test'

test.describe('Load Testing', () => {
  test('should handle concurrent users', async ({ browser }) => {
    const concurrentUsers = 10
    const promises = []

    for (let i = 0; i < concurrentUsers; i++) {
      promises.push(
        (async () => {
          const context = await browser.newContext()
          const page = await context.newPage()

          const startTime = Date.now()
          await page.goto('/markets')
          await page.waitForSelector('[data-testid="market-list"]')
          const loadTime = Date.now() - startTime

          await context.close()
          return loadTime
        })()
      )
    }

    const loadTimes = await Promise.all(promises)
    const averageLoadTime = loadTimes.reduce((a, b) => a + b) / loadTimes.length

    expect(averageLoadTime).toBeLessThan(5000) // 5 seconds
  })

  test('should handle memory usage efficiently', async ({ page }) => {
    await page.goto('/markets')

    // Simulate heavy usage
    for (let i = 0; i < 50; i++) {
      await page.click('[data-testid="refresh-markets"]')
      await page.waitForTimeout(100)
    }

    // Check for memory leaks
    const memoryUsage = await page.evaluate(() => {
      return performance.memory ? performance.memory.usedJSHeapSize : 0
    })

    // Memory usage should be reasonable (less than 50MB)
    expect(memoryUsage).toBeLessThan(50 * 1024 * 1024)
  })
})
```

## Directory Structure
```
src/
├── tests/
│   ├── setup.js                    # Jest setup
│   ├── utils/
│   │   ├── testUtils.js            # Testing utilities
│   │   ├── mockData.js             # Mock data factories
│   │   └── customMatchers.js       # Custom Jest matchers
│   ├── mocks/
│   │   ├── server.js               # MSW server setup
│   │   ├── handlers.js             # API mock handlers
│   │   └── fileMock.js             # File mocks
│   ├── unit/
│   │   ├── components/             # Component tests
│   │   ├── hooks/                  # Hook tests
│   │   ├── utils/                  # Utility tests
│   │   └── services/               # Service tests
│   ├── integration/
│   │   ├── marketFlow.test.js      # Market integration tests
│   │   ├── authFlow.test.js        # Auth integration tests
│   │   └── userJourney.test.js     # User journey tests
│   ├── e2e/
│   │   ├── marketBetting.spec.js   # E2E betting tests
│   │   ├── navigation.spec.js      # Navigation tests
│   │   ├── accessibility.spec.js   # Accessibility tests
│   │   └── mobile.spec.js          # Mobile tests
│   └── performance/
│       ├── lighthouse.test.js      # Lighthouse tests
│       ├── loadTesting.test.js     # Load testing
│       └── memoryLeaks.test.js     # Memory leak tests
├── __mocks__/                      # Global mocks
└── coverage/                       # Coverage reports
```

## Testing Scripts
```json
{
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage",
    "test:integration": "jest --testPathPattern=integration",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui",
    "test:performance": "jest --testPathPattern=performance",
    "test:accessibility": "playwright test --grep='accessibility'",
    "test:all": "npm run test:coverage && npm run test:e2e && npm run test:performance"
  }
}
```

## Benefits
- High code coverage and reliability
- Comprehensive testing strategy
- Automated regression detection
- Performance monitoring
- Accessibility compliance
- Cross-browser compatibility
- Mobile testing coverage
- Load testing capabilities
- Visual regression detection
- Continuous quality assurance

## Quality Gates
- Minimum 70% code coverage
- All E2E tests passing
- Performance budgets met
- Accessibility compliance (WCAG 2.1 AA)
- Cross-browser compatibility
- Mobile responsiveness verified
- Load testing benchmarks met