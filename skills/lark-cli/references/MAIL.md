# Mail Workflows

Mail APIs are user-scoped. You must use a user token.

## Login (user token)

```bash
lark auth user login
```

## List folders

```bash
lark mail folders
```

## List inbox messages

```bash
lark mail list --folder-id INBOX --limit 10
```

## Get message metadata

```bash
lark mail info <MESSAGE_ID>
```

## Get full message content

```bash
lark mail get <MESSAGE_ID>
```

## Send email (text)

```bash
lark mail send --to user@example.com --subject "Hello" --text "Hi"
```

## Send email (raw EML)

```bash
lark mail send --raw-file message.eml
```
