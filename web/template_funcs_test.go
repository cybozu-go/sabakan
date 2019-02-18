package web

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func testArithmeticFuncs(t *testing.T) {
	cases := []struct {
		name    string
		f       func(a, b interface{}) (interface{}, error)
		a       interface{}
		b       interface{}
		expect  interface{}
		wantErr bool
	}{
		{"addInt2", addFunc, int(3), int(-1), int64(2), false},
		{"addInt8Int16", addFunc, int8(3), int16(-1), int64(2), false},
		{"addInt32Int64", addFunc, int32(3), int64(-1), int64(2), false},
		{"addUintUint8", addFunc, uint(3), uint8(1), int64(4), false},
		{"addUint16Uint32", addFunc, uint16(3), uint32(1), int64(4), false},
		{"addUint32Uint64", addFunc, uint32(3), uint64(1), int64(4), false},
		{"addIntFloat32", addFunc, int(3), float32(-1.3), float64(1.7), false},
		{"addFloat64Uint", addFunc, float64(3.3), uint(1), float64(4.3), false},
		{"addFloat32Float64", addFunc, float32(3.3), float64(1), float64(4.3), false},
		{"addInt64String", addFunc, int64(-1), "", nil, true},
		{"addFloat64String", addFunc, float64(1.7), "", nil, true},
		{"subInt2", subFunc, int(3), int(-1), int64(4), false},
		{"subInt8Int16", subFunc, int8(3), int16(-1), int64(4), false},
		{"subInt32Int64", subFunc, int32(3), int64(-1), int64(4), false},
		{"subUintUint8", subFunc, uint(3), uint8(1), int64(2), false},
		{"subUint16Uint32", subFunc, uint16(3), uint32(1), int64(2), false},
		{"subUint32Uint64", subFunc, uint32(3), uint64(1), int64(2), false},
		{"subIntFloat32", subFunc, int(3), float32(-1.3), float64(4.3), false},
		{"subFloat64Uint", subFunc, float64(3.3), uint(1), float64(2.3), false},
		{"subInt64String", subFunc, int64(-1), "", nil, true},
		{"subFloat64String", subFunc, float64(1.7), "", nil, true},
		{"mulInt2", mulFunc, int(3), int(-1), int64(-3), false},
		{"mulInt8Int16", mulFunc, int8(3), int16(-1), int64(-3), false},
		{"mulInt32Int64", mulFunc, int32(3), int64(-1), int64(-3), false},
		{"mulUintUint8", mulFunc, uint(3), uint8(1), int64(3), false},
		{"mulUint16Uint32", mulFunc, uint16(3), uint32(1), int64(3), false},
		{"mulUint32Uint64", mulFunc, uint32(3), uint64(1), int64(3), false},
		{"mulIntFloat32", mulFunc, int(3), float32(-1.3), float64(-3.9), false},
		{"mulFloat64Uint", mulFunc, float64(3.3), uint(1), float64(3.3), false},
		{"mulInt64String", mulFunc, int64(-1), "", nil, true},
		{"mulFloat64String", mulFunc, float64(1.7), "", nil, true},
		{"divInt2", divFunc, int(4), int(2), int64(2), false},
		{"divInt8Int16", divFunc, int8(4), int16(2), int64(2), false},
		{"divInt32Int64", divFunc, int32(4), int64(2), int64(2), false},
		{"divUintUint8", divFunc, uint(4), uint8(2), int64(2), false},
		{"divUint16Uint32", divFunc, uint16(4), uint32(2), int64(2), false},
		{"divUint32Uint64", divFunc, uint32(4), uint64(2), int64(2), false},
		{"divFloat32Int", divFunc, float32(3.9), int(3), float64(1.3), false},
		{"divFloat64Uint", divFunc, float64(3.9), uint(3), float64(1.3), false},
		{"divisionByIntZero", divFunc, int(10), int64(0), nil, true},
		{"divisionByFloat64Zero", divFunc, int(10), float64(0), nil, true},
		{"divInt64String", divFunc, int64(-1), "", nil, true},
		{"divFloat64String", divFunc, float64(1.7), "", nil, true},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			actual, err := tt.f(tt.a, tt.b)
			if err != nil {
				if !tt.wantErr {
					t.Error("unexpected error:", err)
				}
				return
			}

			if !cmp.Equal(tt.expect, actual, testCmp) {
				t.Error("unexpected result:", cmp.Diff(tt.expect, actual, testCmp))
			}
		})
	}
}

var testCmp = cmp.Comparer(func(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
})

func TestTemplateFuncs(t *testing.T) {
	t.Run("arithmetic", testArithmeticFuncs)
}
