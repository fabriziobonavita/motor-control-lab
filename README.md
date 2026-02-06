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

Requirements: Go 1.22+ (see `go.mod`).

### Quick build

```bash
go build -o bin/mcl ./cmd/mcl
```

### Using Make

```bash
make build        # Build binary
make help         # Show all available targets
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

Test disturbance rejection by injecting a load disturbance:

```bash
./bin/mcl sim step \
  --target 1000 \
  --duration 10 \
  --disturbance-enabled \
  --disturbance-start 5.0 \
  --disturbance-duration 2.0 \
  --disturbance-magnitude 50.0
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
- `--deadzone` actuator deadzone threshold in volts (default: `0.0`)
- `--disturbance-enabled` enable load disturbance injection (default: `false`)
- `--disturbance-start` disturbance start time in seconds (default: `5.0`)
- `--disturbance-duration` disturbance duration in seconds, 0 means infinite (default: `2.0`)
- `--disturbance-magnitude` disturbance magnitude in RPM/s (default: `50.0`)
- `--out` base output directory (default: `runs`)

## Simulation model (current)

The current simulation is a first-order DC motor speed plant:

- steady-state velocity proportional to applied voltage (gain)
- exponential approach to steady-state (single time constant)
- voltage saturation

This is a baseline model used to validate the experiment harness and controller behavior.

Non idealities that can be simulated

- **Deadzone**: Actuator deadzone threshold that prevents small commands from affecting the system
- **Load disturbances**: Step load disturbances can be injected to test PID disturbance rejection. The disturbance is modeled as RPM/s deceleration applied to the plant dynamics. Use `--disturbance-enabled` to enable, and configure timing and magnitude with the `--disturbance-*` flags.

### Disturbance injection

Load disturbances can be injected to test controller disturbance rejection. The disturbance is applied as a step function:

- **Model**: External load disturbance `d(t)` modeled as RPM/s deceleration in the plant dynamics:
  ```
  dv/dt = (1/tau) * (K*V - v) - d(t)
  ```
- **Configuration**: Use `--disturbance-enabled` along with timing and magnitude flags
- **Example**: Test disturbance rejection with a 50 RPM/s load starting at 5 seconds for 2 seconds:
  ```bash
  ./bin/mcl sim step \
    --target 1000 \
    --duration 10 \
    --disturbance-enabled \
    --disturbance-start 5.0 \
    --disturbance-duration 2.0 \
    --disturbance-magnitude 50.0
  ```
- **Logging**: Disturbance values are recorded in `samples.csv` under the `disturbance_rpm_per_s` column

### Known limitations

Real systems have effects not yet modeled here (intentionally staged):

- static friction
- continuous or time-varying load torque (only step disturbances are supported)
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

## Development

The project includes a Makefile with common development tasks:

```bash
make help          # Show all available targets
make fmt           # Format Go code
make lint          # Run linters (golangci-lint or go vet)
make test          # Run all tests
make build         # Build the binary
make clean         # Clean build artifacts
make ci            # Run CI checks (fmt + test + build)
```

### Project-specific targets

```bash
make sim-step      # Run a step response simulation with default parameters
                   # (outputs to ./artifacts/)

make plot-latest   # Show path to latest run artifacts
```

The `sim-step` target runs a simulation with sensible defaults and writes artifacts to `./artifacts/` (instead of the default `runs/` directory).

## Releases

Releases are created automatically when a git tag is pushed.

### Creating a release

1. **Tag the release:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **GitHub Actions will automatically:**
   - Build binaries for:
     - Linux (amd64)
     - macOS (amd64, arm64)
     - Windows (amd64)
   - Generate SHA256 checksums
   - Create a GitHub Release with attached binaries

### Tag format

Use semantic versioning: `vX.Y.Z` (e.g., `v1.0.0`, `v0.2.1`).

### Local release build

To build release binaries locally (for testing):

```bash
make release
```

This will:
- Check git state is clean
- Read version from current git tag
- Build binaries for all platforms into `dist/`
- Generate `SHA256SUMS` file

**Note:** The `make release` target requires a git tag on the current commit. For actual releases, push the tag and let GitHub Actions handle the build and release creation.

## Roadmap

Near-term:
- Additional disturbance shapes (ramp, sinusoidal, custom profiles)
- Sensor quantization and noise
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
