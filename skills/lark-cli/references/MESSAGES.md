# Messages Workflows

Messages are chat messages sent to chats or users.

## Send a message

```bash
lark messages send <RECEIVE_ID> --receive-id-type chat_id --text "hello"
```

## List messages in a chat

```bash
lark messages list <CHAT_ID> --limit 10
```

## Search messages by keyword

```bash
lark messages search "incident" --limit 10
```

## Reply to a message

```bash
lark messages reply <MESSAGE_ID> --text "got it"
```

## Add a reaction

```bash
lark messages reactions add <MESSAGE_ID> thumbs_up
```

## Pin or unpin a message

```bash
lark messages pin <MESSAGE_ID>
lark messages unpin <MESSAGE_ID>
```
