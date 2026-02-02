# Chats Workflows

Chats are conversations the bot can access.

## List chats

```bash
lark chats list --limit 10
```

## Create a chat

```bash
lark chats create --name "Project Alpha" --user-id <OPEN_ID>
```

## Get chat details

```bash
lark chats get <CHAT_ID>
```

## Update chat name

```bash
lark chats update <CHAT_ID> --name "New Name"
```

## Read or update announcements

```bash
lark chats announcement get <CHAT_ID>
lark chats announcement update <CHAT_ID> --revision 12 --request '{"requestType":"InsertBlocksRequestType"}'
```
