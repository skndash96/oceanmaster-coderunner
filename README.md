# Ocean Master - Code Runner
This repository is responsible for running matches in a sandboxed environment.

## Workflow

### 1) Entry Point
The journey begins in `cmd/runner/main.go`. At startup, the program:

- Collects configuration from a centralized place (a config package) so all paths, resource limits, and timeouts are controlled in one location. This ensures that nsjail paths, jail mount points, cgroup policy, and timeouts remain consistent across the service.
- Initializes cgroup and jail prerequisites and then writes an `nsjail.cfg` from the centralized configuration. This file encodes CPU/memory/pids limits, tmpfs configuration, mount bindings, and the executable command that starts the Python wrapper inside the jail.

This ensures a deterministic sandbox environment before any matches are started.

### 2) Consuming incoming requests from RabbitMQ
After the initial setup, the service consumes match requests from a RabbitMQ queue. Each message includes enough data to start a match (match ID, player identifiers, and code references—either inline code strings or URLs).

For each incoming message, the program immediately starts match handling in a new goroutine. This is critical:
- Concurrency is achieved by launching a goroutine per request, allowing multiple matches to run in parallel without blocking the consumer thread.
- The concurrency level is bounded by configuration (e.g., max concurrent matches); It should be possible in RabbitMQ consumer to limit the maximum number of pending (un-acknowledged) requests.

### 3) Game Manager: the controller
The main goroutine delegates the match lifecycle to a Game Manager. Conceptually, the Game Manager:

- Maintains a count of ongoing matches.
- Receives the match request:
- Creates resources (files and dirs) for code and logs
- Saves the player code (either inline code strings or URLs)
- Creates a match-specific log file, which becomes the sink for structured JSON logs emitted during simulation.
- Calls the Game Engine
- Cleans up resources after the match completes

### 5) Engine simulation begins
With folders and logs ready, the Game Manager calls the engine’s `Simulate` method to run the match:

- The engine starts two nsjail sandboxes—one for each player—each bind-mounting the corresponding player directory as read-only inside the jail.
- Contexts (with timeouts) are used for critical phases like:
  - The global wall-time budget for the sandbox process.
  - The initial **HANDSHAKE** timeout (waiting for `"__READY__"` from each player’s Python wrapper).
  - Per-turn tick timeouts (waiting for the player’s actions).
- Stderr from each sandbox is streamed concurrently and logged. This is done in background goroutines so action processing is not blocked by error IO.

### 6) Turn loop until the game ends
Once both players successfully handshake:

- The engine enters a turn loop, alternating between Player 1 and Player 2.
- On each turn:
  - The engine sends the current `GameState` (as a single JSON line) to the active player’s stdin.
  - The player’s Python code computes actions by implementing `on_tick()` and returns a list of `Action` objects (JSON).
  - The engine receives those actions and applies them to produce the next `GameState`.
- The loop continues until an end condition is met (e.g., a tick limit or a game-specific victory state) or an error occurs (like timeout or invalid output). Current policy ends the match immediately on a turn error, but this can be adjusted to tolerate N consecutive failures if desired.

Concurrency remains central:
- Each match runs in its own goroutine, fully isolated from other matches.
- Error streaming uses goroutines to continuously read stderr without blocking the main turn loop.
- Shared resources (like the Game Manager’s registry) are protected by synchronization where needed.

### 7) Post-match actions
When the loop ends:
- The engine completes, and the Game Manager finalizes the match: flushes logs, optionally uploads them, and then removes temporary files/directories.
- The Game Manager updates its ongoing match registry, decreasing the active count and freeing capacity for new requests.

This leaves the system ready to process the next RabbitMQ message and spin up the next match goroutine.

## Modifying or Replacing the Game
The engine and game logic are intentionally generic. You control gameplay by editing or swapping the game-specific parts under the engine’s domain (e.g., `GameState`, `Action`, and the update rules). By modifying the engine/game, you can:

- Change how `GameState` evolves with each set of actions.
- Adjust end conditions, validation, and how turns are processed.
- Introduce richer action types, complex state, and multi-step semantics per tick.

Because the sandbox protocol is simply line-oriented JSON for state in and actions out, the surrounding orchestration—RabbitMQ, goroutines, temp folders, nsjail—stays the same while the core game changes.

## Usage

Prerequisites:
- Go 1.21+ (or compatible with the version pinned in `go.mod`).
- Docker and docker compose, if you plan to run in containers.
- nsjail binary and proto definitions under `code-runner/nsjail` (managed via submodules).

Setup:
- `git submodule init`
- `git submodule update --init --recursive`
- Install `protoc` (Protocol Buffers compiler) and Go protobuf plugins as needed.
- Run code generation for nsjail proto:
  - From the repository root: `go generate ./internal/nsjail`
- Build and run locally:
  - `go run ./cmd/runner` (spawns concurrent test matches using `egCode`)
- Or build Docker images:
  - `docker compose up --build`
