# Tasks Workflows

Tasks are personal or shared work items.

## List my tasks

```bash
lark tasks list --limit 10
```

## Create a task

```bash
lark tasks create --summary "Write report" --due 2026-02-05T12:00:00Z
```

## Update a task

```bash
lark tasks update <TASK_GUID> --summary "Write report v2"
```

## Complete a task

```bash
lark tasks update <TASK_GUID> --completed-at 2026-02-03T12:00:00Z
```
