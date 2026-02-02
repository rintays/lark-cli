# Users Workflows

Users are people in your tenant directory. Search requires a user token.

## Search users

```bash
lark users search "Ada" --limit 10
```

Search by email:

```bash
lark users search --email "ada@example.com" --limit 10
```

## Get contact user info

```bash
lark users info --user-id <USER_ID>
```
