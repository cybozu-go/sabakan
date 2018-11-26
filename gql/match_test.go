package gql

import (
	"testing"
	"time"

	"github.com/cybozu-go/sabakan"
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
		having    *MachineParams
		notHaving *MachineParams
		now       time.Time
		expect    bool
	}{
		{
			name: "trivial",
			machine: &sabakan.Machine{
				Spec:   sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{},
			},
			having:    &MachineParams{},
			notHaving: &MachineParams{},
			now:       now,
			expect:    true,
		},
		{
			name: "label-not-found",
			machine: &sabakan.Machine{
				Spec:   sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{},
			},
			having: &MachineParams{
				Labels: []LabelInput{{"foo", "bar"}},
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{
				Labels: []LabelInput{{"foo", "bar"}},
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{
				Labels: []LabelInput{{"foo", "bar"}},
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{
				Labels: []LabelInput{{"foo", "bar"}},
			},
			notHaving: &MachineParams{
				Labels: []LabelInput{{"foo3", "bar3"}},
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
			having: &MachineParams{
				Labels: []LabelInput{{"foo", "bar"}},
			},
			notHaving: &MachineParams{
				Labels: []LabelInput{{"foo2", "bar2"}},
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
			having: &MachineParams{
				Racks: []int{0, 2},
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{
				Racks: []int{0, 2},
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{},
			notHaving: &MachineParams{
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
			having: &MachineParams{
				Roles: []string{"boot"},
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{
				Roles: []string{"boot", "worker"},
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{},
			notHaving: &MachineParams{
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
			having: &MachineParams{
				States: []sabakan.MachineState{sabakan.StateUninitialized},
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{
				States: []sabakan.MachineState{sabakan.StateUninitialized, sabakan.StateHealthy},
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{},
			notHaving: &MachineParams{
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
			having: &MachineParams{
				MinDaysBeforeRetire: testInt(90),
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{
				MinDaysBeforeRetire: testInt(50),
			},
			notHaving: &MachineParams{},
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
			having: &MachineParams{},
			notHaving: &MachineParams{
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
			notHaving: &MachineParams{},
			now:       now,
			expect:    true,
		},
		{
			name: "nil-nothaving",
			machine: &sabakan.Machine{
				Spec:   sabakan.MachineSpec{},
				Status: sabakan.MachineStatus{},
			},
			having: &MachineParams{},
			now:    now,
			expect: true,
		},
	}

	for _, c := range testCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			if matchMachine(c.machine, c.having, c.notHaving, c.now) != c.expect {
				t.Error("unexpected result")
			}
		})
	}
}
