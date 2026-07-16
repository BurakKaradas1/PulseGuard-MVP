# Copilot instructions

This repository contains a compact PulseGuard C2 demo spread across separate Go and Python subprojects.

## Important context
- Use [README.md](../README.md) as the primary source for product context.
- The main executable entry points are [pulseguard-c2/main.go](../pulseguard-c2/main.go) and [pulseguardv1/main.go](../pulseguardv1/main.go).
- The dashboard is [PulseGuard-Dashboard/dashboard.py](../PulseGuard-Dashboard/dashboard.py) and expects the Go server to be reachable at `http://localhost:8080/stats`.

## Agent guidance
- Keep changes localized to the relevant component folder.
- Preserve existing endpoint names and protocol assumptions unless the task explicitly requests an architectural change.
- Favor minimal edits that fit the current demo style and avoid introducing new frameworks or large refactors.
- When adding or editing docs, link back to [README.md](../README.md) instead of duplicating broad project descriptions.
