# Motor Control Lab (`mcl`)

Motor Control Lab is a small, experiment-driven control laboratory for developing, analyzing, and tuning low-level motion controllers.

The core idea:

> The same experiment runner, metrics, and artifacts should work for simulation today and real hardware later.

This repository is designed to grow into a practical toolkit for tuning toy trains, rovers, drones, robots, and (eventually) swarms, without turning into a one-off script or a PID tutorial.

## Features

- CLI tool (`mcl`) built with Cobra
- Deterministic simulation runner (fixed timestep)
- Structured run artifacts per run directory:
  - `samples.csv` (time series)
  - `metadata.json` (configuration + environment)
  - `metrics.json` (objective evaluation)
  - `out.log` (human-readable summary)
  - `velocity.png`, `control.png` (plots)
- Clear separation between:
  - controller
  - system/plant
  - experiments
  - analysis (metrics)
  - artifact generation

## Install / Build

Requirements: Go 1.21+ recommended.

Build:

```bash
go build -o bin/mcl ./cmd/mcl
```

Run help:

```bash
./bin/mcl --help
```

## Quickstart

Run a step response simulation:

```bash
./bin/mcl sim step \
  --target 1000 \
  --duration 10 \
  --dt 0.001 \
  --kp 0.02 \
  --ki 0.05 \
  --kd 0.0 \
  --deadzone 0.0
```

Artifacts are written to `--out` (default: `runs`). A run directory looks like:

```
runs/2026-01-16T09-05-29Z_sim_dc-motor_step/
  metadata.json
  samples.csv
  metrics.json
  out.log
  velocity.png
  control.png
```

## Commands

### `mcl`

Motor Control Lab CLI.

### `mcl sim`

Run simulations.

### `mcl sim step`

Run a closed-loop step response simulation with PID control on a simulated DC motor.

Flags:
- `--kp` proportional gain (default: `0.02`)
- `--ki` integral gain (default: `0.05`)
- `--kd` derivative gain (default: `0.0`)
- `--target` target velocity in RPM (default: `1000`)
- `--duration` simulation duration in seconds (default: `10`)
- `--dt` simulation timestep in seconds (default: `0.001`)
- `--deadzone` add a deadzone to the system (default: `0.0`)
- `--out` base output directory (default: `runs`)

## Simulation model (current)

The current simulation is a first-order DC motor speed plant:

- steady-state velocity proportional to applied voltage (gain)
- exponential approach to steady-state (single time constant)
- voltage saturation

This is a baseline model used to validate the experiment harness and controller behavior.

Non idealities that can be simulated

- deadzone adding a threshold that won't lead to changes

### Known limitations

Real systems have effects not yet modeled here (intentionally staged):

- static friction
- load torque disturbances (terrain, slope, payload)
- supply sag (battery voltage drop under load)
- encoder quantization and noise
- drivetrain backlash or slip

## Metrics

Each run computes objective metrics and writes them to `metrics.json`:

- overshoot (percent)
- settling time (within a band, currently +/-2%)
- steady-state error
- IAE (Integral of Absolute Error)
- saturation fraction

These metrics are designed to support automated comparison and future autotuning.

## Repository structure (high level)

```text
cmd/mcl/                CLI entry point and commands
internal/control/       Controllers (PID)
internal/system/        Simulated plants and future hardware adapters
internal/experiment/    Experiment runners (e.g., step response)
internal/analysis/      Metrics and evaluation
internal/artifacts/     Run directories and file outputs
internal/plotting/      Plot generation
runs/                   Generated run artifacts (gitignored)
```

## Roadmap

Near-term:
- Add simulation non-idealities (disturbances, sensor quantization)
- Batch runs / parameter sweeps
- Baseline autotune (constrained search over gains using metrics)

Hardware:
- Hardware system adapters (PWM + encoder)
- Run identical experiments on real motors

Later:
- System identification pipeline (open-loop tests -> model fitting -> initial gains)
- Cascaded loops (position -> velocity)
- Fault injection (sensor dropouts, delays)
- Multi-agent / swarm experiments using the same runner + metrics backbone
