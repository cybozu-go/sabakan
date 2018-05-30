package web

import "testing"

func testHandleCoreOSKernel(t *testing.T) {
	t.Parallel()
}

func testHandleCoreOSInitRD(t *testing.T) {
	t.Parallel()
}

func TestHandleCoreOS(t *testing.T) {
	t.Run("kernel", testHandleCoreOSKernel)
	t.Run("initrd", testHandleCoreOSInitRD)
}
