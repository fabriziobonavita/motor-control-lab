package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/fabriziobonavita/motor-control-lab/internal/analysis"
	"github.com/fabriziobonavita/motor-control-lab/internal/artifacts"
	"github.com/fabriziobonavita/motor-control-lab/internal/control/pid"
	"github.com/fabriziobonavita/motor-control-lab/internal/experiment"
	"github.com/fabriziobonavita/motor-control-lab/internal/experiment/modifier"
	"github.com/fabriziobonavita/motor-control-lab/internal/plotting"
	"github.com/fabriziobonavita/motor-control-lab/internal/system/sim"
)

var (
	kp       float64
	ki       float64
	kd       float64
	target   float64
	duration float64
	dt       float64
	deadzone float64
	outBase  string
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
	cmd.Flags().StringVar(&outBase, "out", "runs", "base output directory")

	return cmd
}

func runSimStep(cmd *cobra.Command, args []string) error {
	ctrl := pid.New(kp, ki, kd)
	plant := sim.NewDCMotor()

	var mod modifier.Modifier
	if deadzone > 0 {
		mod = modifier.Chain(&modifier.DeadzoneModifier{Threshold: deadzone})
	}

	cfg := experiment.StepConfig{TargetRPM: target, DT: dt, Duration: duration, Modifier: mod}
	samples, wall := experiment.RunStep(plant, ctrl, cfg)
	if len(samples) == 0 {
		return fmt.Errorf("no samples produced")
	}

	params := map[string]any{
		"kp":         kp,
		"ki":         ki,
		"kd":         kd,
		"target_rpm": target,
		"duration_s": duration,
		"dt_s":       dt,
		"deadzone_v": deadzone,
	}

	run, md, err := artifacts.Create(outBase, "sim", "dc-motor", "step", params)
	if err != nil {
		return err
	}
	defer run.Close()

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
	fmt.Fprintf(run.Out(), "run_id=%s\n", md.RunID)
	fmt.Fprintf(run.Out(), "wall_time=%s\n", wall)
	fmt.Fprintf(run.Out(), "final_actual=%.3f\n", last.Actual)
	fmt.Fprintf(run.Out(), "final_error=%.3f\n", last.Error)
	fmt.Fprintf(run.Out(), "overshoot_percent=%.3f\n", metrics.OvershootPercent)
	fmt.Fprintf(run.Out(), "settling_time_seconds=%v\n", metrics.SettlingTimeSeconds)
	fmt.Fprintf(run.Out(), "iae=%.6f\n", metrics.IAE)

	// console output
	fmt.Println("Run:", md.RunID)
	fmt.Println("Artifacts:", run.Dir)
	fmt.Printf("Final: actual=%.2fRPM err=%.2f u=%.2fV\n", last.Actual, last.Error, last.U)
	fmt.Printf("Metrics: overshoot=%.2f%% settling=%v iae=%.3f\n", metrics.OvershootPercent, metrics.SettlingTimeSeconds, metrics.IAE)

	return nil
}
