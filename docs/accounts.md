# Accounts And Login

MATE has persisted accounts, roles, login sessions, and a browser login page.

## Browser flow

- Open `/login`.
- If no account exists, MATE shows the first-owner bootstrap flow.
- After an owner exists, `/login` also allows self-service account creation.
- After an owner exists, `/login` signs in with the account name and password.
- The main app at `/` checks `/api/v1/auth/me` during boot. Unauthenticated users are redirected to `/login`.
- The app header shows the current account in an account menu.
- The account menu includes sign out, browser-local preferences, placeholders for profile/system settings, and account administration for roles that can manage accounts.

## Roles

Supported roles are:

- `owner`: can manage all accounts and write MATE data.
- `admin`: can write MATE data and manage non-admin accounts.
- `editor`: can write MATE data but cannot manage accounts.
- `viewer`: can read data but cannot create, update, delete, archive, or save graph positions.

## Browser account management

MATE 0.13 adds an Account Administration window in the account menu.

- Owners can list accounts, create accounts, assign roles, change roles, and disable accounts.
- Admins can list accounts and manage `editor` and `viewer` accounts.
- Admins cannot create, disable, or change `owner` or `admin` accounts.
- Existing passwords are never displayed. Passwords are only accepted when creating a new account.
- Disabling yourself is blocked.
- Disabling or demoting the last usable owner is blocked by the backend.

## Self-service account creation

The `/login` page has a `Create account` tab after the first owner has been created.

- Self-service accounts are created with the `editor` role.
- Self-service account creation never creates `owner` or `admin` accounts.
- The new account is signed in immediately after creation.
- Owner bootstrap remains separate and is only available while no owner exists.

## Account management API

Account management endpoints require an authenticated `owner` or `admin` session. Admin sessions are subject to the restrictions above.

- `GET /api/v1/accounts`
- `POST /api/v1/accounts`
- `GET /api/v1/accounts/{id}`
- `PATCH /api/v1/accounts/{id}`
- `POST /api/v1/accounts/{id}/disable`

## Sessions

Successful login sets a HTTP-only `mate_session` cookie. Logout deletes the persisted session and expires the cookie.

## Tested 0.7 behavior

- First owner bootstrap.
- Login and logout.
- Current-user lookup.
- Owner creates an additional account and changes its role.
- Admin can manage non-admin accounts.
- Editor account receives `403` for account list/create.

## API access rules

- Health, setup status, login, logout, and static assets are public.
- MATE data reads require a valid session.
- MATE data writes require `owner`, `admin`, or `editor`.
- Account management requires `owner` or `admin`; admin actions are limited to non-admin accounts.
- Direct API calls by unauthenticated users return `401`.
- Direct API write calls by `viewer` return `403`.

## Network ownership

MATE 0.9 adds user-owned networks.

- Each network has one owner account.
- Users only list and open networks they own.
- Network writes require both a write-capable role and network ownership.
- Non-owners cannot edit a network by direct API call.
- Full sharing and delegated edit permissions are intentionally not implemented yet.
