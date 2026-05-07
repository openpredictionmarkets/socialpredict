## One-Time Password Setup Token Plan

### Goal

Replace admin-visible generated passwords with a one-time password-setup token flow.

Desired flow:

1. Admin creates user with just `username`.
2. Backend creates the user with `mustChangePassword=true`.
3. Backend generates a one-time setup token, stores only its hash plus expiry.
4. API returns a setup link or setup token, not a password.
5. New user opens the link and sets their own password.
6. Backend clears the token and flips `mustChangePassword=false`.

This removes plaintext passwords from:

* API responses
* frontend state
* clipboard copy
* admin UI rendering

### Target Files

Backend:

* `backend/internal/domain/users/admin_creation.go`
* `backend/handlers/admin/adduser.go`
* `backend/handlers/admin/adduser_http_test.go`
* `backend/internal/domain/users/admin_creation_test.go`
* `backend/models/user.go`
* `backend/internal/repository/users/repository.go`
* `backend/server/server.go`
* `backend/handlers/users/dto/profile.go`
* new handler file for `POST /v0/password-setup`
* migration file(s) if schema changes are required

Frontend:

* `frontend/src/components/layouts/admin/AddUser.jsx`
* `frontend/src/helpers/AppRoutes.jsx`
* new password-setup page/layout/component
* any shared API helpers needed for the new endpoint

### Recommended Contract

Prefer returning a one-time setup URL from the admin create-user endpoint, with the raw token embedded only in that one response. Store only a hash of the token in the database.

Suggested response shape:

```json
{
  "message": "User created successfully",
  "username": "freshuser",
  "usertype": "REGULAR",
  "passwordSetupUrl": "http://localhost/setup-password?token=...",
  "passwordSetupExpiresAt": "2026-05-08T12:00:00Z"
}
```

Suggested password setup request:

```json
{
  "token": "raw-one-time-token",
  "newPassword": "ChosenPassword123"
}
```

### Step-by-Step Plan

#### 1. Change the admin-created user domain contract

Update `AdminManagedUserCreateResult` in `backend/internal/domain/users/admin_creation.go`:

* Remove `Password`.
* Add fields for token delivery metadata, for example:
  * `SetupToken`
  * `SetupExpiresAt`
* Keep `Username` and `UserType`.

Reason:

* The domain service should no longer produce a reusable password for the admin.

#### 2. Add persistence for the one-time setup token

Extend the user persistence model so the backend can validate and expire setup tokens:

* Add fields to the user record such as:
  * `PasswordSetupTokenHash`
  * `PasswordSetupExpiresAt`
* Decide whether you also want:
  * `PasswordSetupIssuedAt`
  * `PasswordSetupUsedAt`

Work items:

* Update `backend/models/user.go`
* Update repository mapping in `backend/internal/repository/users/repository.go`
* Add a migration for the new columns

Constraints:

* Store only a hash of the token, never the raw token.
* Add an expiry timestamp at creation time.

#### 3. Refactor admin-managed user creation

Update `CreateAdminManagedUser` in `backend/internal/domain/users/admin_creation.go`:

* Remove generated password creation with `gofakeit.Password(...)`.
* Generate a cryptographically strong random setup token instead.
* Hash the setup token before storing it.
* Set `MustChangePassword=true`.
* Persist the user with:
  * no usable plaintext password returned to the caller
  * token hash + expiry stored on the user

Decision point:

* Either create the user with a random unusable password hash that the user never sees, or introduce a dedicated “password not set yet” flow in the credentials layer.
* The simpler path is to store a random secret-derived hash and require the setup token flow before any real login can succeed.

#### 4. Change the admin create-user handler and API response

Update `backend/handlers/admin/adduser.go`:

* Stop returning `"password": result.Password`.
* Return only:
  * `message`
  * `username`
  * `usertype`
  * setup-link or setup-token metadata

Also update:

* `backend/handlers/admin/adduser_http_test.go`
* any OpenAPI or API docs covering `/v0/admin/createuser`

Success criteria:

* No response body from `/v0/admin/createuser` contains a password.

#### 5. Add the password setup endpoint

Add a new public endpoint such as `POST /v0/password-setup`.

Handler responsibilities:

* Decode `{ token, newPassword }`
* Validate password policy
* Hash the raw token from the request
* Look up the user by token hash
* Reject invalid, expired, or already-used tokens
* Set the new password hash
* Clear token hash and expiry
* Clear `mustChangePassword`

Files to add/update:

* new handler under `backend/handlers/users/`
* request DTO in `backend/handlers/users/dto/profile.go` or a new DTO file
* service/domain method in `backend/internal/domain/users/service.go` or a focused credential/setup slice
* repository methods needed to find and clear tokens
* `backend/server/server.go` route registration

Suggested behavior:

* Token is single-use.
* Expired token returns `400` or `410`, not `500`.
* Successful setup returns a simple success envelope, not a login token.

#### 6. Keep change-password and setup-password flows separate

Do not overload the existing `POST /v0/changepassword` flow.

Reason:

* `changepassword` is for authenticated users with a current password.
* `password-setup` is for first-time onboarding with a one-time token.

Implementation guidance:

* Reuse password validation logic where possible.
* Keep the HTTP contracts distinct.

#### 7. Update the admin UI

Refactor `frontend/src/components/layouts/admin/AddUser.jsx`:

* Remove `password` state.
* Remove password rendering.
* Remove clipboard content that includes the password.
* Replace it with setup-link or setup-token display and copy behavior.

New UX:

* After create-user succeeds, show:
  * username
  * password setup link
  * expiry message

Success criteria:

* No frontend state stores the created user password.
* No admin screen renders or copies a created user password.

#### 8. Add a password setup page in the frontend

Create a new page and route, for example:

* `/setup-password`

Frontend behavior:

* Read the `token` from the query string or path
* Render:
  * new password
  * confirm password
* Submit to `POST /v0/password-setup`
* On success:
  * redirect to login
  * show a success message

Files to update:

* `frontend/src/helpers/AppRoutes.jsx`
* new page/layout/component files under `frontend/src/pages/` and `frontend/src/components/layouts/`

#### 9. Review the auth gate behavior

Confirm the updated flow still behaves correctly:

* Newly created users cannot use protected routes before setting a password.
* `mustChangePassword` remains true until password setup succeeds.
* After password setup, normal login works and protected routes are available.

Specific check:

* Make sure the frontend does not rely on a returned password anywhere in the onboarding flow.

#### 10. Add tests for the new security properties

Backend tests:

* Admin create-user no longer returns `password`
* Created user record stores token hash, not raw token
* Password setup succeeds with valid token
* Invalid token fails
* Expired token fails
* Reused token fails
* Successful setup clears token fields and `mustChangePassword`

Frontend tests:

* Admin add-user UI renders setup link, not password
* Password setup page validates confirm-password mismatch
* Successful setup redirects correctly

End-to-end flow to verify:

1. Admin creates user with username only.
2. Response includes setup link/token and expiry.
3. User cannot access protected resources before setup.
4. User sets password with token.
5. User logs in with the chosen password.
6. `mustChangePassword` is now false.

### Implementation Order

Recommended order for the actual coding work:

1. Update domain result types and persistence model.
2. Add migration and repository support for token hash + expiry.
3. Refactor admin creation to generate and store setup tokens.
4. Change `/v0/admin/createuser` response contract.
5. Add `POST /v0/password-setup`.
6. Update admin UI to show setup link instead of password.
7. Add frontend setup-password page and route.
8. Update tests and docs.

### Definition of Done

This checkpoint is complete when all of the following are true:

* No backend response returns a created user password.
* No frontend component stores, renders, or copies a created user password.
* The backend stores only a hashed setup token plus expiry.
* Setup tokens are single-use and expire.
* A new user can complete onboarding without any admin-visible password.
* Existing change-password behavior for authenticated users still works.
