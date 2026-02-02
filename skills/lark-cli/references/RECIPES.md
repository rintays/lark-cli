# Recipes (Common Agent Tasks)

These examples are safe starting points. Prefer `--json` for automation.

## Who am I?

```bash
lark whoami
```

## List chats

```bash
lark chats list --limit 10
```

## Send a message

```bash
lark messages send <CHAT_ID> --text "hello"
```

## Search users

```bash
lark users search "Ada" --json
```

## Read a doc

```bash
lark docs get <DOCX_TOKEN> --format md
```

## Search Drive

```bash
lark drive search "Q1" --limit 10 --json
```

## Mail: list inbox

```bash
lark mail list --limit 10
```
