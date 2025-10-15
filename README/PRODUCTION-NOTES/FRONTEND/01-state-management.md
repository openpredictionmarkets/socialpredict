# State Management & Architecture Implementation Plan

## Overview
Implement a scalable state management solution and establish proper frontend architecture patterns to handle complex application state, data flow, and component communication.

## Current State Analysis
- Basic React Context for authentication state
- No global state management solution
- Component state scattered across the application
- Direct API calls from components
- No data persistence strategy
- Limited state debugging capabilities

## Implementation Steps

### Step 1: State Management Solution Selection
**Timeline: 1-2 days**

Evaluate and implement a comprehensive state management solution:

```javascript
// Option 1: Redux Toolkit (Recommended)
// store/index.js
import { configureStore } from '@reduxjs/toolkit'
import { persistStore, persistReducer } from 'redux-persist'
import storage from 'redux-persist/lib/storage'
import authSlice from './slices/authSlice'
import marketsSlice from './slices/marketsSlice'
import uiSlice from './slices/uiSlice'
import { api } from './api/apiSlice'

const persistConfig = {
  key: 'root',
  storage,
  whitelist: ['auth'] // Only persist auth data
}

const rootReducer = {
  auth: persistReducer(persistConfig, authSlice),
  markets: marketsSlice,
  ui: uiSlice,
  [api.reducerPath]: api.reducer,
}

export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
      },
    }).concat(api.middleware),
  devTools: process.env.NODE_ENV !== 'production',
})

export const persistor = persistStore(store)
export type RootState = ReturnType<typeof store.getState>
export type AppDispatch = typeof store.dispatch
```

**State management features:**
- Centralized state store
- Predictable state updates
- Time-travel debugging
- State persistence
- Middleware support
- TypeScript integration

### Step 2: Authentication State Management
**Timeline: 2 days**

Replace the current Context-based auth with Redux:

```javascript
// store/slices/authSlice.js
import { createSlice, createAsyncThunk } from '@reduxjs/toolkit'
import { authAPI } from '../../services/authAPI'

// Async thunks
export const loginUser = createAsyncThunk(
  'auth/loginUser',
  async ({ username, password }, { rejectWithValue }) => {
    try {
      const response = await authAPI.login(username, password)
      return response.data
    } catch (error) {
      return rejectWithValue(error.response.data)
    }
  }
)

export const refreshToken = createAsyncThunk(
  'auth/refreshToken',
  async (_, { rejectWithValue }) => {
    try {
      const response = await authAPI.refreshToken()
      return response.data
    } catch (error) {
      return rejectWithValue(error.response.data)
    }
  }
)

export const logoutUser = createAsyncThunk(
  'auth/logoutUser',
  async (_, { dispatch }) => {
    await authAPI.logout()
    dispatch(clearAuthState())
  }
)

const authSlice = createSlice({
  name: 'auth',
  initialState: {
    user: null,
    token: null,
    refreshToken: null,
    isAuthenticated: false,
    isLoading: false,
    error: null,
    mustChangePassword: false,
    permissions: [],
    lastActivity: null,
  },
  reducers: {
    clearAuthState: (state) => {
      state.user = null
      state.token = null
      state.refreshToken = null
      state.isAuthenticated = false
      state.error = null
      state.mustChangePassword = false
      state.permissions = []
    },
    updateLastActivity: (state) => {
      state.lastActivity = Date.now()
    },
    setMustChangePassword: (state, action) => {
      state.mustChangePassword = action.payload
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loginUser.pending, (state) => {
        state.isLoading = true
        state.error = null
      })
      .addCase(loginUser.fulfilled, (state, action) => {
        state.isLoading = false
        state.isAuthenticated = true
        state.user = action.payload.user
        state.token = action.payload.token
        state.refreshToken = action.payload.refreshToken
        state.mustChangePassword = action.payload.mustChangePassword
        state.permissions = action.payload.permissions || []
        state.lastActivity = Date.now()
      })
      .addCase(loginUser.rejected, (state, action) => {
        state.isLoading = false
        state.error = action.payload.message
        state.isAuthenticated = false
      })
  },
})

export const { clearAuthState, updateLastActivity, setMustChangePassword } = authSlice.actions
export default authSlice.reducer

// Selectors
export const selectAuth = (state) => state.auth
export const selectIsAuthenticated = (state) => state.auth.isAuthenticated
export const selectUser = (state) => state.auth.user
export const selectUserPermissions = (state) => state.auth.permissions
export const selectMustChangePassword = (state) => state.auth.mustChangePassword
```

### Step 3: Markets State Management
**Timeline: 2-3 days**

Implement comprehensive markets state management:

```javascript
// store/slices/marketsSlice.js
import { createSlice, createAsyncThunk, createEntityAdapter } from '@reduxjs/toolkit'
import { marketsAPI } from '../../services/marketsAPI'

// Entity adapter for normalized state
const marketsAdapter = createEntityAdapter({
  selectId: (market) => market.id,
  sortComparer: (a, b) => new Date(b.createdAt) - new Date(a.createdAt),
})

// Async thunks
export const fetchMarkets = createAsyncThunk(
  'markets/fetchMarkets',
  async (params = {}) => {
    const response = await marketsAPI.getMarkets(params)
    return response.data
  }
)

export const fetchMarketById = createAsyncThunk(
  'markets/fetchMarketById',
  async (marketId) => {
    const response = await marketsAPI.getMarket(marketId)
    return response.data
  }
)

export const createMarket = createAsyncThunk(
  'markets/createMarket',
  async (marketData, { rejectWithValue }) => {
    try {
      const response = await marketsAPI.createMarket(marketData)
      return response.data
    } catch (error) {
      return rejectWithValue(error.response.data)
    }
  }
)

export const placeBet = createAsyncThunk(
  'markets/placeBet',
  async ({ marketId, betData }, { rejectWithValue }) => {
    try {
      const response = await marketsAPI.placeBet(marketId, betData)
      return { marketId, bet: response.data }
    } catch (error) {
      return rejectWithValue(error.response.data)
    }
  }
)

const marketsSlice = createSlice({
  name: 'markets',
  initialState: marketsAdapter.getInitialState({
    loading: false,
    error: null,
    filters: {
      status: 'all',
      category: 'all',
      search: '',
      sortBy: 'created_at',
      sortOrder: 'desc',
    },
    pagination: {
      page: 1,
      limit: 20,
      total: 0,
      hasMore: true,
    },
    selectedMarket: null,
    userBets: {},
    marketStats: {},
  }),
  reducers: {
    setFilters: (state, action) => {
      state.filters = { ...state.filters, ...action.payload }
    },
    setPagination: (state, action) => {
      state.pagination = { ...state.pagination, ...action.payload }
    },
    setSelectedMarket: (state, action) => {
      state.selectedMarket = action.payload
    },
    clearMarketsError: (state) => {
      state.error = null
    },
    updateMarketInList: (state, action) => {
      marketsAdapter.updateOne(state, action.payload)
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchMarkets.pending, (state) => {
        state.loading = true
        state.error = null
      })
      .addCase(fetchMarkets.fulfilled, (state, action) => {
        state.loading = false
        marketsAdapter.setAll(state, action.payload.markets)
        state.pagination = {
          ...state.pagination,
          total: action.payload.total,
          hasMore: action.payload.hasMore,
        }
      })
      .addCase(fetchMarkets.rejected, (state, action) => {
        state.loading = false
        state.error = action.error.message
      })
      .addCase(createMarket.fulfilled, (state, action) => {
        marketsAdapter.addOne(state, action.payload)
      })
      .addCase(placeBet.fulfilled, (state, action) => {
        const { marketId, bet } = action.payload
        if (!state.userBets[marketId]) {
          state.userBets[marketId] = []
        }
        state.userBets[marketId].push(bet)
      })
  },
})

export const {
  setFilters,
  setPagination,
  setSelectedMarket,
  clearMarketsError,
  updateMarketInList,
} = marketsSlice.actions

// Export the adapter selectors
export const {
  selectAll: selectAllMarkets,
  selectById: selectMarketById,
  selectIds: selectMarketIds,
} = marketsAdapter.getSelectors((state) => state.markets)

// Custom selectors
export const selectMarketsLoading = (state) => state.markets.loading
export const selectMarketsError = (state) => state.markets.error
export const selectMarketsFilters = (state) => state.markets.filters
export const selectMarketsPagination = (state) => state.markets.pagination
export const selectSelectedMarket = (state) => state.markets.selectedMarket
export const selectUserBets = (state) => state.markets.userBets

export default marketsSlice.reducer
```

### Step 4: UI State Management
**Timeline: 1-2 days**

Manage UI-specific state centrally:

```javascript
// store/slices/uiSlice.js
import { createSlice } from '@reduxjs/toolkit'

const uiSlice = createSlice({
  name: 'ui',
  initialState: {
    theme: 'dark',
    sidebarOpen: false,
    notifications: [],
    modals: {
      loginModal: false,
      createMarketModal: false,
      betModal: false,
    },
    loading: {
      global: false,
      markets: false,
      bets: false,
    },
    toasts: [],
    breadcrumbs: [],
  },
  reducers: {
    setTheme: (state, action) => {
      state.theme = action.payload
    },
    toggleSidebar: (state) => {
      state.sidebarOpen = !state.sidebarOpen
    },
    setSidebarOpen: (state, action) => {
      state.sidebarOpen = action.payload
    },
    openModal: (state, action) => {
      state.modals[action.payload] = true
    },
    closeModal: (state, action) => {
      state.modals[action.payload] = false
    },
    closeAllModals: (state) => {
      Object.keys(state.modals).forEach(key => {
        state.modals[key] = false
      })
    },
    setLoading: (state, action) => {
      const { key, loading } = action.payload
      state.loading[key] = loading
    },
    addNotification: (state, action) => {
      state.notifications.push({
        id: Date.now(),
        ...action.payload,
        timestamp: new Date().toISOString(),
      })
    },
    removeNotification: (state, action) => {
      state.notifications = state.notifications.filter(
        notification => notification.id !== action.payload
      )
    },
    addToast: (state, action) => {
      state.toasts.push({
        id: Date.now(),
        ...action.payload,
        timestamp: new Date().toISOString(),
      })
    },
    removeToast: (state, action) => {
      state.toasts = state.toasts.filter(toast => toast.id !== action.payload)
    },
    setBreadcrumbs: (state, action) => {
      state.breadcrumbs = action.payload
    },
  },
})

export const {
  setTheme,
  toggleSidebar,
  setSidebarOpen,
  openModal,
  closeModal,
  closeAllModals,
  setLoading,
  addNotification,
  removeNotification,
  addToast,
  removeToast,
  setBreadcrumbs,
} = uiSlice.actions

// Selectors
export const selectTheme = (state) => state.ui.theme
export const selectSidebarOpen = (state) => state.ui.sidebarOpen
export const selectModals = (state) => state.ui.modals
export const selectLoading = (state) => state.ui.loading
export const selectNotifications = (state) => state.ui.notifications
export const selectToasts = (state) => state.ui.toasts
export const selectBreadcrumbs = (state) => state.ui.breadcrumbs

export default uiSlice.reducer
```

### Step 5: Redux Hooks and Provider Setup
**Timeline: 1 day**

Create typed hooks and update the app provider:

```javascript
// hooks/redux.js
import { useDispatch, useSelector, TypedUseSelectorHook } from 'react-redux'
import type { RootState, AppDispatch } from '../store'

// Typed hooks
export const useAppDispatch = () => useDispatch<AppDispatch>()
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector

// Custom hooks for common operations
export const useAuth = () => {
  const dispatch = useAppDispatch()
  const auth = useAppSelector(selectAuth)
  
  const login = (credentials) => dispatch(loginUser(credentials))
  const logout = () => dispatch(logoutUser())
  const refreshUserToken = () => dispatch(refreshToken())
  
  return {
    ...auth,
    login,
    logout,
    refreshToken: refreshUserToken,
  }
}

export const useMarkets = () => {
  const dispatch = useAppDispatch()
  const markets = useAppSelector(selectAllMarkets)
  const loading = useAppSelector(selectMarketsLoading)
  const error = useAppSelector(selectMarketsError)
  const filters = useAppSelector(selectMarketsFilters)
  
  const fetchMarketsData = (params) => dispatch(fetchMarkets(params))
  const createNewMarket = (marketData) => dispatch(createMarket(marketData))
  const updateFilters = (newFilters) => dispatch(setFilters(newFilters))
  
  return {
    markets,
    loading,
    error,
    filters,
    fetchMarkets: fetchMarketsData,
    createMarket: createNewMarket,
    updateFilters,
  }
}

// App.jsx updated to use Redux
import React from 'react'
import { Provider } from 'react-redux'
import { PersistGate } from 'redux-persist/integration/react'
import { store, persistor } from './store'
import { BrowserRouter as Router } from 'react-router-dom'
import { ErrorBoundary } from 'react-error-boundary'
import Footer from './components/footer/Footer'
import AppRoutes from './helpers/AppRoutes'
import Sidebar from './components/sidebar/Sidebar'
import LoadingSpinner from './components/common/LoadingSpinner'

function App() {
  return (
    <Provider store={store}>
      <PersistGate loading={<LoadingSpinner />} persistor={persistor}>
        <ErrorBoundary FallbackComponent={ErrorFallback}>
          <Router>
            <div className='App bg-primary-background min-h-screen text-white flex flex-col md:flex-row'>
              <Sidebar />
              <div className='flex flex-col flex-grow'>
                <main className='flex-grow p-4 sm:p-6 overflow-y-auto'>
                  <AppRoutes />
                </main>
              </div>
            </div>
          </Router>
        </ErrorBoundary>
      </PersistGate>
    </Provider>
  )
}

export default App
```

### Step 6: Data Normalization and Caching
**Timeline: 2 days**

Implement data normalization and smart caching:

```javascript
// store/api/apiSlice.js
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

const baseQuery = fetchBaseQuery({
  baseUrl: '/api/v0/',
  prepareHeaders: (headers, { getState }) => {
    const token = getState().auth.token
    if (token) {
      headers.set('authorization', `Bearer ${token}`)
    }
    return headers
  },
})

export const api = createApi({
  reducerPath: 'api',
  baseQuery,
  tagTypes: ['Market', 'User', 'Bet', 'Stats'],
  endpoints: (builder) => ({
    getMarkets: builder.query({
      query: (params) => ({
        url: 'markets',
        params,
      }),
      providesTags: ['Market'],
      // Transform response to normalize data
      transformResponse: (response) => ({
        markets: response.data,
        total: response.meta.total,
        hasMore: response.meta.hasMore,
      }),
      // Optimistic updates
      async onQueryStarted(params, { dispatch, queryFulfilled }) {
        try {
          await queryFulfilled
        } catch {
          // Handle error
        }
      },
    }),
    getMarket: builder.query({
      query: (id) => `markets/${id}`,
      providesTags: (result, error, id) => [{ type: 'Market', id }],
    }),
    createMarket: builder.mutation({
      query: (marketData) => ({
        url: 'markets',
        method: 'POST',
        body: marketData,
      }),
      invalidatesTags: ['Market'],
      // Optimistic update
      async onQueryStarted(marketData, { dispatch, queryFulfilled }) {
        const patchResult = dispatch(
          api.util.updateQueryData('getMarkets', {}, (draft) => {
            draft.markets.unshift({
              id: 'temp-' + Date.now(),
              ...marketData,
              status: 'creating',
            })
          })
        )
        try {
          await queryFulfilled
        } catch {
          patchResult.undo()
        }
      },
    }),
  }),
})

export const {
  useGetMarketsQuery,
  useGetMarketQuery,
  useCreateMarketMutation,
} = api
```

## Directory Structure
```
src/
├── store/
│   ├── index.js                # Store configuration
│   ├── slices/
│   │   ├── authSlice.js        # Authentication state
│   │   ├── marketsSlice.js     # Markets state
│   │   ├── uiSlice.js          # UI state
│   │   └── userSlice.js        # User profile state
│   ├── api/
│   │   ├── apiSlice.js         # RTK Query API
│   │   └── middleware.js       # API middleware
│   └── selectors/
│       ├── authSelectors.js    # Memoized auth selectors
│       └── marketSelectors.js  # Memoized market selectors
├── hooks/
│   ├── redux.js               # Typed Redux hooks
│   ├── useAuth.js             # Authentication hooks
│   └── useMarkets.js          # Markets hooks
├── services/
│   ├── authAPI.js             # Authentication API calls
│   ├── marketsAPI.js          # Markets API calls
│   └── baseAPI.js             # Base API configuration
└── utils/
    ├── stateUtils.js          # State utility functions
    └── persistenceUtils.js    # Data persistence utilities
```

## State Structure
```javascript
{
  auth: {
    user: User | null,
    token: string | null,
    isAuthenticated: boolean,
    loading: boolean,
    error: string | null,
    permissions: string[],
  },
  markets: {
    entities: { [id]: Market },
    ids: string[],
    loading: boolean,
    error: string | null,
    filters: FilterState,
    pagination: PaginationState,
  },
  ui: {
    theme: 'light' | 'dark',
    modals: { [modalName]: boolean },
    loading: { [key]: boolean },
    notifications: Notification[],
    toasts: Toast[],
  },
  api: {
    // RTK Query cache
  }
}
```

## Benefits
- Predictable state updates
- Time-travel debugging
- Centralized state management
- Optimistic updates
- Automatic caching and invalidation
- TypeScript support
- DevTools integration
- State persistence
- Normalized data structure
- Performance optimizations

## Migration Strategy
1. Install Redux Toolkit and related dependencies
2. Set up store and provider
3. Migrate authentication context to Redux
4. Implement markets state management
5. Add UI state management
6. Update components to use Redux hooks
7. Implement RTK Query for API calls
8. Add state persistence
9. Performance optimization with selectors
10. Testing and debugging