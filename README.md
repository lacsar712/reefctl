# reefctl

Reefer cold-chain unit controller coordinating compressor staging, evaporator defrost cycles, duct airflow setpoints, cargo-zone routing, and damper interlocks.

## Build

```bash
make build
make test
```

## Run

```bash
go run ./cmd/reefctl
```

## Packages

- `internal/reefer` — unit orchestration, ducts, zones, dampers, airflow
- `internal/compressor` — compressor coordinator and staging
- `internal/defrost` — defrost cycle windows and temperature sensors
- `internal/fsm` — reefer and compressor state machines
- `internal/clock` — process clock for simulation ticks
- `internal/interlock` — damper valve locks
- `internal/store` — in-memory snapshots and schedules
- `internal/app` — application wiring and CLI runner
