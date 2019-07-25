package main

import (
	"fmt"
	"github.com/cybozu-go/log"
	"github.com/google/go-tpm/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

const tpmOffsetHex = 0x01000000

var tpmOffset = tpmutil.Handle(tpmOffsetHex)

func main() {
	rw, err := tpm2.OpenTPM("/dev/tpm0")
	if err != nil {
		log.ErrorExit(err)
	}
	ek, err := tpm2.NVReadEx(rw, tpmOffset, tpm2.HandleOwner, "", 0)
	if err != nil {
		log.ErrorExit(err)
	}
	fmt.Printf("%x", ek)
}
