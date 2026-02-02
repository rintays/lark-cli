# Drive Workflows

Drive is the file storage layer for Docs, Sheets, Slides, Minutes, and uploaded files.

## List files in a folder

```bash
lark drive list --folder-id <FOLDER_TOKEN> --limit 20
```

## Search files by keyword

```bash
lark drive search "Q1" --limit 10 --json
```

If you hit permission errors, re-run with a user token:

```bash
lark auth user login
```

## Inspect a file

```bash
lark drive info <FILE_TOKEN>
```

## Get share URLs

```bash
lark drive urls <FILE_TOKEN>
```

## Download a file

```bash
lark drive download <FILE_TOKEN> --out ./file.bin
```

## Upload a file

```bash
lark drive upload ./report.pdf --folder-token <FOLDER_TOKEN>
```

## Manage permissions

```bash
lark drive permissions add <FILE_TOKEN> openid <OPEN_ID> --type docx --perm view --member-kind user
lark drive permissions list <FILE_TOKEN> --type docx
lark drive permissions update <FILE_TOKEN> openid <OPEN_ID> --type docx --perm edit
lark drive permissions delete <FILE_TOKEN> openid <OPEN_ID> --type docx
```
