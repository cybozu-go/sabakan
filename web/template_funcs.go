package web

import (
	"encoding/json"
	"errors"
)

var (
	errNotInt       = errors.New("not an integer")
	errNotFloat     = errors.New("not a float")
	errZeroDivision = errors.New("zero division")
)

func jsonFunc(i interface{}) (string, error) {
	data, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getAsInt64(a interface{}) (int64, error) {
	switch i := a.(type) {
	case int:
		return int64(i), nil
	case int8:
		return int64(i), nil
	case int16:
		return int64(i), nil
	case int32:
		return int64(i), nil
	case int64:
		return int64(i), nil
	case uint:
		return int64(i), nil
	case uint8:
		return int64(i), nil
	case uint16:
		return int64(i), nil
	case uint32:
		return int64(i), nil
	case uint64:
		return int64(i), nil
	}

	return 0, errNotInt
}

func getAsFloat64(a interface{}) (float64, error) {
	switch f := a.(type) {
	case float32:
		return float64(f), nil
	case float64:
		return f, nil
	}

	return 0, errNotFloat
}

func getAsInts(a, b interface{}) (int64, int64, error) {
	ia, err := getAsInt64(a)
	if err != nil {
		return 0, 0, err
	}
	ib, err := getAsInt64(b)
	if err != nil {
		return 0, 0, err
	}

	return ia, ib, nil
}

func getAsFloats(a, b interface{}) (float64, float64, error) {
	fa, err := getAsFloat64(a)
	if err != nil {
		ia, err2 := getAsInt64(a)
		if err2 != nil {
			return 0, 0, err
		}
		fa = float64(ia)
	}
	fb, err := getAsFloat64(b)
	if err != nil {
		ib, err2 := getAsInt64(b)
		if err2 != nil {
			return 0, 0, err
		}
		fb = float64(ib)
	}

	return fa, fb, nil
}

func addFunc(a, b interface{}) (interface{}, error) {
	ia, ib, err := getAsInts(a, b)
	if err == nil {
		return ia + ib, nil
	}
	fa, fb, err := getAsFloats(a, b)
	if err == nil {
		return fa + fb, nil
	}
	return nil, err
}

func subFunc(a, b interface{}) (interface{}, error) {
	ia, ib, err := getAsInts(a, b)
	if err == nil {
		return ia - ib, nil
	}
	fa, fb, err := getAsFloats(a, b)
	if err == nil {
		return fa - fb, nil
	}
	return nil, err
}

func mulFunc(a, b interface{}) (interface{}, error) {
	ia, ib, err := getAsInts(a, b)
	if err == nil {
		return ia * ib, nil
	}
	fa, fb, err := getAsFloats(a, b)
	if err == nil {
		return fa * fb, nil
	}
	return nil, err
}

func divFunc(a, b interface{}) (interface{}, error) {
	ia, ib, err := getAsInts(a, b)
	if err == nil {
		if ib == 0 {
			return nil, errZeroDivision
		}
		return ia / ib, nil
	}
	fa, fb, err := getAsFloats(a, b)
	if err == nil {
		if fb == 0 {
			return nil, errZeroDivision
		}
		return fa / fb, nil
	}
	return nil, err
}
