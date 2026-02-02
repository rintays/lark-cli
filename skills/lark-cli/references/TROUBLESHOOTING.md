# Troubleshooting

## Permission / scope errors

Symptoms:
- 401/403 errors
- error text mentioning missing scopes or permissions

Fix:
- Use a user token for user-scoped APIs: `lark auth user login`
- Re-login and grant missing scopes
- Verify the account: `lark auth status --json`

## Command expects positional IDs

Many commands require IDs as positional args (not `--id`). Example:

```bash
lark docs info <DOC_ID>
```

## Not sure which token to use

- Tenant token: app/admin operations
- User token: Drive search, Mail send, user mailboxes, etc.

If unsure, try user token first for user data.

## JSON parsing errors

- Ensure you pass `--json`.
- Logs/errors go to stderr; parse stdout only.
