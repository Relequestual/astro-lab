---
description: Feasibility and implementation notes for Astrometrics Lab, a CLI for managing GitHub stars and star lists via GitHub GraphQL with optional GitHub CLI auth fallback.
---

# Astrometrics Lab - feasibility and technical notes

## Contents

- [Decision Summary](#decision-summary)
- [What Was Verified](#what-was-verified)
- [Human Guide](#human-guide)
- [Agent Playbook](#agent-playbook)
- [Authentication Strategy](#authentication-strategy)
- [Secret Storage Across OS](#secret-storage-across-os)
- [API Limits and Retries](#api-limits-and-retries)
- [Pagination Requirements](#pagination-requirements)
- [Mutation Safety Guardrails](#mutation-safety-guardrails)
- [Data and Sync Strategy](#data-and-sync-strategy)
- [Storage Recommendation](#storage-recommendation)
- [Project Plan](#project-plan)
- [MVP Acceptance Checklist](#mvp-acceptance-checklist)
- [Agent Execution Strategy](#agent-execution-strategy)
- [AI Builder Prompt Template](#ai-builder-prompt-template)
- [Bubble Tea and Language Choice](#bubble-tea-and-language-choice)
- [External Resources](#external-resources)

## Decision Summary

Astrometrics Lab is feasible with low to medium implementation risk.

The best architecture is a Go CLI with Bubble Tea UI, direct GitHub GraphQL calls, and dual auth providers.

Preferred auth order:
- Explicit auth flag.
- Environment token.
- Stored token from keychain.
- GitHub CLI token fallback.

A hybrid sync approach is recommended:
- Fast delta sync for new stars since last sync time.
- Full baseline sync on first run, then periodic reconciliation.

## What Was Verified

These capabilities were confirmed with live queries on the authenticated account:

- Lists are available on the GraphQL viewer object.
- List metadata includes id, name, slug, description, privacy, updatedAt, and lastAddedAt.
- List items can be queried, including repository id and nameWithOwner.
- Mutations exist for:
  - addStar
  - removeStar
  - createUserList
  - updateUserList
  - deleteUserList
  - updateUserListsForItem
- Star timestamps are available through viewer.starredRepositories edges.starredAt.

## Human Guide

Use direct GitHub API requests as the default transport.

Prerequisites:
- One valid GitHub auth source.

Supported auth sources:
- API token provided by the user.
- Existing gh authentication as fallback.

Recommended command style:
- Direct API in app runtime.
- gh-based examples only for manual verification and debugging.

Example read query for lists and counts:

```bash
GH_PAGER=cat gh api graphql -f query='query {
  viewer {
    lists(first: 100) {
      nodes {
        id
        name
        slug
        description
        isPrivate
        items(first: 1) { totalCount }
      }
    }
  }
}'
```

## Authentication Strategy

Support both auth methods.

- token provider:
  - Read `GITHUB_TOKEN` first, or prompt user and store in OS keychain.
  - Use this as first-class auth for all runtime API requests.
- gh provider:
  - Use `gh auth token` as a convenience fallback.
  - Resolve token once, then use direct API requests.

Do not shell out to `gh api` for every request in the Go app.

Recommended command set:
- `astlab auth login --method token`
- `astlab auth login --method gh`
- `astlab auth status`
- `astlab auth logout`

Startup auth validation:
- Validate token usability at startup with a lightweight `viewer { login }` query.
- If write operations are requested, validate that star and list mutations are permitted.
- Return actionable errors for insufficient scopes or missing capabilities.

Security rules:
- Never store tokens in plain JSON files.
- Store secrets in keychain, store only metadata in local state files.
- Redact tokens from logs and error output.

## Secret Storage Across OS

Support a backend chain instead of assuming macOS keychain.

Recommended backend order:

- OS credential store through a cross-platform keyring library.
  - macOS: Keychain.
  - Windows: Credential Manager.
  - Linux: Secret Service, KWallet, or pass backend.
- No-persist mode for fully ephemeral sessions.

Implementation guidance:

- Keep secrets out of `metadata.json`, `stars.json`, `lists.json`, and `memberships.json`.
- Persist only non-secret auth metadata, for example provider type and account login.
- Offer explicit command controls:
  - `astlab auth login --method token --store keyring`
  - `astlab auth login --method token --store none`
- In CI or headless environments, prefer `GITHUB_TOKEN` with `--store none`.

## API Limits and Retries

Rate limits and transient API failures are a first-class concern.

Policy:
- Parse GraphQL errors and HTTP status codes for rate-limit and abuse signals.
- Retry transient failures with bounded exponential backoff and jitter.
- Do not retry on hard auth errors or malformed query errors.
- Surface remaining rate budget in verbose logs when available.

Recommended defaults:
- Max retries: 4.
- Base delay: 500ms.
- Max delay: 8s.
- Max concurrent page fetches: 1 for mutation flows, small bounded pool for read-only sync.

## Pagination Requirements

All connection-style reads must be fully paginated.

Required pagination loops:
- `viewer.starredRepositories` for full star baselines and long delta windows.
- `viewer.lists` when account list count exceeds first page.
- `list.items` for every list included in full sync or targeted refresh.

Implementation rule:
- Continue until `pageInfo.hasNextPage` is false.
- Persist cursor progress in memory during a run.
- Treat partial-page failures as retryable units, not global reset events.

## Mutation Safety Guardrails

Mutations must be protected by explicit safety checks.

Rules:
- For `updateUserListsForItem`, fetch current list memberships first.
- Compute a before and after diff and display it in dry run mode.
- Require explicit confirmation when any membership removal is detected.
- Support a non-interactive force flag for automation workflows.
- Make mutation operations idempotent when repeated with unchanged inputs.

Recommended command behavior:
- Default to dry run for batch operations.
- Require `--apply` to execute batch mutation changes.

Example read query for latest starred repositories:

```bash
GH_PAGER=cat gh api graphql -f query='query {
  viewer {
    starredRepositories(first: 50, orderBy: { field: STARRED_AT, direction: DESC }) {
      edges {
        starredAt
        node {
          id
          nameWithOwner
        }
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}'
```

## Agent Playbook

### 1) Resolve list IDs

First read all lists and cache name to id mapping.

```graphql
query Lists {
  viewer {
    lists(first: 100) {
      nodes {
        id
        name
        slug
        updatedAt
        lastAddedAt
      }
    }
  }
}
```

### 2) Resolve repository IDs

For assignment operations, identify repository node IDs from starredRepositories.

```graphql
query StarredPage($first: Int!, $after: String) {
  viewer {
    starredRepositories(first: $first, after: $after, orderBy: { field: STARRED_AT, direction: DESC }) {
      edges {
        starredAt
        node {
          id
          nameWithOwner
        }
      }
      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
}
```

### 3) Create list

```graphql
mutation CreateList($input: CreateUserListInput!) {
  createUserList(input: $input) {
    list {
      id
      name
      slug
      isPrivate
    }
  }
}
```

Input shape:
- name: String, required
- description: String, optional
- isPrivate: Boolean, optional

### 4) Update list metadata

```graphql
mutation UpdateList($input: UpdateUserListInput!) {
  updateUserList(input: $input) {
    list {
      id
      name
      slug
      description
      isPrivate
      updatedAt
    }
  }
}
```

Input shape:
- listId: ID, required
- name, description, isPrivate: optional

### 5) Delete list

```graphql
mutation DeleteList($input: DeleteUserListInput!) {
  deleteUserList(input: $input) {
    clientMutationId
  }
}
```

Input shape:
- listId: ID, required

### 6) Assign list membership for one item

```graphql
mutation UpdateListsForItem($input: UpdateUserListsForItemInput!) {
  updateUserListsForItem(input: $input) {
    item {
      __typename
      ... on Repository {
        id
        nameWithOwner
      }
    }
  }
}
```

Input shape:
- itemId: ID, required
- listIds: [ID!]!, required
- suggestedListIds: [ID!], optional

Important behavior note:
- Treat this operation as set membership semantics.
- Read current memberships first, then compute desired memberships.
- Apply with dry run and confirmation for batch operations.

### 7) Star and unstar

```graphql
mutation AddStar($input: AddStarInput!) {
  addStar(input: $input) {
    starrable {
      __typename
      id
    }
  }
}
```

```graphql
mutation RemoveStar($input: RemoveStarInput!) {
  removeStar(input: $input) {
    starrable {
      __typename
      id
    }
  }
}
```

## Data and Sync Strategy

Short answer:
- You can do incremental updates for stars using starredAt.
- You should still run periodic full reconciliation for correctness.

Recommended model:

- First run:
  - Full fetch of all starred repositories.
  - Full fetch of all lists and all list items.
  - Persist normalized local snapshot.

- Normal run:
  - Fetch newly starred repos by walking starredRepositories in descending STARRED_AT until reaching lastSyncedAt.
  - Refresh all list metadata.
  - Refresh list items only for lists whose lastAddedAt or updatedAt changed since last sync.

- Reconciliation run:
  - Full sync every N days or when anomalies are detected.

Why this is practical:
- Saves API calls for day to day usage.
- Keeps eventual correctness with periodic full refresh.

Potential blind spots to guard against:
- Unstars older than lastSyncedAt are not visible in a pure append-only delta approach.
- List removals may not be inferable from timestamps alone if state changed outside tool assumptions.

Mitigation:
- Track known repo IDs and detect absences during periodic full sync.
- Offer a user command: astlab sync --full.

## Storage Recommendation

Your proposed approach is reasonable and wise.

Use in-memory state during a command plus a small local object store between sessions.

Suggested local layout:

- metadata.json
  - schemaVersion
  - accountLogin
  - lastSyncedAt
  - lastFullSyncAt

- stars.json
  - byRepoId: repo metadata and starredAt
  - byNameWithOwner index

- lists.json
  - byListId: list metadata
  - bySlug index

- memberships.json
  - listId to repoId set
  - repoId to listId set

Operational notes:
- Keep writes atomic by writing temp files then rename.
- Keep this store small and human inspectable.
- Add a cache bust command: astlab cache clear.

## Project Plan

### Phase 1: Foundation

- Define command surface:
  - astlab auth
  - astlab lists
  - astlab stars
  - astlab move
  - astlab sync
- Build GitHub GraphQL adapter with auth provider abstraction.
- Implement token and gh fallback auth providers.
- Implement startup auth capability checks and clear scope errors.
- Implement read-only list and star commands.

### Phase 2: Mutations

- Implement list create, update, delete.
- Implement star add and remove.
- Implement repo list assignment with dry run.

### Phase 3: Sync and cache

- Implement local object store and schema versioning.
- Add incremental sync strategy.
- Add full sync command and reconciliation checks.
- Implement full pagination for stars, lists, and list items.
- Implement retry and backoff policy for transient API failures.

### Phase 4: UX quality

- Add robust error messages and retry guidance.
- Add JSON output mode and script-friendly exit codes.
- Add minimal telemetry log file for troubleshooting.
- Add auth diagnostics in `astlab auth status`.

## MVP Acceptance Checklist

The MVP is done when all items below are true.

- Auth supports both token and gh fallback paths.
- `astlab auth status` clearly reports active provider and account login.
- `astlab lists` shows list names, IDs, and item counts.
- `astlab stars` shows starred repositories with stable pagination behavior.
- List item sync handles multi-page lists correctly.
- `astlab move` supports dry run and confirm mode.
- Batch mutation flows require `--apply` and show before and after diffs.
- `astlab sync` performs delta updates using starredAt.
- `astlab sync --full` rebuilds local state and reconciles drift.
- Secrets use OS keyring when available, otherwise no-persist mode, and are never written in plain local JSON files.
- JSON output mode is available for automation.
- Startup auth scope and capability checks produce actionable failures.
- Test suite includes unit tests, fixture-based integration tests, CLI command tests, and auth backend behavior tests for keyring available vs no-persist fallback.

## Agent Execution Strategy

Default to one primary agent for implementation.

Guidelines:
- A document around 500 lines is not large enough to require forced task sharding.
- Use one agent for coding, testing, and refinement when task flow is linear.
- Use sub-agents only for independent workstreams that can run in parallel, for example broad repo research, separate docs lookup, or independent test fixture generation.
- Avoid creating one sub-agent per small task. This adds overhead and often harms coherence.

Practical rule of thumb:
- If tasks share state heavily, keep one agent.
- If tasks are independent and read-only, split to sub-agents.

## AI Builder Prompt Template

Use this prompt as the baseline for implementation runs.

```text
Build Astrometrics Lab as a Go CLI with Bubble Tea UI for managing GitHub stars and star lists.

Product goals:
- Manage stars and list memberships safely.
- Support dual auth: token first, gh fallback.
- Persist non-secret star and list cache, but never persist secrets outside OS keyring.
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

Safety constraints:
- Default batch mutations to dry run.
- Require --apply to execute batch mutations.
- For updateUserListsForItem, fetch current memberships first and show before/after diff.
- Require explicit confirmation if any list membership removal is detected.

API and reliability constraints:
- Use direct GitHub GraphQL calls in runtime.
- Do not shell out to gh api for each request.
- Implement pagination for starredRepositories, lists, and list.items.
- Implement bounded retries with exponential backoff for transient failures.
- Validate auth capability at startup with viewer login check and write-capability checks for mutation flows.

Testing requirements:
- Unit tests for query building, normalization, sync decisions, and membership diffs.
- Fixture-based integration tests for GraphQL response handling.
- CLI command tests for auth, lists, stars, move, sync, and sync --full.
- Auth backend tests for keyring available and no-persist fallback.

Execution strategy:
- Use one primary implementation agent.
- Use sub-agents only for independent parallel research tasks.

Done criteria:
- All MVP Acceptance Checklist items in this spec are satisfied.
```

## Bubble Tea and Language Choice

Astrometrics Lab uses Go with Bubble Tea and Charm ecosystem libraries.

Bubble Tea is a Go framework, so Node.js cannot use it natively as a library.

If a Node.js version is ever needed, it would require a separate Node TUI stack.

Node.js alternatives:
- Ink (React style terminal UI).
- Blessed and Blessed Contrib.
- Enquirer or prompts for lighter interactive flows.

Current recommendation:
- Keep the project on Go for first release.
- Keep dual auth support so users can use API token or existing gh login.

## External Resources

These links should be referenced directly in implementation prompts so AI agents can retrieve authoritative docs.

- Bubble Tea framework: [https://github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
- Charm ecosystem index: [https://charm.land](https://charm.land)
- Bubbles component library: [https://github.com/charmbracelet/bubbles](https://github.com/charmbracelet/bubbles)
- Lip Gloss styling library: [https://github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
- Huh form library: [https://github.com/charmbracelet/huh](https://github.com/charmbracelet/huh)
- GitHub CLI manual: [https://cli.github.com/manual/](https://cli.github.com/manual/)
- GitHub GraphQL docs: [https://docs.github.com/en/graphql](https://docs.github.com/en/graphql)
- GitHub GraphQL Explorer: [https://docs.github.com/en/graphql/overview/explorer](https://docs.github.com/en/graphql/overview/explorer)
- GitHub token guidance: [https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
- Go keyring library: [https://github.com/99designs/keyring](https://github.com/99designs/keyring)

Notes for AI agents:
- `charm.land` is a website URL for the Charm ecosystem, not a package name.
- Bubble Tea is one part of a suite. Most production TUIs combine Bubble Tea with Bubbles and Lip Gloss.
- Prefer official docs and repositories above for API signatures and examples.
