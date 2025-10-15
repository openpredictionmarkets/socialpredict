# SocialPredict API Documentation

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Base URL](#base-url)
4. [Response Format](#response-format)
5. [Error Handling](#error-handling)
6. [Endpoints](#endpoints)
   - [Public Endpoints](#public-endpoints)
   - [Authentication Endpoints](#authentication-endpoints)
   - [Configuration Endpoints](#configuration-endpoints)
   - [Statistics & Metrics](#statistics--metrics)
   - [Markets](#markets)
   - [Users](#users)
   - [User Profile Management](#user-profile-management)
   - [Betting & Trading](#betting--trading)
   - [Market Management](#market-management)
   - [Administration](#administration)
7. [Data Models](#data-models)

## Overview

SocialPredict is a prediction market platform where users can create markets, place bets on outcomes, and track their performance. The API provides endpoints for user management, market operations, betting, and administrative functions.

**Version**: 1.0.0  
**License**: MIT  
**Contact**: [SocialPredict Team](https://github.com/raisch/socialpredict)

## Authentication

Most endpoints require authentication using JWT Bearer tokens. To authenticate:

1. Obtain a JWT token by calling the `/v0/login` endpoint
2. Include the token in the `Authorization` header of subsequent requests:
   ```
   Authorization: Bearer <your-jwt-token>
   ```

## Base URL

```
http://localhost:8080
```

## Response Format

All API responses are in JSON format. Successful responses include the requested data, while error responses follow a standard error format.

## Error Handling

Error responses follow this format:

```json
{
  "error": "error_type",
  "message": "Human-readable error message"
}
```

Common HTTP status codes:
- **200**: Success
- **201**: Created
- **400**: Bad Request
- **401**: Unauthorized
- **403**: Forbidden
- **404**: Not Found
- **500**: Internal Server Error

---

## Endpoints

### Public Endpoints

These endpoints do not require authentication.

#### GET /v0/home

Get home page data and verify API connectivity.

**Response**:
```json
{
  "message": "Data From the Backend!"
}
```

---

### Authentication Endpoints

#### POST /v0/login

Authenticate user and receive JWT token.

**Request Body**:
```json
{
  "username": "string",    // Required, 3-30 characters
  "password": "string"     // Required, minimum 1 character
}
```

**Response** (200):
```json
{
  "token": "jwt-token-string",
  "username": "user123",
  "usertype": "standard",
  "mustChangePassword": false
}
```

**Error Response** (401):
```json
{
  "error": "unauthorized",
  "message": "Invalid credentials"
}
```

---

### Configuration Endpoints

#### GET /v0/setup

Get application setup and economics configuration.

**Response** (200):
```json
{
  "marketcreation": {
    "initialMarketProbability": 0.5,
    "initialMarketSubsidization": 1000,
    "initialMarketYes": 500,
    "initialMarketNo": 500,
    "minimumFutureHours": 24.0
  },
  "marketincentives": {
    "createMarketCost": 100,
    "traderBonus": 10
  },
  "user": {
    "initialAccountBalance": 10000,
    "maximumDebtAllowed": 1000
  },
  "betting": {
    "minimumBet": 1,
    "maxDustPerSale": 5,
    "betFees": {
      "initialBetFee": 1,
      "buySharesFee": 2,
      "sellSharesFee": 2
    }
  }
}
```

---

### Statistics & Metrics

#### GET /v0/stats

Get general application statistics.

**Response** (200):
```json
{
  // Statistics object (structure varies)
}
```

#### GET /v0/system/metrics

Get system performance and health metrics.

**Response** (200):
```json
{
  // System metrics object (structure varies)
}
```

#### GET /v0/global/leaderboard

Get the global user leaderboard.

**Response** (200):
```json
{
  // Global leaderboard object (structure varies)
}
```

---

### Markets

#### GET /v0/markets

List all markets (random selection, up to 100).

**Response** (200):
```json
{
  "markets": [
    {
      "market": {
        "id": 1,
        "questionTitle": "Will it rain tomorrow?",
        "description": "Weather prediction for tomorrow",
        "outcomeType": "binary",
        "resolutionDateTime": "2025-10-08T12:00:00Z",
        "finalResolutionDateTime": "2025-10-08T18:00:00Z",
        "utcOffset": 0,
        "isResolved": false,
        "resolutionResult": null,
        "initialProbability": 0.5,
        "creatorUsername": "weatherman"
      },
      "creator": {
        "username": "weatherman",
        "displayname": "Weather Expert",
        "usertype": "standard",
        // ... other user fields
      },
      "lastProbability": 0.65,
      "numUsers": 25,
      "totalVolume": 5000
    }
  ]
}
```

#### GET /v0/markets/search

Search for markets based on query parameters.

**Query Parameters**:
- `q` (string): Search query

**Response**: Same format as `/v0/markets`

#### GET /v0/markets/active

List all active (unresolved) markets.

**Response**: Same format as `/v0/markets`

#### GET /v0/markets/closed

List all closed markets.

**Response**: Same format as `/v0/markets`

#### GET /v0/markets/resolved

List all resolved markets.

**Response**: Same format as `/v0/markets`

#### GET /v0/markets/{marketId}

Get detailed information about a specific market.

**Path Parameters**:
- `marketId` (integer): Market ID

**Response** (200):
```json
{
  "id": 1,
  "questionTitle": "Will it rain tomorrow?",
  "description": "Weather prediction for tomorrow",
  "outcomeType": "binary",
  "resolutionDateTime": "2025-10-08T12:00:00Z",
  "finalResolutionDateTime": "2025-10-08T18:00:00Z",
  "utcOffset": 0,
  "isResolved": false,
  "resolutionResult": null,
  "initialProbability": 0.5,
  "creatorUsername": "weatherman"
}
```

#### GET /v0/marketprojection/{marketId}/{amount}/{outcome}/

Calculate the new probability if a bet of specified amount and outcome were placed.

**Path Parameters**:
- `marketId` (integer): Market ID
- `amount` (integer): Bet amount
- `outcome` (string): Bet outcome

**Response** (200):
```json
{
  "newProbability": 0.68
}
```

#### GET /v0/markets/bets/{marketId}

Get all bets for a specific market.

**Path Parameters**:
- `marketId` (integer): Market ID

**Response** (200):
```json
[
  {
    "id": 1,
    "action": "buy",
    "username": "trader1",
    "marketId": 1,
    "amount": 100,
    "placedAt": "2025-10-07T10:30:00Z",
    "outcome": "yes"
  }
]
```

#### GET /v0/markets/positions/{marketId}

Get all positions for a specific market.

**Path Parameters**:
- `marketId` (integer): Market ID

**Response** (200):
```json
{
  // Market positions object (structure varies)
}
```

#### GET /v0/markets/positions/{marketId}/{username}

Get a specific user's positions in a market.

**Path Parameters**:
- `marketId` (integer): Market ID
- `username` (string): Username

**Response** (200):
```json
{
  // User positions in market (structure varies)
}
```

#### GET /v0/markets/leaderboard/{marketId}

Get the leaderboard for a specific market.

**Path Parameters**:
- `marketId` (integer): Market ID

**Response** (200):
```json
{
  // Market leaderboard object (structure varies)
}
```

---

### Users

#### GET /v0/userinfo/{username}

Get public information about a user.

**Path Parameters**:
- `username` (string): Username

**Response** (200):
```json
{
  "username": "trader1",
  "displayname": "Trader One",
  "usertype": "standard",
  "initialAccountBalance": 10000,
  "accountBalance": 9500,
  "personalEmoji": "📈",
  "description": "Professional trader",
  "personalink1": "https://twitter.com/trader1",
  "personalink2": "",
  "personalink3": "",
  "personalink4": ""
}
```

#### GET /v0/usercredit/{username}

Get user credit/balance information.

**Path Parameters**:
- `username` (string): Username

**Response** (200):
```json
{
  "accountBalance": 9500
}
```

#### GET /v0/portfolio/{username}

Get a user's investment portfolio.

**Path Parameters**:
- `username` (string): Username

**Response** (200):
```json
{
  // User portfolio object (structure varies)
}
```

#### GET /v0/users/{username}/financial

Get financial information for a user.

**Path Parameters**:
- `username` (string): Username

**Response** (200):
```json
{
  // User financial information (structure varies)
}
```

---

### User Profile Management

These endpoints require authentication and operate on the authenticated user's profile.

#### GET /v0/privateprofile

Get private profile information for the authenticated user.

**Response** (200):
```json
{
  "username": "trader1",
  "displayname": "Trader One",
  "usertype": "standard",
  "initialAccountBalance": 10000,
  "accountBalance": 9500,
  "personalEmoji": "📈",
  "description": "Professional trader",
  "personalink1": "https://twitter.com/trader1",
  "personalink2": "",
  "personalink3": "",
  "personalink4": "",
  "email": "trader1@example.com",
  "apiKey": "api-key-string"
}
```

#### POST /v0/changepassword

Change the authenticated user's password.

**Request Body**:
```json
{
  "currentPassword": "oldpassword",  // Optional
  "newPassword": "newpassword"       // Required
}
```

**Response** (200): Success (no body)

#### POST /v0/profilechange/displayname

Change the authenticated user's display name.

**Request Body**:
```json
{
  "displayName": "New Display Name"
}
```

**Response** (200): Success (no body)

#### POST /v0/profilechange/emoji

Change the authenticated user's personal emoji.

**Request Body**:
```json
{
  "emoji": "🚀"
}
```

**Response** (200): Success (no body)

#### POST /v0/profilechange/description

Change the authenticated user's profile description.

**Request Body**:
```json
{
  "description": "New profile description"
}
```

**Response** (200): Success (no body)

#### POST /v0/profilechange/links

Change the authenticated user's personal links.

**Request Body**:
```json
{
  "personalLink1": "https://twitter.com/user",
  "personalLink2": "https://linkedin.com/in/user",
  "personalLink3": "https://user.blog",
  "personalLink4": "https://user.website"
}
```

**Response** (200): Success (no body)

---

### Betting & Trading

#### POST /v0/bet

Place a bet on a market outcome.

**Request Body**:
```json
{
  "marketId": 1,        // Required
  "amount": 100,        // Required (int64)
  "outcome": "yes"      // Required
}
```

**Response** (201):
```json
{
  "id": 123,
  "action": "buy",
  "username": "trader1",
  "marketId": 1,
  "amount": 100,
  "placedAt": "2025-10-07T14:30:00Z",
  "outcome": "yes"
}
```

#### GET /v0/userposition/{marketId}

Get the authenticated user's position in a specific market.

**Path Parameters**:
- `marketId` (integer): Market ID

**Response** (200):
```json
{
  // User position in market (structure varies)
}
```

#### POST /v0/sell

Sell shares in a market position.

**Request Body**:
```json
{
  "marketId": 1,
  "amount": 50,
  "outcome": "yes"
}
```

**Response** (200): Success (no body)

---

### Market Management

#### POST /v0/create

Create a new prediction market.

**Request Body**:
```json
{
  "questionTitle": "Will it snow next week?",      // Required
  "description": "Weather prediction",            // Required
  "outcomeType": "binary",                        // Required
  "resolutionDateTime": "2025-10-15T12:00:00Z",  // Required
  "utcOffset": 0,                                 // Optional
  "initialProbability": 0.3                       // Optional
}
```

**Response** (201):
```json
{
  "id": 2,
  "questionTitle": "Will it snow next week?",
  "description": "Weather prediction",
  "outcomeType": "binary",
  "resolutionDateTime": "2025-10-15T12:00:00Z",
  "finalResolutionDateTime": "2025-10-15T18:00:00Z",
  "utcOffset": 0,
  "isResolved": false,
  "resolutionResult": null,
  "initialProbability": 0.3,
  "creatorUsername": "weatherman"
}
```

#### POST /v0/resolve/{marketId}

Resolve a market with the final outcome.

**Path Parameters**:
- `marketId` (integer): Market ID

**Request Body**:
```json
{
  "resolutionResult": "yes"  // Required
}
```

**Response** (200): Success (no body)

---

### Administration

These endpoints require admin privileges.

#### POST /v0/admin/createuser

Create a new user account (admin only).

**Request Body**:
```json
{
  "username": "newuser",           // Required
  "displayName": "New User",       // Required
  "email": "newuser@example.com",  // Required
  "password": "password123",       // Required
  "userType": "standard"           // Required
}
```

**Response** (201):
```json
{
  "id": 456,
  "username": "newuser",
  "displayname": "New User",
  "usertype": "standard",
  "initialAccountBalance": 10000,
  "accountBalance": 10000,
  "personalEmoji": "",
  "description": "",
  "personalink1": "",
  "personalink2": "",
  "personalink3": "",
  "personalink4": "",
  "mustChangePassword": true
}
```

---

## Data Models

### User

Complete user model with all fields:

```json
{
  "id": 1,
  "username": "trader1",
  "displayname": "Trader One",
  "usertype": "standard",
  "initialAccountBalance": 10000,
  "accountBalance": 9500,
  "personalEmoji": "📈",
  "description": "Professional trader",
  "personalink1": "https://twitter.com/trader1",
  "personalink2": "",
  "personalink3": "",
  "personalink4": "",
  "mustChangePassword": false
}
```

### PublicUser

Public user information (subset of User):

```json
{
  "username": "trader1",
  "displayname": "Trader One",
  "usertype": "standard",
  "initialAccountBalance": 10000,
  "accountBalance": 9500,
  "personalEmoji": "📈",
  "description": "Professional trader",
  "personalink1": "https://twitter.com/trader1",
  "personalink2": "",
  "personalink3": "",
  "personalink4": ""
}
```

### PrivateUser

Private user information (sensitive data):

```json
{
  "email": "trader1@example.com",
  "apiKey": "api-key-string"
}
```

### Market

Market information:

```json
{
  "id": 1,
  "questionTitle": "Will it rain tomorrow?",
  "description": "Weather prediction for tomorrow",
  "outcomeType": "binary",
  "resolutionDateTime": "2025-10-08T12:00:00Z",
  "finalResolutionDateTime": "2025-10-08T18:00:00Z",
  "utcOffset": 0,
  "isResolved": false,
  "resolutionResult": null,
  "initialProbability": 0.5,
  "creatorUsername": "weatherman"
}
```

### Bet

Bet/trade information:

```json
{
  "id": 1,
  "action": "buy",
  "username": "trader1",
  "marketId": 1,
  "amount": 100,
  "placedAt": "2025-10-07T10:30:00Z",
  "outcome": "yes"
}
```

### MarketOverview

Market overview with additional statistics:

```json
{
  "market": {
    // PublicResponseMarket object
  },
  "creator": {
    // PublicUser object
  },
  "lastProbability": 0.65,
  "numUsers": 25,
  "totalVolume": 5000
}
```

### EconomicsConfig

Application economics configuration:

```json
{
  "marketcreation": {
    "initialMarketProbability": 0.5,
    "initialMarketSubsidization": 1000,
    "initialMarketYes": 500,
    "initialMarketNo": 500,
    "minimumFutureHours": 24.0
  },
  "marketincentives": {
    "createMarketCost": 100,
    "traderBonus": 10
  },
  "user": {
    "initialAccountBalance": 10000,
    "maximumDebtAllowed": 1000
  },
  "betting": {
    "minimumBet": 1,
    "maxDustPerSale": 5,
    "betFees": {
      "initialBetFee": 1,
      "buySharesFee": 2,
      "sellSharesFee": 2
    }
  }
}
```

---

## Notes

- All timestamps are in ISO 8601 format (RFC 3339)
- All monetary amounts are represented as integers (smallest currency unit)
- User types include: "standard", "admin", etc.
- Outcome types include: "binary", etc.
- Market resolution results depend on outcome type (e.g., "yes"/"no" for binary)
- JWT tokens expire after 24 hours
- Rate limiting may apply to certain endpoints
- Some response structures may vary based on the specific implementation details