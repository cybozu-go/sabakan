package itest

import (
	"bytes"
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("assets", func() {

	It("should be distributed between servers", func() {
		f := localTempFile("test")
		defer os.Remove(f.Name())

		sabactl("assets", "upload", "test", f.Name())

		Eventually(func() bool {
			var info struct {
				Urls []string `json:"urls"`
			}

			stdout := sabactl("assets", "info", "test")
			err := json.Unmarshal(stdout, &info)
			if err != nil {
				return false
			}
			return len(info.Urls) == 3
		}).Should(BeTrue())

		sabactl("assets", "delete", "test")
		stdout := sabactl("assets", "index")
		var assets []string
		err := json.NewDecoder(bytes.NewReader(stdout)).Decode(&assets)
		if err != nil {
			Fail(err.Error())
		}
		Expect(assets).NotTo(ContainElement("test"))
	})
})
