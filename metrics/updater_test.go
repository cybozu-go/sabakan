package metrics

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cybozu-go/sabakan/v2"
	"github.com/cybozu-go/sabakan/v2/models/mock"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type gaugeStatus struct {
	expectedLabels map[string]string
	expectedValue  float64
	otherwise      float64
}

type gaugeTestCase struct {
	name           string
	input          func() (*sabakan.Model, error)
	expectedMetric gaugeStatus
}

func testMachineStatus(t *testing.T) {
	testCases := []gaugeTestCase{
		{
			name:  "all machines is uninitialized",
			input: twoMachinesWithUninitialized,
			expectedMetric: gaugeStatus{
				map[string]string{"status": sabakan.StateUninitialized.String()}, 1, 0,
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			model, err := tt.input()
			if err != nil {
				t.Error(err)
			}
			updater := NewUpdater(10*time.Millisecond, model)

			ctx := context.Background()
			defer ctx.Done()

			go updater.UpdateLoop(ctx)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			GetHandler().ServeHTTP(w, req)
			metricsFamily, err := parseMetrics(w.Result())
			if err != nil {
				t.Error(err)
			}
			for _, m := range metricsFamily {
				switch *m.Name {
				case "sabakan_machine_status":
					for _, m := range m.Metric {
						lm := labelToMap(m.Label)
						if hasLabels(lm, tt.expectedMetric.expectedLabels) {
							if *m.Gauge.Value != tt.expectedMetric.expectedValue {
								t.Error("not uninitialized")
							}
						} else if *m.Gauge.Value != tt.expectedMetric.otherwise {
							t.Errorf("%v has unexpected value; %f", m, tt.expectedMetric.otherwise)
						}
					}
				}
			}
		})
	}
}

func testAssetMetrics(t *testing.T) {
	fmt.Println("hoge1")

	testCases := []gaugeTestCase{
		{
			name:  "get total size of assets",
			input: twoAssets,
			expectedMetric: gaugeStatus{
				nil, 6, 0,
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			model, err := tt.input()
			if err != nil {
				t.Error(err)
			}
			updater := NewUpdater(10*time.Millisecond, model)

			ctx := context.Background()
			defer ctx.Done()

			go updater.UpdateLoop(ctx)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			GetHandler().ServeHTTP(w, req)
			metricsFamily, err := parseMetrics(w.Result())
			if err != nil {
				t.Error(err)
			}
			for _, m := range metricsFamily {
				switch *m.Name {
				case "sabakan_assets_bytes_total":
					for _, m := range m.Metric {
						if *m.Gauge.Value != tt.expectedMetric.expectedValue {
							t.Error("does not record the correct size")
						}
					}
				}
			}
		})
	}
}

func twoMachinesWithUninitialized() (*sabakan.Model, error) {
	model := mock.NewModel()
	var machines []*sabakan.Machine
	for i := 0; i < 2; i++ {
		machines = append(machines, &sabakan.Machine{
			Spec:   sabakan.MachineSpec{Serial: "001", Rack: 1, IndexInRack: 1, IPv4: []string{"10.0.0.1"}},
			Status: sabakan.MachineStatus{State: sabakan.StateUninitialized}})
	}
	err := model.Machine.Register(context.Background(), machines)
	return &model, err
}

func twoAssets() (*sabakan.Model, error) {
	model := mock.NewModel()

	for i := 0; i < 2; i++ {
		_, err := model.Asset.Put(context.Background(), "asset"+strconv.Itoa(i), "ctype"+strconv.Itoa(i), nil, nil, strings.NewReader("bar"))
		if err != nil {
			return nil, err
		}
	}

	return &model, nil
}

func labelToMap(labelPair []*dto.LabelPair) map[string]string {
	res := make(map[string]string)
	for _, l := range labelPair {
		res[*l.Name] = *l.Value
	}
	return res
}

func hasLabels(lm map[string]string, expectedLabels map[string]string) bool {
	for ek, ev := range expectedLabels {
		found := false
		for k, v := range lm {
			if k == ek && v == ev {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func parseMetrics(resp *http.Response) ([]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	parsed, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, err
	}
	var result []*dto.MetricFamily
	for _, mf := range parsed {
		result = append(result, mf)
	}
	return result, nil
}

func TestMetrics(t *testing.T) {
	t.Run("machine status", testMachineStatus)
	t.Run("asset metrics", testAssetMetrics)
}
