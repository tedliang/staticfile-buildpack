package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("deploy an app that shows the directory index", func() {
	var app *cutlass.App
	// AfterEach(func() {
	// 	if app != nil {
	// 		app.Destroy()
	// 	}
	// 	app = nil
	// })

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "directory_index"))
	})

	It("runs", func() {
		PushAppAndConfirm(app)

		body, err := app.GetBody("/")
		Expect(err).To(BeNil())
		Expect(body).To(ContainSubstring("find-me-too.html"))
		Expect(body).To(ContainSubstring("find-me.html"))

		Expect(app.GetBody("/subdir")).To(ContainSubstring("This index file should still load normally when viewing a directory; and not a directory index."))
	})
})
