# Minutes Workflows

Minutes are meeting recordings/transcripts stored in Drive.

## List minutes

```bash
lark minutes list --limit 20
```

## Filter by folder or title

```bash
lark minutes list --folder-id <FOLDER_TOKEN> --query "Weekly" --limit 10
```

## Get minutes details

```bash
lark minutes info <MINUTE_TOKEN>
```

## Update sharing permissions

```bash
lark minutes update <MINUTE_TOKEN> --link-share tenant_readable --external-access=false
```

## Delete minutes

```bash
lark minutes delete <MINUTE_TOKEN>
```
