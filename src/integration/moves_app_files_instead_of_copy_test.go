package integration_test

import (
	"integration/cutlass"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing a static app with dummy file in root", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "public_unspecified"))
		Expect(app.Push()).To(Succeed())
		Expect(app.InstanceStates()).To(Equal([]string{"RUNNING"}))
	})

	It("should only have dummy file in public", func() {
		files, err := app.Files("app")
		Expect(err).To(BeNil())

		Expect(files).To(ContainElement("app/public/dummy_file"))
		Expect(files).ToNot(ContainElement("app/dummy_file"))
	})
})
