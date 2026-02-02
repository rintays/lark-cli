# Config Workflows

Config controls defaults like platform, mailbox, and app credentials.

## Show config

```bash
lark config info
```

## Set platform or base URL

```bash
lark config set --platform feishu
lark config set --base-url https://open.larksuite.com
```

## Set default mailbox ID

```bash
lark config set --default-mailbox-id <MAILBOX_ID>
```

## Set default token type

```bash
lark config set --default-token-type user
```

## Set app credentials

```bash
lark config set --app-id <APP_ID>
lark config set --app-secret <APP_SECRET>
```

## Unset values

```bash
lark config unset --base-url true
lark config unset --default-mailbox-id true
lark config unset --default-token-type true
lark config unset --default-user-account true
lark config unset --user-tokens true
```

## List supported keys

```bash
lark config list-keys
```
