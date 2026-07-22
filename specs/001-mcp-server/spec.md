# Feature Specification: MCP Server

**Feature Branch**: `feature/mcp-server`

**Created**: 2026-07-22

**Status**: Draft

**Input**: User description: "I want to add MCP server functionality to SocialPredict. It should be able to work when deployed on a vps i.e. it should work in dev, localhost AND production"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - AI Assistant Reads Market Data (Priority: P1)

A person using an AI assistant (any MCP-compatible client) connects it to a SocialPredict
instance and asks questions about the markets: "What markets are open?", "What's the current
probability on the election market?", "How much volume has this market seen?". The assistant
answers using live data from the instance without the person opening the website.

**Why this priority**: This is the core value of MCP integration — making the platform's
public market data available to AI assistants. It is useful entirely on its own and defines
the feature's whole v1 surface.

**Independent Test**: Can be fully tested by connecting a standard MCP client to a running
instance, listing markets, and verifying the returned data matches what the website shows.

**Acceptance Scenarios**:

1. **Given** a running SocialPredict instance with open markets, **When** an MCP client
   connects and requests the list of markets, **Then** it receives the markets with their
   titles, current probabilities, status, and closing times.
2. **Given** a specific market, **When** the MCP client requests its details, **Then** it
   receives the same probability, volume, and status information shown on the market's web page.
3. **Given** an MCP client connects, **When** it asks what capabilities are available,
   **Then** it receives a discoverable list of the read-only data queries the server offers.
4. **Given** an MCP client, **When** it attempts any action that would change platform state
   or access private account data, **Then** the request is refused — no such capability is
   offered in v1.

---

### User Story 2 - Operator Enables MCP on Any Deployment (Priority: P2)

An instance operator deploys SocialPredict the standard way — locally for development, on
localhost via the documented setup, or on a public VPS with HTTPS — and the MCP capability
works in all three without environment-specific rework. AI assistants can reach the
instance's MCP endpoint at a predictable address derived from the instance's domain.

**Why this priority**: The user's explicit constraint — dev, localhost, AND production must
all work. Without this, the feature is a demo, not a capability of the product.

**Independent Test**: Can be tested by running the documented deployment for each of the
three environments and connecting the same MCP client configuration (adjusted only for
address) to each.

**Acceptance Scenarios**:

1. **Given** a developer running the local development setup, **When** they connect an MCP
   client to the local address, **Then** all MCP capabilities function.
2. **Given** a production VPS deployment with HTTPS on a public domain, **When** a remote
   MCP client connects over the public internet, **Then** all MCP capabilities function over
   the encrypted connection.
3. **Given** the standard deployment documentation, **When** an operator follows it, **Then**
   enabling MCP requires no undocumented or environment-specific steps.

---

### Edge Cases

- What happens when a client requests a capability that does not exist (including any write
  or account action)? (Clear refusal; the advertised catalog contains only read queries.)
- How does the system handle an assistant issuing rapid repeated requests (e.g., a polling
  loop)? (Standard abuse protections apply; platform performance for other users is
  unaffected.)
- What happens when a market closes or resolves between two reads? (Subsequent reads reflect
  the new state; no stale-state guarantees beyond what the website provides.)
- What happens when the client disconnects mid-request? (No effect on platform state — all
  capabilities are read-only.)
- How does a client reach the endpoint in production behind a reverse proxy / HTTPS
  terminator? (Same public entry point as the rest of the application.)
- What happens when the instance has no markets yet? (An empty list is returned, not an
  error.)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST expose an MCP-compatible interface that AI assistant clients can
  connect to and discover the available capabilities.
- **FR-002**: System MUST provide public market data through MCP: market list and per-market
  details (probability, volume, status, closing/resolution info), consistent with what the
  website displays.
- **FR-003**: The v1 MCP surface MUST be strictly read-only: no capability may change
  platform state (bets, balances, markets, users) or expose data not visible to a logged-out
  website visitor.
- **FR-004**: Requests for anything outside the advertised read-only catalog MUST be refused
  with a clear error and no information leakage.
- **FR-005**: The MCP capability MUST function in all three documented environments — local
  development, localhost deployment, and production VPS with HTTPS — using the standard
  deployment path and configuration surface.
- **FR-006**: System MUST apply the same abuse protections (e.g., rate limiting) to MCP
  traffic as to equivalent public web traffic, so automated clients cannot degrade the
  platform.

### Key Entities

- **MCP Capability**: A named read-only data query the server advertises to clients
  (e.g., "list markets", "get market details"). The catalog of these defines the entire
  v1 surface.
- **MCP Session/Connection**: A client's anonymous connection over which capabilities are
  invoked. No account linkage exists in v1.
- **Market**: Existing platform entity; MCP exposes its public attributes (title,
  probability, volume, status, timing) read-only. No new entities are introduced.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A user with a standard MCP-compatible AI client can go from "never connected"
  to receiving live market data in under 5 minutes using only the feature's documentation.
- **SC-002**: The same client configuration (changing only the server address) works against
  dev, localhost, and production deployments — verified on all three with zero
  environment-specific fixes.
- **SC-003**: 100% of requests outside the read-only catalog are refused; zero platform
  state changes originate from MCP traffic.
- **SC-004**: Market data returned through MCP matches the website's displayed values for
  the same moment in time.
- **SC-005**: Sustained automated MCP request load does not degrade website responsiveness
  for regular users beyond normal public-traffic behavior.

## Assumptions

- v1 is deliberately read-only. Account linkage (personal credentials), trading actions, and
  market creation via MCP are explicitly deferred to a future iteration.
- "Public market data" means exactly what a logged-out website visitor can see; MCP adds no
  new visibility.
- The standard Docker-based deployment remains the delivery vehicle; the MCP endpoint is
  reachable through the instance's existing public entry point in production.
- No web frontend changes are required for v1.
- v1 targets a single SocialPredict instance per MCP connection; cross-instance aggregation
  is out of scope.
