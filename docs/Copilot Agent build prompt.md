---
description: Copy-paste prompt for GitHub Copilot Agent to implement Astrometrics Lab in a new repository.
---

# Copilot Agent build prompt

Use this prompt in the target Astrometrics Lab repository.

## Prompt

Build Astrometrics Lab as a Go CLI with Bubble Tea UI for managing GitHub stars and star lists.

Execution mode:
- Use one primary agent for implementation and testing.
- Only use sub-agents for independent read-only research tasks.

Primary goals:
- Manage stars and list memberships safely.
- Support dual auth: token first, gh fallback.
- Persist non-secret star and list cache between sessions.
- Never persist secrets in local JSON files.
- If keyring is unavailable, require no-persist auth mode.

Core commands:
- astlab auth login --method token|gh --store keyring|none
- astlab auth status
- astlab auth logout
- astlab lists
- astlab stars
- astlab move
- astlab sync
- astlab sync --full

Architecture constraints:
- Use direct GitHub GraphQL calls at runtime.
- Do not shell out to gh api for each request.
- It is acceptable to use gh auth token as a fallback credential source.
- Keep auth provider abstraction explicit.
- Keep sync and mutation logic separate from Bubble Tea UI logic.

Auth and security constraints:
- Auth selection precedence:
  - Explicit auth flag
  - GITHUB_TOKEN
  - Stored keyring token
  - gh fallback token
- Validate auth at startup with viewer login query.
- For mutation commands, validate required capabilities and return actionable errors.
- Redact secrets from logs and error output.

Sync and reliability constraints:
- Implement full pagination for:
  - viewer.starredRepositories
  - viewer.lists
  - list.items
- Implement bounded retries for transient failures with exponential backoff and jitter.
- Do not retry hard auth failures or malformed query errors.
- Delta sync uses starredAt.
- Full reconciliation via astlab sync --full.

Mutation safety constraints:
- For updateUserListsForItem:
  - Read current memberships first.
  - Compute before/after diff.
  - Default batch operations to dry run.
  - Require --apply to execute batch changes.
  - Require explicit confirmation if removals are detected.
  - Keep operations idempotent where possible.

Local storage constraints:
- Persist non-secret state only:
  - metadata.json
  - stars.json
  - lists.json
  - memberships.json
- Include schemaVersion in metadata.
- Use atomic file writes.

Testing requirements:
- Unit tests:
  - query builders
  - response normalization
  - sync decision logic
  - membership diff logic
- Fixture-based integration tests for GraphQL response handling.
- CLI command tests for auth, lists, stars, move, sync, sync --full.
- Auth backend tests for:
  - keyring available
  - no keyring -> no-persist fallback

MVP done criteria:
- astlab auth status reports provider and account login.
- astlab lists shows list names, IDs, and item counts.
- astlab stars supports stable pagination.
- List item sync handles multi-page lists.
- astlab move supports dry run and apply flow with before/after diff.
- astlab sync performs delta updates.
- astlab sync --full reconciles drift.
- JSON output mode exists for automation.
- Tests pass in CI.

Reference links:
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Charm ecosystem: https://charm.land
- Bubbles: https://github.com/charmbracelet/bubbles
- Lip Gloss: https://github.com/charmbracelet/lipgloss
- Huh: https://github.com/charmbracelet/huh
- GitHub CLI manual: https://cli.github.com/manual/
- GitHub GraphQL docs: https://docs.github.com/en/graphql
- GraphQL Explorer: https://docs.github.com/en/graphql/overview/explorer
- Token guidance: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens
- Go keyring library: https://github.com/99designs/keyring

Implementation expectations:
- Start by generating a concise plan.
- Implement in small commits or logically grouped changes.
- Run tests frequently.
- End with a short summary that maps delivered behavior to MVP done criteria.
