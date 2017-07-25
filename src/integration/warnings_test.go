package integration_test

import (
	"integration/cutlass"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("deploy has nginx/conf directory", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "fixtures", "nginx_conf"))
		Expect(app.Push()).To(Succeed())
		Expect(app.InstanceStates()).To(Equal([]string{"RUNNING"}))
	})

	It("warns user to set root", func() {
		Expect(app.Stdout).To(ContainSubstring("You have an nginx/conf directory, but have not set *root*."))
		Expect(app.GetBody("/")).To(ContainSubstring("Test warnings"))
	})
})
