# Meetings Workflows

Meetings are scheduled video meetings.

## List meetings (defaults to last 6 months)

```bash
lark meetings list --limit 10
```

## Get meeting details

```bash
lark meetings info <MEETING_ID>
```

## Create a meeting (basic)

```bash
lark meetings create --topic "Weekly" --end-time 2026-02-03T01:30:00Z
```

## Update a meeting reservation

```bash
lark meetings update <RESERVE_ID> --topic "Weekly Sync"
```

## Delete a meeting reservation

```bash
lark meetings delete <RESERVE_ID>
```
