# Issues/Concerns with the Backend API

In my experimentation with the backend, I've discovered a number of issues that
I feel should be addressed before we can consider the API "fully baked".

## Routes That Fail To Handle Errors Correctly

Once I reconstructed the openapi.yaml file from the source code rather than
the documentation, I found a number of routes that bubble unhandled errors
up to the top-level service.

The first I found was `POST /v0/markets` which failed to handle validation correctly
resulting in `500 Server failed` responses. After digging into the code, I found
that a new market created with a `resolutionDateTime` that was, according to the
business rules, too short in duration, would fail to trap the condition and bubbled
up this 500.

I fixed this code in commit `(af21cd85) - fix 500 resp when market resolution time is too short`.

However, further investigation has found several other handlers that have the same problem:

- Stats
  - backend/handlers/stats/statshandler.go
    - On failures computing or encoding stats, uses:
      - `http.Error(w, "Failed to calculate financial stats: "+err.Error(), http.StatusInternalServerError))`
      - `http.Error(w, "Failed to load setup configuration: "+err.Error(), http.StatusInternalServerError))`
      - `http.Error(w, "Failed to encode stats response: "+err.Error(), http.StatusInternalServerError))`
- CMS Homepage
  - backend/handlers/cms/homepage/http/handler.go
    - Several branches write raw errors:
      - `http.Error(w, err.Error(), http.StatusBadRequest)` in the parse/validation path.
- Bets – Buying/Selling
  - backend/handlers/bets/selling/sellpositionhandler.go
    - Uses `httpErr.Error()` and various `err.Error()` / `dustErr.Error()` for 4xx/5xx:
      - e.g. `http.Error(w, err.Error(), http.StatusBadRequest/Conflict/UnprocessableEntity/InternalServerError)`
  - backend/handlers/bets/buying/buypositionhandler.go
    - Similar pattern:
      - `http.Error(w, httpErr.Error(), httpErr.StatusCode)`
      - `http.Error(w, err.Error(), http.StatusBadRequest/Conflict/UnprocessableEntity/InternalServerError)`
- Positions
  - `backend/handlers/positions/positionshandler.go`
    - On certain service errors, uses:
      - `http.Error(w, "Internal server error", http.StatusInternalServerError)` (generic, not leaking message)
    - But other errors are mapped to literal strings (“Market not found”, etc.), not raw `err.Error()`.
- Users – Profile / Position
  - `backend/handlers/users/changedisplayname.go`
    - Final catch-all path uses:
      - `http.Error(w, err.Error(), http.StatusInternalServerError)`
  - backend/handlers/users/userpositiononmarkethandler.go
    - On some paths:
      - `http.Error(w, "Failed to fetch user position", 500)`
      - `http.Error(w, err.Error(), http.StatusInternalServerError)`
- Users – Profile helpers
  - `backend/handlers/users/profile_helpers.go`
    - writeProfileError inspects `err.Error()` and, for validation errors, passes the message straight through in JSON:
      - `writeProfileJSONError(w, http.StatusBadRequest, message)`
    - For some error types it maps to sanitized messages (“User not found”, etc.), but in the generic case it uses the raw error string.

### Thoughts

Generally I really don't favor relying on HTTP Response Codes to indicate error conditions.

Originally when we formalized the HTTP Protocol, its Response Codes were designed specifically to alert clients of problematic conditions with the server-layer rather than the operational-layer. In other words, these codes were to inform clients of a failure of the server itself, *not the operations it undertook.*

Most developers do not have the luxury of this hindsight and continue to try to force their error conditions into the procrustean bed of HTTP Response Codes.

A perfect example of this is these responses returned by `POST /v0/markets`:

- '400': Bad Request - Validation failed while creating the market.
  - From a server standpoint, this operation is not a bad request, the payload was received and handled properly by the server. Rather, it's an issue with payload validation.
- '403': Forbidden - Password change required before creating markets.
  - This is an abuse of the HTTP standard since the issue is one of business rules, not that the user cannot complete the activity. (I have thoughts on the login process which I'll discuss below.)

So, rather than trying to shoehorn everything into Response Codes, what is a better pattern?

All requests that have been *processed correctly by the backend* return 2XX (200, 201, or 204 as appropriate) along with a payload like:

```json
{
  "ok": true,
  "result": any
}
```

for successful operations, and

```json
{
  "ok": false,
  "reason": string
}
```

for failed operations.

A quick check on the value of `ok` informs the client whether to look for `result` or `reason`.

## The Current Login Process

As you know, the current login flow follows:

- Initial login for new user with initial (temporary) credentials, returns (temporary) JWT
- Client then must change password using JWT, returns string
- Client must then re-login with new (correct) credentials, returns JWT.

This flow is problematic for a number of reasons:

**1. You’re fully authenticating a user you don’t fully trust.**

On first login you return:

```json
{
  "mustChangePassword": true,
  "token": "...",
  "username": "...",
  "usertype": "..."
}
```

Even if the client UI says “you must change your password first,” a malicious client can just ignore mustChangePassword and start using the JWT.

Unless every single protected endpoint:

- parses the JWT,
- looks up the user,
- checks mustChangePassword == false,
- and denies access if not…

…then this user effectively has full access with a default/compromised password.

This is:

- error-prone (easy to forget the check on a new endpoint),
  - The following routes do not return 401:
    - GET /health
    - GET /v0/home
    - GET /v0/setup
    - GET /v0/setup/frontend
    - GET /v0/markets
    - GET /v0/markets/search
    - GET /v0/markets/status
    - GET /v0/markets/status/{status}
    - GET /v0/markets/{id}
    - GET /v0/markets/{id}/leaderboard
    - GET /v0/markets/{id}/projection
    - GET /v0/marketprojection/{marketId}/{amount}/{outcome}
    - GET /v0/markets/bets/{marketId}
    - GET /v0/markets/positions/{marketId}
    - GET /v0/markets/positions/{marketId}/{username}
    - GET /v0/userinfo/{username}
    - GET /v0/usercredit/{username}
    - GET /v0/portfolio/{username}
    - GET /v0/users/{username}/financial
    - GET /v0/stats
    - GET /v0/system/metrics
    - GET /v0/global/leaderboard
    - GET /v0/content/home
  - If it is your intent to make all of these routes freely accessible, I'd suggest you reconsider since many of these leak information that could be used nefariously by person or persons unknown.
- a violation of least privilege, and
- gives attackers a very nice “foot in the door” with default credentials.

Safer pattern:

If mustChangePassword is true, don’t issue a normal access token. Instead:

- Return a distinct error / status like `{ ok: false, reason: "PASSWORD_CHANGE_REQUIRED" }`, or
- Return a limited-scope, short-lived token this is only allowed to call `/changepassword`.

**2. You’re encoding workflow in a flag the server may not actually enforce**

Right now your security story is “the client will do the right thing.”

Security should **always** assume a hostile client.

If the server is relying on `mustChangePassword` being respected by the front-end, but not enforcing it everywhere server-side, you’ve got:

- Inconsistent authorization rules
- Lots of subtle bugs waiting to happen
- Potential privilege escalation if any endpoint forgets the mustChangePassword check

**3. Unnecessary double-login and token churn**

Your flow:

Login → get token A (mustChangePassword: true)

Change password (probably using username+password body, not token)

Login again → get token B (mustChangePassword: false)

Issues:

- Two logins where one would do.
- Now token A is still valid unless you implement revocation / invalidation on password change.
- If token A remains usable, you’ve just given a long-lived token to someone with a weak/default password.

Cleaner flow:

- First login attempt with default password → respond with `{ ok: false, reason: "PASSWORD_CHANGE_REQUIRED" }` + password-reset token (or reuse the same endpoint but no normal JWT yet).
- Call `/changepassword` using that special token.
- If successful, issue a new normal JWT in the same response or require a regular login from then on.

**4. Inconsistent API semantics & UX**

A few design smells (less severe, but still worth fixing):

- You return 200 OK with a JWT even though the user cannot (or should not) use the system yet. That’s semantically weird; 4xx (“extra action required before full login”) is more natural than “OK, here’s a token, but also no.”

- `/changepassword` returns text/plain instead of JSON, which is inconsistent with the rest of the API shape and makes clients do special-casing.

- You mix authentication concerns (issuing tokens) with account-state workflow in an ad hoc way (mustChangePassword + “but here’s a full token”).

**5. Default / temporary credentials become more dangerous**

If new users are created with known/guessable passwords (common in enterprise setups):

Anyone who learns those default creds can log in and immediately get a full JWT.

Even if you intend to require a password change, in practice they may already have enough access to cause harm unless you’ve perfectly locked down all other endpoints to mustChangePassword == false.

If instead you:

Treat “must change password” as a hard gate to issuing a real access token,
you dramatically reduce the damage possible from leaked default credentials.

### Bottom Line

This login flow is brittle and requires special handling on the server **and on the client**.

If, on initial login, you reply with:

```json
{
  "ok": false,
  "reason": "MUST CHANGE PASSWORD"
}
```

The client can then call `/changepassword` with:

```json
{
  "username": string,
  "password": string,
  "newPassword": string
}
```

which

1. authenticates username and password
2. updates password with newPassword
3. return a JWT that allows the user access to the rest of the routes.

## Route Organization

REST is about **Resources** and all of its commands are meant to retrieve and otherwise manipulate resources.

In the backend, I identify the following resources:

- Users
- Content
- Markets
- Bets (Buys/Sells) - N.B. rename this Trades to disconnect semantically from gambling
- Configuration

As it is, there are a number of routes that seem miss-identified related to their resource, e.g. `/v0/markets/positions/{marketId}` is identified as a 'Users' route.

And the Users routes are very oddly organized. `POST /v0/profilechange/description` and its ilk really seem like they could be better designed, i.e. more RESTully.

Here's how I'd redesign the User routes using Create/Retrieve/Update/Delete (CRUD) semantics:

- Create:
  - POST /user - creates a new user

- Retrieve:
  - GET /user[?filter=fields] - paged list of users, optionally filtered
  - GET /user[?name=username] - info for a specific user
    - replaces:
      - GET /userinfo/{username}
  - GET /user/my - info for the currently logged in user
    - replaces:
      - GET /privateprofile
  - GET /user/{id} - info for a specific user

- Update:
  - PUT /user/{id} - replaces a user record
  - PATCH /user/{id} - updates a part of a user record
    - replaces:
      - POST /profilechange/description
      - POST /profilechange/displayname
      - POST /profilechange/emoji
      - POST /profilechange/link

- Delete:
  - DELETE /user/{id} - deletes a user

And the Markets routes: (TO BE COMPLETED... I ran out of steam)

- Create
  - POST /market - creates a new market
- Retrieve
  - GET /market[?filter=fields] - paged list of markets, optionally filtered
  - GET /market[?username=name] - paged list of markets owned by name
  - GET /market[?status=value] - paged list of markets by status
  - GET /market[?search=query] - paged list of markets found using query
    - Note: These query params can be combined, e.g. [?filter=id,questionTitle,description&search=foobar]
  - GET /market/my - list of markets owned by current user
  - GET /market/{id} - get a market by id

- Update
  - PUT /market/{id}
  - PATCH /market/{id} -
    - replaces:
      - POST /markets/{id}/resolve
        - payload: {isResolved:true}
- Delete
  - DELETE /market/{id} - deletes a market (and its associated records in the DB)
