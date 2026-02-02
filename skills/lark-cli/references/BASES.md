# Bitable Bases (Base) Workflows

Bitable is Lark/Feishu's database product. Most commands require a base app token.

## List bases (apps)

```bash
lark bases list --limit 10
```

## List tables in a base

```bash
lark bases table list --app-token <APP_TOKEN>
```

## Create a table

```bash
lark bases table create "Leads" --app-token <APP_TOKEN>
```

## Create a field

```bash
lark bases field create <TABLE_ID> --app-token <APP_TOKEN> \
  --name "Owner" --type user
```

## Create a record

```bash
lark bases record create <TABLE_ID> --app-token <APP_TOKEN> \
  --field Name=Acme --field Score:=42
```

## Search records

```bash
lark bases record search <TABLE_ID> --app-token <APP_TOKEN> --json
```
