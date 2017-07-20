package integration_test

import (
	"integration/cutlass"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("deploy an app that shows the directory index", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "directory_index"))
	})

	It("runs", func() {
		Expect(app.Push()).To(Succeed())
		Expect(app.InstanceStates()).To(Equal([]string{"RUNNING"}))

		body, err := app.GetBody("/")
		Expect(err).To(BeNil())
		Expect(body).To(ContainSubstring("find-me-too.html"))
		Expect(body).To(ContainSubstring("find-me.html"))

		body, err = app.GetBody("/subdir")
		Expect(err).To(BeNil())
		Expect(body).To(ContainSubstring("This index file should still load normally when viewing a directory; and not a directory index."))
	})
})
