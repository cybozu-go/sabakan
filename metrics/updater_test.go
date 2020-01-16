package metrics

import (
	"context"
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

type expectedMachineStatus struct {
	status string
	labels map[string]string
}

type machineStatusTestCase struct {
	name            string
	input           func() (*sabakan.Model, error)
	expectedMetrics []expectedMachineStatus
}

type assetsTestCase struct {
	name          string
	input         func() (*sabakan.Model, error)
	expectedName  string
	expectedValue float64
}

func testMachineStatus(t *testing.T) {
	testCases := []machineStatusTestCase{
		{
			name:  "1 uninitialized, 1 healthy",
			input: twoMachines,
			expectedMetrics: []expectedMachineStatus{
				{
					status: sabakan.StateUninitialized.String(),
					labels: map[string]string{
						"address":      "10.0.0.1",
						"serial":       "001",
						"rack":         "1",
						"role":         "cs",
						"machine-type": "cray-1",
					},
				},
				{
					status: sabakan.StateHealthy.String(),
					labels: map[string]string{
						"address":      "10.0.0.2",
						"serial":       "002",
						"rack":         "2",
						"role":         "ss",
						"machine-type": "cray-2",
					},
				},
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
			updater.UpdateAllMetrics(ctx)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			GetHandler().ServeHTTP(w, req)
			metricsFamily, err := parseMetrics(w.Result())
			if err != nil {
				t.Error(err)
			}
			for _, mf := range metricsFamily {
				switch *mf.Name {
				case "sabakan_machine_status":
					for _, em := range tt.expectedMetrics {
						states := make(map[string]bool)
						for _, m := range mf.Metric {
							lm := labelToMap(m.Label)
							if hasLabels(lm, em.labels) {
								states[lm["status"]] = true
								if lm["status"] == em.status {
									if *m.Gauge.Value != 1 {
										t.Errorf("value of expected status %q of %q must be 1 but %f", lm["status"], em.labels["serial"], *m.Gauge.Value)
									}
								} else {
									if *m.Gauge.Value != 0 {
										t.Errorf("value of unexpected status %q of %q must be 0 but %f", lm["status"], em.labels["serial"], *m.Gauge.Value)
									}
								}
							}
						}
						for _, s := range sabakan.StateList {
							if !states[s.String()] {
								t.Errorf("metrics for %q was not found", em.labels["serial"])
							}
						}
					}
				}
			}
		})
	}
}

func testAssetsMetrics(t *testing.T) {
	testCases := []assetsTestCase{
		{
			name:          "get total size of assets",
			input:         twoAssets,
			expectedName:  "sabakan_assets_bytes_total",
			expectedValue: 6,
		},
		{
			name:          "get total item numbers of assets",
			input:         twoAssets,
			expectedName:  "sabakan_assets_items_total",
			expectedValue: 2,
		},
		{
			name:          "get total size of images",
			input:         threeImages,
			expectedName:  "sabakan_images_bytes_total",
			expectedValue: 18,
		},
		{
			name:          "get total item numbers of images",
			input:         threeImages,
			expectedName:  "sabakan_images_items_total",
			expectedValue: 3,
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
			updater.UpdateAllMetrics(ctx)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/metrics", nil)
			GetHandler().ServeHTTP(w, req)
			metricsFamily, err := parseMetrics(w.Result())
			if err != nil {
				t.Error(err)
			}

			found := false
			for _, mf := range metricsFamily {
				if *mf.Name == tt.expectedName {
					found = true
					for _, m := range mf.Metric {
						if *m.Gauge.Value != tt.expectedValue {
							t.Errorf("value for %q is wrong.  expected: %f, actual: %f", tt.expectedName, tt.expectedValue, *m.Gauge.Value)
						}
					}
				}
			}
			if !found {
				t.Errorf("metrics %q was not found", tt.expectedName)
			}
		})
	}
}

func twoMachines() (*sabakan.Model, error) {
	model := mock.NewModel()
	machines := []*sabakan.Machine{
		{
			Spec: sabakan.MachineSpec{
				Serial: "001",
				Rack:   1,
				Role:   "cs",
				Labels: map[string]string{"machine-type": "cray-1"},
				IPv4:   []string{"10.0.0.1"},
			},
			Status: sabakan.MachineStatus{
				State: sabakan.StateUninitialized,
			},
		},
		{
			Spec: sabakan.MachineSpec{
				Serial: "002",
				Rack:   2,
				Role:   "ss",
				Labels: map[string]string{"machine-type": "cray-2"},
				IPv4:   []string{"10.0.0.2"},
			},
			Status: sabakan.MachineStatus{
				State: sabakan.StateHealthy,
			},
		},
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

func threeImages() (*sabakan.Model, error) {
	model := mock.NewModel()

	for i := 0; i < 3; i++ {
		err := model.Image.Upload(context.Background(), "ubuntu", "image"+strconv.Itoa(i), strings.NewReader("foobar"))
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
	t.Run("MachineStatus", testMachineStatus)
	t.Run("AssetsImagesMetrics", testAssetsMetrics)
}
