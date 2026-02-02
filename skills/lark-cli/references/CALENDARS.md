# Calendars Workflows

Calendars contain events. Times are RFC3339 (UTC or with timezone offset).

## List events

```bash
lark calendars list --limit 10 --start 2026-02-01T00:00:00Z --end 2026-02-08T00:00:00Z
```

## Search events by keyword

```bash
lark calendars search "standup" --limit 10 --start 2026-02-01T00:00:00Z --end 2026-02-08T00:00:00Z
```

## Get event details

```bash
lark calendars get <EVENT_ID>
```
