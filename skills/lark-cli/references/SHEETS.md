# Sheets Workflows

Sheets commands accept a spreadsheet token or a spreadsheet URL.

## Read

```bash
lark sheets read <SHEET_TOKEN> "Sheet1!A1:C10"
```

Use a sheet ID plus a simple range:

```bash
lark sheets read <SHEET_TOKEN> A1:C10 --sheet-id <SHEET_ID>
```

## Update a range

Inline JSON values:

```bash
lark sheets update <SHEET_TOKEN> "Sheet1!A1:B2" \
  --values '[["Name","Score"],["Ada",42]]'
```

Values from file (JSON/CSV/TSV):

```bash
lark sheets update <SHEET_TOKEN> "Sheet1!A1:B2" --values-file values.json
```

## Append rows

```bash
lark sheets append <SHEET_TOKEN> "Sheet1!A1" --values '[["New","Row"]]'
```

## Clear

```bash
lark sheets clear <SHEET_TOKEN> "Sheet1!A1:C10"
```
