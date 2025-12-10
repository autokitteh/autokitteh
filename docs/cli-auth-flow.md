# CLI Authentication Flow

This document explains how the `ak auth login` command authenticates users.

## Overview

The CLI login flow uses OAuth-based authentication with a local callback server. The flow involves:

1. CLI starts a local HTTP server
2. Opens browser to the AutoKitteh server
3. User authenticates via OAuth provider (Google, GitHub, or Descope)
4. Server redirects back to CLI with a token
5. CLI stores the token locally

## Step-by-Step Flow

### 1. User Runs Login Command

```bash
ak auth login
```

**Code**: `cmd/ak/cmd/auth/login.go:20-41`

- Starts local HTTP server on random port (e.g., `54321`)
- Builds URL: `https://api.autokitteh.cloud/auth/cli-login?p=54321`
- Opens browser to this URL
- Waits for callback with token

### 2. Browser Opens to Server

**URL**: `/auth/cli-login?p=54321`

**Code**: `internal/backend/auth/authloginhttpsvc/svc.go:104-117`

- Extracts port number from query parameter
- Sets cookie `redir=/auth/finish-cli-login?p=54321` (for later redirect)
- Redirects to `/login`

### 3. Login Page Displayed

**URL**: `/login`

**Code**: `internal/backend/auth/authloginhttpsvc/svc.go:88-102`

Shows available authentication providers:

- Google OAuth
- GitHub OAuth
- Descope

If only one provider is configured, automatically redirects to it.

### 4. User Authenticates

**Example with Descope**: `/auth/descope/login`

**Code**: `internal/backend/auth/authloginhttpsvc/descope.go:36-77`

**First visit** (no JWT parameter):

- Shows Descope login form
- User enters credentials
- Descope authenticates user

**Second visit** (with JWT parameter):

- URL: `/auth/descope/login?jwt=<TOKEN>`
- Server validates JWT with Descope API
- Extracts user email and name from JWT claims
- Calls success handler

### 5. Success Handler Processes User

**Code**: `internal/backend/auth/authloginhttpsvc/svc.go:213-302`

1. **Look up user** by email in database
2. **If user exists**:
   - If status is `invited`: activate user
   - If status is not `active`: reject login
3. **If user doesn't exist**:
   - If `RejectNewUsers=true`: reject login
   - Otherwise: create new user
4. **Create session cookies**:
   - `ak_user_session` = JWT token (HttpOnly, 14 days)
   - `ak_logged_in` = "true" (14 days)
5. **Redirect** to stored destination from `redir` cookie

### 6. Redirect to Finish CLI Login

**URL**: `/auth/finish-cli-login?p=54321`

**Code**: `internal/backend/auth/authloginhttpsvc/svc.go:119-139`

- Gets authenticated user from session cookie
- Creates **CLI access token** (JWT)
- Redirects to: `http://localhost:54321/?token=<CLI_TOKEN>`

### 7. CLI Receives Token

**Code**: `cmd/ak/cmd/auth/login.go:52-55`

- Local HTTP server receives callback
- Extracts token from query parameter
- Displays: "You can close this tab now"
- Returns token to waiting goroutine

### 8. Token Stored Locally

**Code**: `cmd/ak/common/creds.go:55-79`

- Reads existing credentials file (if any)
- Stores token by hostname
- Writes to: `~/.config/autokitteh/credentials`
- File format (YAML):
  ```yaml
  api.autokitteh.cloud: eyJhbGciOiJIUzI1NiIs...
  localhost:9980: eyJhbGciOiJIUzI1NiIs...
  ```
- File permissions: `0600` (read/write for user only)

## Visual Flow Diagram

```
┌─────────────┐
│ User runs:  │
│ ak auth     │
│ login       │
└──────┬──────┘
       │
       v
┌──────────────────────────────────┐
│ CLI starts local HTTP server     │
│ Port: 54321 (random)             │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Opens browser to:                │
│ /auth/cli-login?p=54321          │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Server sets cookie:              │
│ redir=/auth/finish-cli-login     │
│        ?p=54321                  │
│                                  │
│ Redirects to: /login             │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Shows login page:                │
│ - Google OAuth                   │
│ - GitHub OAuth                   │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ User selects provider            │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Provider authenticates user      │
│ Returns with JWT                 │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Server validates JWT             │
│ Gets/creates user in DB          │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Sets session cookies:            │
│ - ak_user_session (JWT)          │
│ - ak_logged_in (true)            │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Reads redir cookie               │
│ Redirects to:                    │
│ /auth/finish-cli-login?p=54321   │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Gets user from session           │
│ Creates CLI access token         │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Redirects to:                    │
│ http://localhost:54321/          │
│   ?token=<CLI_TOKEN>             │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ CLI receives token               │
│ Shows: "You can close this tab"  │
└──────┬───────────────────────────┘
       │
       v
┌──────────────────────────────────┐
│ Stores token in:                 │
│ ~/.config/autokitteh/credentials │
└──────┬───────────────────────────┘
       │
       v
   ┌───────┐
   │ Done! │
   └───────┘
```

## Key Files

| File                                                | Purpose                                   |
| --------------------------------------------------- | ----------------------------------------- |
| `cmd/ak/cmd/auth/login.go`                          | CLI login command & local callback server |
| `cmd/ak/common/creds.go`                            | Token storage on filesystem               |
| `cmd/ak/common/config.go`                           | Server URL configuration                  |
| `internal/backend/auth/authloginhttpsvc/svc.go`     | Main HTTP routes & success handler        |
| `internal/backend/auth/authloginhttpsvc/redir.go`   | Redirect cookie management                |
| `internal/backend/auth/authloginhttpsvc/descope.go` | Descope provider integration              |
| `internal/backend/auth/authloginhttpsvc/google.go`  | Google OAuth integration                  |
| `internal/backend/auth/authloginhttpsvc/github.go`  | GitHub OAuth integration                  |
| `internal/backend/auth/authsessions/store.go`       | Session cookie creation & validation      |

## Two Types of Tokens

### 1. Session Token (Browser)

- **Created**: After OAuth login (Line 53 in `authsessions/store.go`)
- **Storage**: Browser cookie `ak_user_session`
- **Purpose**: Web dashboard authentication
- **Lifetime**: 14 days
- **Format**: JWT (HttpOnly, Secure)

### 2. CLI Access Token

- **Created**: At `/auth/finish-cli-login` (Line 132 in `authloginhttpsvc/svc.go`)
- **Storage**: `~/.config/autokitteh/credentials` file
- **Purpose**: CLI command authentication
- **Lifetime**: Configurable (default: long-lived)
- **Usage**: Sent as `Authorization: Bearer <token>` header

## Configuration

### Server URL

- **Default**: `https://api.autokitteh.cloud`
- **Override**: Set `http.service_url` in `~/.config/autokitteh/config.yaml`
- **Development**: Often `http://localhost:9980`

### OAuth Providers

Configure in server config:

- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`
- `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_REDIRECT_URL`
- `DESCOPE_PROJECT_ID`

### User Creation

- `RejectNewUsers`: If `true`, only invited users can log in
- If `false`, new users are automatically created on first login

## Security Features

1. **Random local port**: Prevents port conflicts
2. **HttpOnly cookies**: JavaScript cannot access session tokens
3. **Secure cookies**: Transmitted only over HTTPS (in production)
4. **Short-lived redirect**: `redir` cookie used only once
5. **File permissions**: Credentials file is `0600` (user-only access)
6. **JWT validation**: All tokens are cryptographically verified

## Troubleshooting

### Login fails with "unregistered user"

- Server has `RejectNewUsers=true`
- Admin must invite user first

### Token not stored

- Check permissions on `~/.config/autokitteh/` directory
- Ensure write access to credentials file

### Browser doesn't open

- Check `BROWSER` environment variable
- Manually open the URL printed by CLI

### "Invalid port" error

- Port must be between 0-65535
- Check firewall/network settings blocking local connections
