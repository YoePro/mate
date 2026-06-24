# Accounts And Login

MATE 0.7 has persisted accounts, roles, login sessions, and a browser login page.

## Browser flow

- Open `/login`.
- If no account exists, MATE shows the first-owner bootstrap flow.
- After an owner exists, `/login` signs in with the account name and password.
- The main app at `/` checks `/api/v1/auth/me` during boot. Unauthenticated users are redirected to `/login`.
- The app header shows the current account and includes `Sign out`.

## Roles

Supported roles are:

- `owner`: can manage accounts and write MATE data.
- `admin`: can write MATE data; account management remains owner-only in 0.8.
- `editor`: can write MATE data but cannot manage accounts.
- `viewer`: can read data but cannot create, update, delete, archive, or save graph positions.

## API-only account management

The account management endpoints exist, but there is no full account management UI yet. Owner-only endpoints are:

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
- Editor account receives `403` for account list/create.

## API access rules

- Health, setup status, login, logout, and static assets are public.
- MATE data reads require a valid session.
- MATE data writes require `owner`, `admin`, or `editor`.
- Account management requires `owner`.
- Direct API calls by unauthenticated users return `401`.
- Direct API write calls by `viewer` return `403`.
