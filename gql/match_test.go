package gql

import (
	"testing"
	"time"

	"github.com/cybozu-go/sabakan/v3"
	"github.com/cybozu-go/sabakan/v3/gql/graph/model"
)

func testInt(i int) *int {
	return &i
}

func TestMatchMachine(t *testing.T) {
	now := time.Date(2018, time.November, 26, 0, 0, 0, 0, time.UTC)
	nowPlus60 := now.Add(time.Hour * 24 * 60)

	testCases := []struct {
		name      string
		machine   *sabakan.Machine
		having    *model.MachineParams
		notHaving *model.MachineParams
		now       time.Time
		expect    bool
	}{
		{
			name: "trivial",
			machine: &sabakan.Machine{
				Spec:   sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{},
			},
			having:    &model.MachineParams{},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    true,
		},
		{
			name: "label-not-found",
			machine: &sabakan.Machine{
				Spec:   sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				Labels: []*model.LabelInput{{Name: "foo", Value: "bar"}},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    false,
		},
		{
			name: "label-data-mismatch",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Labels: map[string]string{
						"foo": "zot",
					},
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				Labels: []*model.LabelInput{{Name: "foo", Value: "bar"}},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    false,
		},
		{
			name: "label-match",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Labels: map[string]string{
						"foo":  "bar",
						"foo2": "bar2",
					},
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				Labels: []*model.LabelInput{{Name: "foo", Value: "bar"}},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    true,
		},
		{
			name: "label-match2",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Labels: map[string]string{
						"foo":  "bar",
						"foo2": "bar2",
					},
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				Labels: []*model.LabelInput{{Name: "foo", Value: "bar"}},
			},
			notHaving: &model.MachineParams{
				Labels: []*model.LabelInput{{Name: "foo3", Value: "bar3"}},
			},
			now:    now,
			expect: true,
		},
		{
			name: "label-found",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Labels: map[string]string{
						"foo":  "bar",
						"foo2": "bar2",
					},
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				Labels: []*model.LabelInput{{Name: "foo", Value: "bar"}},
			},
			notHaving: &model.MachineParams{
				Labels: []*model.LabelInput{{Name: "foo2", Value: "bar2"}},
			},
			now:    now,
			expect: false,
		},
		{
			name: "rack-mismatch",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Rack: 1,
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				Racks: []int{0, 2},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    false,
		},
		{
			name: "rack-match",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Rack: 2,
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				Racks: []int{0, 2},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    true,
		},
		{
			name: "rack-found",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Rack: 2,
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{},
			notHaving: &model.MachineParams{
				Racks: []int{0, 2},
			},
			now:    now,
			expect: false,
		},
		{
			name: "role-mismatch",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Role: "worker",
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				Roles: []string{"boot"},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    false,
		},
		{
			name: "role-match",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Role: "worker",
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				Roles: []string{"boot", "worker"},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    true,
		},
		{
			name: "role-found",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					Role: "worker",
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{},
			notHaving: &model.MachineParams{
				Roles: []string{"boot", "worker"},
			},
			now:    now,
			expect: false,
		},
		{
			name: "state-mismatch",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{
					State: sabakan.StateHealthy,
				},
			},
			having: &model.MachineParams{
				States: []sabakan.MachineState{sabakan.StateUninitialized},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    false,
		},
		{
			name: "state-match",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{
					State: sabakan.StateHealthy,
				},
			},
			having: &model.MachineParams{
				States: []sabakan.MachineState{sabakan.StateUninitialized, sabakan.StateHealthy},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    true,
		},
		{
			name: "state-found",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{
					State: sabakan.StateHealthy,
				},
			},
			having: &model.MachineParams{},
			notHaving: &model.MachineParams{
				States: []sabakan.MachineState{sabakan.StateHealthy},
			},
			now:    now,
			expect: false,
		},
		{
			name: "days-short",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					RetireDate: nowPlus60,
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				MinDaysBeforeRetire: testInt(90),
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    false,
		},
		{
			name: "days-match",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					RetireDate: nowPlus60,
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{
				MinDaysBeforeRetire: testInt(50),
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    true,
		},
		{
			name: "days-not-having",
			machine: &sabakan.Machine{
				Spec: sabakan.MachineSpec{
					RetireDate: nowPlus60,
				},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{},
			notHaving: &model.MachineParams{
				MinDaysBeforeRetire: testInt(50),
			},
			now:    now,
			expect: false,
		},
		{
			name: "nil-having",
			machine: &sabakan.Machine{
				Spec:   sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{},
			},
			notHaving: &model.MachineParams{},
			now:       now,
			expect:    true,
		},
		{
			name: "nil-nothaving",
			machine: &sabakan.Machine{
				Spec:   sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{},
			},
			having: &model.MachineParams{},
			now:    now,
			expect: true,
		},
	}

	for _, c := range testCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if MatchMachine(c.machine, c.having, c.notHaving, c.now) != c.expect {
				t.Error("unexpected result")
			}
		})
	}
}
