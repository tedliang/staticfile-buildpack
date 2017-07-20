package integration_test

import (
	"integration/cutlass"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("deploy includes headers", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "include_headers"))
		Expect(app.Push()).To(Succeed())
		Expect(app.InstanceStates()).To(Equal([]string{"RUNNING"}))
	})

	It("adds headers", func() {
		body, headers, err := app.Get("/", map[string]string{})
		Expect(err).To(BeNil())
		Expect(body).To(ContainSubstring("Test add headers"))
		Expect(headers).To(HaveKey("X-Superspecial"))
	})
})
