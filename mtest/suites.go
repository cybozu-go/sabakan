package mtest

import . "github.com/onsi/ginkgo"

// FunctionsSuite is a test suite that tests all test cases
var FunctionsSuite = func() {
	Context("assets", TestAssets)
	Context("netboot", TestNetboot)
}
