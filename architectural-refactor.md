# Architectural Refactor Status

## Summary

The refactor described in this repo has effectively landed in a simpler form than the original migration plan.

The current architecture is:

- `main.go` for process startup, repo validation, and final selected-path output
- `app/` for state, business flow, effects, and async result handling
- `ui/` for Bubble Tea integration, rendering, and screen-local widget state
- `worktree/` and `branch/` as the Git-facing service packages

The earlier plan for a separate in-app docs feature and `docs/` service is no longer relevant.

## What The Refactor Achieved

- `worktree.Manager` was moved behind the `worktree.Service` interface
- a dedicated `branch.Service` now owns branch-specific Git behavior
- `app.App` sits between services and the Bubble Tea runtime
- reusable Bubble Tea components live under `ui/components/`
- screen key handling is organized around normalized keys and per-screen binding maps

## Current Layer Responsibilities

### `main`

- parse startup flags
- construct services
- construct `app.App`
- run Bubble Tea
- preserve the selected-path `stdout` contract

### `app`

- own application state
- own screen transitions
- own loading, dialog, and status semantics
- return typed effects for async work
- consume typed async results

### `ui`

- adapt `app` to Bubble Tea
- render screens
- hold local widget state
- execute service-backed effects

### `worktree`

- worktree list/add/remove operations
- branch existence checks for add flow
- porcelain parsing and enrichment

### `branch`

- local and remote-tracking branch listing
- branch scope toggling
- branch checkout
- branch deletion and bulk deletion
- branch fetch
- recent commit previews

## Current Screen Set

- change worktree screen
- add worktree screen
- branch management screen

There is no docs screen or docs service anymore.

## Ongoing Guidance

When extending the app:

- keep Bubble Tea concerns in `ui/`
- keep business decisions in `app/`
- keep Git subprocess behavior in `worktree/` or `branch/`
- prefer typed effects and typed results over direct service calls from screen handlers
- keep rendering side-effect free

## Notes

This file now serves as a short status note rather than a step-by-step migration plan, since the original plan diverged from the final implementation.
