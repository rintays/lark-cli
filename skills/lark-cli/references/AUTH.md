# Auth & Tokens

## Token types

- Tenant token: app-only access using your bot/app identity (admin/app-level APIs).
- User token: user-scoped access on behalf of a specific user (Drive search, Mail send, user mailbox, etc.).

## Store app credentials

```bash
lark auth login --app-id <APP_ID> --app-secret <APP_SECRET>
```

Optionally store the app secret in the OS keychain:

```bash
lark auth login --app-id <APP_ID> --app-secret <APP_SECRET> --store-secret-in-keyring
```

## Get tokens

Tenant token:

```bash
lark auth tenant
```

User token:

```bash
lark auth user login
```

## Multiple accounts

Use `--account` or `LARK_ACCOUNT` to select a user account.

## Scope errors

If a command fails with missing permissions, check:

- Whether it requires a user token.
- Whether the account granted required scopes.

See `references/TROUBLESHOOTING.md` for quick fixes.
