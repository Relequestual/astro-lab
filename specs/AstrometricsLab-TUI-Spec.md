# Astrometrics Lab — TUI Product Specification

**Purpose**  
Design and implement a highly interactive Bubble Tea-based terminal UI for managing hundreds of GitHub stars and star lists. The TUI must be smooth, beautiful, and a joy to use for power users organizing large star collections.

## Entry and Launch
- On launch, if no local cache exists, the TUI guides the user through setup and authentication (inline, not error message).
- On launch, if a cache exists, the TUI:
  - Checks if data is stale and shows an "Enter to sync" prompt if out of date.
  - Pressing Enter triggers sync.
- The TUI should always favor a welcoming, informative layout.

## Navigation Structure
- **Dashboard (Home) Screen**:  
  - Centrally positioned "button"-like selectors for all primary actions (see below).
  - Displays key summary stats (e.g., total starred repos, number of lists).
- **Main Views/Screens**:
  1. **Dashboard** – the entry hub with navigation to major workflows.
  2. **Lists Panel** – shows all user lists; filterable and supports creation/renaming/deletion.
  3. **Repos-in-List Panel** – shows all repos in a selected list.
  4. **Repo Detail View** – shows full details for a single repository, actions, and metadata.
  5. **All Stars (Unorganized)** – shows all starred repositories, with rich repo metadata and filter/sort controls.
- **Layout Options**:
  - Both split-pane and full-screen navigation should be supported.  
  - Implementation should use established UI best practices and principles to determine the best default, and adapt as the project evolves.
- **Navigation**:
  - Use arrow keys for focus and selection. Never require Vim-style.
  - Allow users to *batch select* multiple repos for operations.

## Core Workflows & Functionality
- **List Management:**  
  - Browse, create, rename, and delete lists.  
  - Deleting lists must trigger an explicit "are you sure?" warning.
- **Repo Organization:**  
  - View, search, and filter repos in any list or unorganized repos.
  - Add a repo to one or more lists (multi-select target lists).
  - Remove repos from lists.
  - Move repos between lists (with the ability to add to multiple).
  - Batch operations are fully supported.
- **Sync:**  
  - Manual sync from within the TUI (`s` or onscreen button); progress is indicated in a non-blocking way. UI remains responsive.
- **Dry-Run Workflow:**  
  - All modifying actions (list/repo changes) produce a dry-run/preview panel showing the *exact* effects before user confirmation. User must explicitly confirm or cancel.
- **Undo:**  
  - Provide an "undo" feature for mutations where possible.  
    - This should allow the user to revert the last mutation(s) (add/remove/move/list changes) if feasible given data state/caching.
    - Undo complexity may depend on how the mutation and caching system is designed, but it is desirable; design for it early.
- **Information Richness:**  
  - Show, at a minimum, for each repo: name, description, language, star count, fork count, and URL.

## Search, Filter, and Sort
- Fuzzy, incremental search available from any repo list (use `/` key to activate; design for familiar TUI UX).
- Support for filtering:
  - By list membership (e.g., "not in any list")
  - By language, star date, and any available repo metadata.
- Sorting is available on all columns/fields with intuitive keys and visual cues.
- Result sets (e.g., 1000+ stars) may be virtually paged or streamed—implement per UI best practices.

## Visual Style & Guidance
- The visual "feel" should make management easy even with hundreds/thousands of stars.
- No explicit color palette required at this time; use good defaults and respect terminal themes where possible.
- Strive for the UX coherence and delight of `lazygit` with the visual clarity/niceness of `superfile` (Charmbracelet's Superfile project).
- Use Bubble Tea (`bubbletea`), Bubbles, and Lip Gloss (`lipgloss`) for appearance and interaction.
- Persistent context bar (header/footer) with:
  - Logged-in user
  - GitHub rate limit info
  - Last sync time
  - Key context actions when relevant
- "?" key always brings up a keybinding/feature help overlay.
- All navigation and interactive elements should be discoverable and never require hidden/unusual controls.
- Panel minimum/maximum sizing, and resizing, should follow TUI best practices and be adapted as the project matures.

## Authentication
- Auth is first-class: TUI must handle setup, token entry (before any action can be taken), and storage (to keyring/env as supported by existing code).
- Guides first-time users as part of the main flow, not as a side-error.

## Outstanding / Implementation Questions
- For layout, paging, and sizing, follow modern TUI best practices. Adapt default choices as the user base and usability testing suggest.
- **Undo support:** Strive to implement for all mutations (add/remove/move/list changes), but if architectural constraints make it difficult for any mutation type, document this and note fallback to dry-run and confirmation.
- "Are you sure?" confirmations are mandatory for all destructive actions (primarily deletions).

---

## Implementation Guidance for Agents
- All interactive logic must be written to maximize responsiveness (never block UI).
- Compose workflows as discrete, visually clear "flows" (select, preview, confirm, execute).
- Use the model types (`internal/models/types.go`) and GitHub code (`internal/github/queries.go`) as ground truth for data and operations.
- Avoid assumptions about auth/token handling outside what exists: always relay errors with actionable advice.
- Heavily comment data-fetch, dry-run, and sync logic to ensure testability and future expansion.
- Prioritize accessibility and usability: ensure clear focus and feedback for all TUI actions.
- Write all new code in idiomatic Go and project style.
- Place all new Bubble Tea code in `internal/tui/`, but refactor/divide by screen/component as complexity grows.
- Keep this spec up to date as further questions are resolved or additional requirements emerge.

---

_This document is the living source of truth for Astrometrics Lab TUI development. Keep it up-to-date if decisions change or more questions are resolved._
