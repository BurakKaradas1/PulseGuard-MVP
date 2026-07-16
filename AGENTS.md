# AGENTS.md

## Project map
- This workspace is a small PulseGuard C2 demo with three separate components:
  - [pulseguard-c2/main.go](pulseguard-c2/main.go): Go-based C2 server prototype.
  - [pulseguardv1/main.go](pulseguardv1/main.go): Go-based agent/v1 prototype.
  - [PulseGuard-Dashboard/dashboard.py](PulseGuard-Dashboard/dashboard.py): Python dashboard that polls the C2 server.
- The high-level overview is documented in [README.md](README.md).

## Working conventions
- Prefer small, focused changes and avoid introducing unrelated dependencies.
- Preserve the existing demo/security-oriented behavior unless a task explicitly requires a protocol or endpoint change.
- Keep the current HTTP flow intact when editing server/agent code: the server currently expects `/receive` and `/stats` endpoints and the dashboard reads `/stats` from `http://localhost:8080`.
- If you add new files, place them in the most relevant existing component folder rather than creating a new top-level structure.

## Run commands
- Start the C2 server:
  - `cd pulseguard-c2 && go run .`
- Start the v1 prototype:
  - `cd pulseguardv1 && go run .`
- Start the dashboard:
  - `cd PulseGuard-Dashboard && python dashboard.py`

## Notes for agents
- The repository currently uses separate module folders, so run commands from the appropriate subdirectory instead of the repo root.
- The Go modules are independent; do not assume a shared package layout between [pulseguard-c2](pulseguard-c2) and [pulseguardv1](pulseguardv1).
- If documentation needs to be updated, prefer linking to [README.md](README.md) rather than duplicating project context in multiple files.
