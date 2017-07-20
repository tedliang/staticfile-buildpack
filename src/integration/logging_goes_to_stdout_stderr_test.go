package integration_test

import (
	"integration/cutlass"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("nginx logs go to stdout and stderr", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		app = cutlass.New(filepath.Join(bpDir, "cf_spec", "fixtures", "staticfile_app"))
		Expect(app.Push()).To(Succeed())
		Expect(app.InstanceStates()).To(Equal([]string{"RUNNING"}))
	})

	It("", func() {
		By("writes regular logs to stdout and does not write to actual log files", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("This is an example app for Cloud Foundry that is only static HTML/JS/CSS assets."))
			Eventually(app.Stdout).Should(MatchRegexp("OUT.*GET / HTTP/1.1"))
			command := exec.Command("cf", "ssh", app.Name, "-c", "ls -l /app/nginx/logs/ | grep access.log")
			Expect(command.Output()).To(ContainSubstring(" vcap 0 "))
		})

		By("writes error logs to stderr and does not write to actual log files", func() {
			Expect(app.GetBody("/idontexist")).To(ContainSubstring("404 Not Found"))
			Eventually(app.Stdout).Should(MatchRegexp("ERR.*GET /idontexist HTTP/1.1"))
			command := exec.Command("cf", "ssh", app.Name, "-c", "ls -l /app/nginx/logs/ | grep error.log")
			Expect(command.Output()).To(ContainSubstring(" vcap 0 "))
		})
	})
})
