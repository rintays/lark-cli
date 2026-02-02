# Wiki Workflows

Wiki organizes content into spaces and nodes. Many commands accept either tenant or user tokens, but space creation is user-token only.

## List spaces

```bash
lark wiki space list --limit 10
```

## Create a space (user token required)

```bash
lark wiki space create "Team KB" --space-type team --visibility private
```

## List nodes in a space

```bash
lark wiki node list --space-id <SPACE_ID> --limit 50
```

## Render a node tree

```bash
lark wiki node tree --space-id <SPACE_ID> --depth 3
```

## Create a node (link a Doc/Sheet/etc.)

```bash
lark wiki node create docx <DOCX_TOKEN> --space-id <SPACE_ID> --title "Design"
```

## Move a node

```bash
lark wiki node move <NODE_TOKEN> --space-id <SPACE_ID> --target-parent-node-token <PARENT_NODE_TOKEN>
```

## Update node title

```bash
lark wiki node update-title <NODE_TOKEN> "New Title" --space-id <SPACE_ID>
```

## Note on permissions

If a node points to a Drive object (doc/sheet/file), use `lark drive permissions` to manage collaborators for the underlying `obj_token`.
