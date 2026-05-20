# Quickstart: Identity API CLI Commands

**Feature**: 003-identity-api-cli

## Common Workflows

### List all tracked users

```bash
atadmin users list --filter tracked
```

### Search for a user by email

```bash
atadmin users list --search alice@example.com --search-type email
```

### Inspect a user (get their ID and revision)

```bash
atadmin users get 12345
```

### Update a user's display name

```bash
# Revision is auto-fetched; no need to pass it manually
atadmin users update 12345 --display-name "Alice Smith"
```

### Set a user's timezone

```bash
atadmin users update 12345 --timezone "America/Chicago"
```

### Stop tracking a user

```bash
atadmin users update 12345 --tracked=false
```

### Add a user to a group

```bash
atadmin users groups add 12345 42
```

### Add a user to multiple groups

```bash
atadmin users groups add 12345 --group-ids 42,43,44
```

### Remove a user from a group

```bash
atadmin users groups remove 12345 42
```

### Stop tracking multiple users at once

```bash
atadmin users bulk stop-tracking --ids 12345,12346,12347
```

### Delete a user (interactive)

```bash
atadmin users delete 12345
# Prompts: Delete user 12345? [y/N]
```

### Delete a user (script / non-interactive)

```bash
atadmin users delete 12345 --yes
```

### List all agent devices

```bash
atadmin agents list
```

### Export all users as JSON (for scripting)

```bash
atadmin users list --json | jq '.results[] | {id, displayName: .displayName.value, status}'
```

### Pipe user IDs into another command

```bash
# Get IDs of all unlicensed users
atadmin users list --filter unlicensed --json | jq -r '.results[].id'
```

## Performance Tips

- Pass `--revision <n>` to skip the auto-fetch round-trip when you already have a fresh revision from a prior `get` call.
- Use `--cursor` with `--limit` for paginated exports of large accounts.
- For bulk operations on many users, `users bulk` sends one API call instead of N sequential update calls.
