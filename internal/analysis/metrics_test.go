package analysis

import (
	"math"
	"testing"

	"github.com/fabriziobonavita/motor-control-lab/internal/experiment"
)

const eps = 1e-6

func TestOvershootPercent(t *testing.T) {
	tests := []struct {
		name    string
		target  float64
		max     float64
		want    float64
		samples []experiment.Sample
	}{
		{
			name:   "10% overshoot",
			target: 100.0,
			max:    110.0,
			want:   10.0,
			samples: makeSamples(100.0, []float64{0, 50, 100, 110, 105, 100}, 0.1),
		},
		{
			name:   "no overshoot",
			target: 100.0,
			max:    95.0,
			want:   0.0,
			samples: makeSamples(100.0, []float64{0, 50, 95, 100}, 0.1),
		},
		{
			name:   "zero target",
			target: 0.0,
			max:    10.0,
			want:   0.0, // overshoot is 0 when target is 0
			samples: makeSamples(0.0, []float64{0, 5, 10, 0}, 0.1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := Compute(tt.samples, 0.02)
			if math.Abs(metrics.OvershootPercent-tt.want) > eps {
				t.Errorf("OvershootPercent = %v, want %v", metrics.OvershootPercent, tt.want)
			}
		})
	}
}

func TestSettlingTime(t *testing.T) {
	tests := []struct {
		name    string
		target  float64
		samples []experiment.Sample
		want    float64
		wantNaN bool
	}{
		{
			name:   "settles at t=2.0",
			target: 100.0,
			samples: func() []experiment.Sample {
				// Error > band until t=2.0, then within band
				samples := make([]experiment.Sample, 0, 50)
				dt := 0.1
				for i := 0; i < 50; i++ {
					t := float64(i) * dt
					actual := 0.0
					if t >= 2.0 {
						actual = 100.0 // within 2% band
					} else {
						actual = 50.0 // outside band
					}
					samples = append(samples, experiment.Sample{
						T:      t,
						DT:     dt,
						Target: 100.0,
						Actual: actual,
						Error:  100.0 - actual,
					})
				}
				return samples
			}(),
			want:    2.0,
			wantNaN: false,
		},
		{
			name:   "never settles",
			target: 100.0,
			samples: func() []experiment.Sample {
				// Error always outside band
				samples := make([]experiment.Sample, 0, 20)
				dt := 0.1
				for i := 0; i < 20; i++ {
					t := float64(i) * dt
					samples = append(samples, experiment.Sample{
						T:      t,
						DT:     dt,
						Target: 100.0,
						Actual: 50.0, // always 50% error
						Error:  50.0,
					})
				}
				return samples
			}(),
			want:    math.NaN(),
			wantNaN: true,
		},
		{
			name:   "settles immediately",
			target: 100.0,
			samples: func() []experiment.Sample {
				// Error within band from start
				samples := make([]experiment.Sample, 0, 10)
				dt := 0.1
				for i := 0; i < 10; i++ {
					t := float64(i) * dt
					samples = append(samples, experiment.Sample{
						T:      t,
						DT:     dt,
						Target: 100.0,
						Actual: 100.0, // exactly on target
						Error:  0.0,
					})
				}
				return samples
			}(),
			want:    0.0,
			wantNaN: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := Compute(tt.samples, 0.02)
			if tt.wantNaN {
				if !math.IsNaN(metrics.SettlingTimeSeconds) {
					t.Errorf("SettlingTimeSeconds = %v, want NaN", metrics.SettlingTimeSeconds)
				}
			} else {
				if math.IsNaN(metrics.SettlingTimeSeconds) {
					t.Errorf("SettlingTimeSeconds = NaN, want %v", tt.want)
				} else if math.Abs(metrics.SettlingTimeSeconds-tt.want) > 0.1 {
					// Allow some tolerance for settling time
					t.Errorf("SettlingTimeSeconds = %v, want %v", metrics.SettlingTimeSeconds, tt.want)
				}
			}
		})
	}
}

func TestIAE(t *testing.T) {
	tests := []struct {
		name    string
		samples []experiment.Sample
		want    float64
	}{
		{
			name: "constant error",
			// error=1.0 for 10 steps at dt=0.1 => IAE = 1.0 * 0.1 * 10 = 1.0
			samples: func() []experiment.Sample {
				samples := make([]experiment.Sample, 0, 10)
				dt := 0.1
				for i := 0; i < 10; i++ {
					t := float64(i) * dt
					samples = append(samples, experiment.Sample{
						T:      t,
						DT:     dt,
						Target: 100.0,
						Actual: 99.0, // error = 1.0
						Error:  1.0,
					})
				}
				return samples
			}(),
			want: 1.0,
		},
		{
			name: "zero error",
			samples: func() []experiment.Sample {
				samples := make([]experiment.Sample, 0, 5)
				dt := 0.1
				for i := 0; i < 5; i++ {
					t := float64(i) * dt
					samples = append(samples, experiment.Sample{
						T:      t,
						DT:     dt,
						Target: 100.0,
						Actual: 100.0,
						Error:  0.0,
					})
				}
				return samples
			}(),
			want: 0.0,
		},
		{
			name: "varying error",
			// errors: [1, 2, 0.5] with dt=0.1 => IAE = (1+2+0.5)*0.1 = 0.35
			samples: []experiment.Sample{
				{T: 0.0, DT: 0.1, Target: 100.0, Actual: 99.0, Error: 1.0},
				{T: 0.1, DT: 0.1, Target: 100.0, Actual: 98.0, Error: 2.0},
				{T: 0.2, DT: 0.1, Target: 100.0, Actual: 99.5, Error: 0.5},
			},
			want: 0.35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := Compute(tt.samples, 0.02)
			if math.Abs(metrics.IAE-tt.want) > eps {
				t.Errorf("IAE = %v, want %v", metrics.IAE, tt.want)
			}
		})
	}
}

func TestSaturationFraction(t *testing.T) {
	tests := []struct {
		name    string
		samples []experiment.Sample
		want    float64
	}{
		{
			name: "half saturated",
			samples: func() []experiment.Sample {
				samples := make([]experiment.Sample, 0, 10)
				dt := 0.1
				for i := 0; i < 10; i++ {
					t := float64(i) * dt
					samples = append(samples, experiment.Sample{
						T:         t,
						DT:        dt,
						Target:    100.0,
						Actual:    50.0,
						Error:     50.0,
						Saturated: i < 5, // first 5 are saturated
					})
				}
				return samples
			}(),
			want: 0.5, // 5/10
		},
		{
			name: "all saturated",
			samples: func() []experiment.Sample {
				samples := make([]experiment.Sample, 0, 5)
				dt := 0.1
				for i := 0; i < 5; i++ {
					t := float64(i) * dt
					samples = append(samples, experiment.Sample{
						T:         t,
						DT:        dt,
						Target:    100.0,
						Actual:    50.0,
						Error:     50.0,
						Saturated: true,
					})
				}
				return samples
			}(),
			want: 1.0,
		},
		{
			name: "none saturated",
			samples: func() []experiment.Sample {
				samples := make([]experiment.Sample, 0, 5)
				dt := 0.1
				for i := 0; i < 5; i++ {
					t := float64(i) * dt
					samples = append(samples, experiment.Sample{
						T:         t,
						DT:        dt,
						Target:    100.0,
						Actual:    50.0,
						Error:     50.0,
						Saturated: false,
					})
				}
				return samples
			}(),
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := Compute(tt.samples, 0.02)
			if math.Abs(metrics.SaturationFraction-tt.want) > eps {
				t.Errorf("SaturationFraction = %v, want %v", metrics.SaturationFraction, tt.want)
			}
		})
	}
}

func TestEmptySamples(t *testing.T) {
	metrics := Compute(nil, 0.02)
	if !math.IsNaN(metrics.SettlingTimeSeconds) {
		t.Errorf("SettlingTimeSeconds for empty samples = %v, want NaN", metrics.SettlingTimeSeconds)
	}
}

// makeSamples creates a slice of samples with given target and actual values
func makeSamples(target float64, actuals []float64, dt float64) []experiment.Sample {
	samples := make([]experiment.Sample, 0, len(actuals))
	for i, actual := range actuals {
		t := float64(i) * dt
		samples = append(samples, experiment.Sample{
			T:      t,
			DT:     dt,
			Target: target,
			Actual: actual,
			Error:  target - actual,
		})
	}
	return samples
}
