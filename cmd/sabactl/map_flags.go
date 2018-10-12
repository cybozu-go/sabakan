package main

import (
	"errors"
	"fmt"
	"strings"
)

type mapFlags map[string]string

func (i *mapFlags) String() string {
	return fmt.Sprint(*i)
}

func (i *mapFlags) Set(value string) error {
	kv := strings.SplitN(value, "=", 2)
	if len(kv) != 2 {
		return errors.New("invalid options value: " + value)
	}
	(*i)[kv[0]] = kv[1]
	return nil
}
