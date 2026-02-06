package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/fabriziobonavita/motor-control-lab/internal/analysis"
	"github.com/fabriziobonavita/motor-control-lab/internal/artifacts"
	"github.com/fabriziobonavita/motor-control-lab/internal/control/pid"
	"github.com/fabriziobonavita/motor-control-lab/internal/experiment"
	"github.com/fabriziobonavita/motor-control-lab/internal/experiment/modifier"
	"github.com/fabriziobonavita/motor-control-lab/internal/plotting"
	"github.com/fabriziobonavita/motor-control-lab/internal/system"
	"github.com/fabriziobonavita/motor-control-lab/internal/system/sim"
	"github.com/fabriziobonavita/motor-control-lab/internal/system/wrap"
)

var (
	kp                 float64
	ki                 float64
	kd                 float64
	target             float64
	duration           float64
	dt                 float64
	deadzone           float64
	disturbanceEnabled bool
	disturbanceStart   float64
	disturbanceDur     float64
	disturbanceMag     float64
	outBase            string
)

func newSimStepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "step",
		Short: "Run a step response simulation",
		Long:  "Run a step response simulation with PID control on a DC motor.",
		RunE:  runSimStep,
	}

	cmd.Flags().Float64Var(&kp, "kp", 0.02, "proportional gain")
	cmd.Flags().Float64Var(&ki, "ki", 0.05, "integral gain")
	cmd.Flags().Float64Var(&kd, "kd", 0.0, "derivative gain")
	cmd.Flags().Float64Var(&target, "target", 1000.0, "target velocity (RPM)")
	cmd.Flags().Float64Var(&duration, "duration", 10.0, "simulation duration (s)")
	cmd.Flags().Float64Var(&dt, "dt", 0.001, "simulation timestep (s)")
	cmd.Flags().Float64Var(&deadzone, "deadzone", 0.0, "actuator deadzone threshold (V)")
	cmd.Flags().BoolVar(&disturbanceEnabled, "disturbance-enabled", false, "enable load disturbance injection")
	cmd.Flags().Float64Var(&disturbanceStart, "disturbance-start", 5.0, "disturbance start time (s)")
	cmd.Flags().Float64Var(&disturbanceDur, "disturbance-duration", 2.0, "disturbance duration (s, 0 = infinite)")
	cmd.Flags().Float64Var(&disturbanceMag, "disturbance-magnitude", 50.0, "disturbance magnitude (RPM/s)")
	cmd.Flags().StringVar(&outBase, "out", "runs", "base output directory")

	return cmd
}

func runSimStep(cmd *cobra.Command, args []string) error {
	ctrl := pid.New(kp, ki, kd)
	plant := sim.NewDCMotor()

	// Wrap plant with DisturbedSystem if disturbance is enabled
	var sys system.System = plant
	if disturbanceEnabled {
		disturbanceCfg := wrap.StepDisturbanceConfig{
			Enabled:          disturbanceEnabled,
			StartS:           disturbanceStart,
			DurationS:        disturbanceDur,
			MagnitudeRPMPerS: disturbanceMag,
		}
		sys = wrap.NewDisturbedSystem(plant, disturbanceCfg)
	}

	var mod modifier.Modifier
	if deadzone > 0 {
		mod = modifier.Chain(&modifier.DeadzoneModifier{Threshold: deadzone})
	}

	cfg := experiment.StepConfig{
		TargetRPM: target,
		DT:        dt,
		Duration:  duration,
		Modifier:  mod,
	}
	samples, wall := experiment.RunStep(sys, ctrl, cfg)
	if len(samples) == 0 {
		return fmt.Errorf("no samples produced")
	}

	params := map[string]any{
		"kp":                              kp,
		"ki":                              ki,
		"kd":                              kd,
		"target_rpm":                      target,
		"duration_s":                      duration,
		"dt_s":                            dt,
		"deadzone_v":                      deadzone,
		"disturbance_enabled":             disturbanceEnabled,
		"disturbance_start_s":             disturbanceStart,
		"disturbance_duration_s":          disturbanceDur,
		"disturbance_magnitude_rpm_per_s": disturbanceMag,
	}

	run, md, err := artifacts.Create(outBase, "sim", "dc-motor", "step", params)
	if err != nil {
		return err
	}
	defer func() {
		if err := run.Close(); err != nil {
			// Log error but don't fail - cleanup operation
			fmt.Fprintf(os.Stderr, "warning: failed to close run directory: %v\n", err)
		}
	}()

	// samples.csv
	if err := run.WriteSamplesCSV(samples); err != nil {
		return err
	}

	// metrics.json
	metrics := analysis.Compute(samples, 0.02)
	if err := artifacts.WriteJSON(filepath.Join(run.Dir, "metrics.json"), metrics); err != nil {
		return err
	}

	// plots
	if err := plotting.WriteVelocityPlot(run.Dir, samples); err != nil {
		return err
	}
	if err := plotting.WriteControlPlot(run.Dir, samples); err != nil {
		return err
	}

	// out.log (human-oriented summary)
	last := samples[len(samples)-1]
	_, _ = fmt.Fprintf(run.Out(), "run_id=%s\n", md.RunID)
	_, _ = fmt.Fprintf(run.Out(), "wall_time=%s\n", wall)
	_, _ = fmt.Fprintf(run.Out(), "final_actual=%.3f\n", last.Actual)
	_, _ = fmt.Fprintf(run.Out(), "final_error=%.3f\n", last.Error)
	_, _ = fmt.Fprintf(run.Out(), "overshoot_percent=%.3f\n", metrics.OvershootPercent)
	_, _ = fmt.Fprintf(run.Out(), "settling_time_seconds=%v\n", metrics.SettlingTimeSeconds)
	_, _ = fmt.Fprintf(run.Out(), "iae=%.6f\n", metrics.IAE)

	// console output
	fmt.Println("Run:", md.RunID)
	fmt.Println("Artifacts:", run.Dir)
	fmt.Printf("Final: actual=%.2fRPM err=%.2f u=%.2fV\n", last.Actual, last.Error, last.U)
	fmt.Printf("Metrics: overshoot=%.2f%% settling=%v iae=%.3f\n", metrics.OvershootPercent, metrics.SettlingTimeSeconds, metrics.IAE)

	return nil
}
