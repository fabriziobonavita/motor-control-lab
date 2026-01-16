package plotting

import (
	"path/filepath"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"github.com/fabriziobonavita/motor-control-lab/internal/experiment"
)

func WriteVelocityPlot(runDir string, samples []experiment.Sample) error {
	if len(samples) == 0 {
		return nil
	}

	p := plot.New()
	p.Title.Text = "Velocity Response"
	p.X.Label.Text = "Time (s)"
	p.Y.Label.Text = "Velocity (RPM)"
	p.Legend.Top = true

	// Create plotter for actual velocity
	actualPoints := make(plotter.XYs, len(samples))
	for i, s := range samples {
		actualPoints[i].X = s.T
		actualPoints[i].Y = s.Actual
	}
	actualLine, err := plotter.NewLine(actualPoints)
	if err != nil {
		return err
	}
	actualLine.Color = plotutil.Color(0)
	actualLine.Width = vg.Points(1.5)
	p.Add(actualLine)
	p.Legend.Add("Actual", actualLine)

	// Create plotter for target velocity
	targetPoints := make(plotter.XYs, len(samples))
	for i, s := range samples {
		targetPoints[i].X = s.T
		targetPoints[i].Y = s.Target
	}
	targetLine, err := plotter.NewLine(targetPoints)
	if err != nil {
		return err
	}
	targetLine.Color = plotutil.Color(1)
	targetLine.Width = vg.Points(1.5)
	targetLine.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
	p.Add(targetLine)
	p.Legend.Add("Target", targetLine)

	// Save the plot
	if err := p.Save(8*vg.Inch, 4*vg.Inch, filepath.Join(runDir, "velocity.png")); err != nil {
		return err
	}

	return nil
}

func WriteControlPlot(runDir string, samples []experiment.Sample) error {
	if len(samples) == 0 {
		return nil
	}

	p := plot.New()
	p.Title.Text = "Control Signal"
	p.X.Label.Text = "Time (s)"
	p.Y.Label.Text = "Voltage (V)"
	p.Legend.Top = true

	// Create plotter for control signal
	controlPoints := make(plotter.XYs, len(samples))
	for i, s := range samples {
		controlPoints[i].X = s.T
		controlPoints[i].Y = s.U
	}
	controlLine, err := plotter.NewLine(controlPoints)
	if err != nil {
		return err
	}
	controlLine.Color = plotutil.Color(2)
	controlLine.Width = vg.Points(1.5)
	p.Add(controlLine)
	p.Legend.Add("Control (U)", controlLine)

	// Save the plot
	if err := p.Save(8*vg.Inch, 4*vg.Inch, filepath.Join(runDir, "control.png")); err != nil {
		return err
	}

	return nil
}
