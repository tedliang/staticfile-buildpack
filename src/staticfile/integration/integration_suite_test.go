package integration_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var bpDir string
var buildpackVersion string
var UpdateBuildpack bool

func init() {
	flag.StringVar(&buildpackVersion, "version", "", "version to use")
	flag.BoolVar(&cutlass.Cached, "cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "128M", "default disk for pushed apps")
	flag.Parse()
	fmt.Println("cutlass.Cached", cutlass.Cached)
}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func PushAppAndConfirm(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 10*time.Second).Should(Equal([]string{"RUNNING"}))
	Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
}

func findRoot() string {
	file := "VERSION"
	for {
		files, err := filepath.Glob(file)
		Expect(err).To(BeNil())
		if len(files) == 1 {
			file, err = filepath.Abs(filepath.Dir(file))
			Expect(err).To(BeNil())
			return file
		}
		file = filepath.Join("..", file)
	}
}

var _ = BeforeSuite(func() {
	bpDir = findRoot()

	if buildpackVersion == "" {
		data, err := ioutil.ReadFile(filepath.Join(bpDir, "VERSION"))
		Expect(err).NotTo(HaveOccurred())
		buildpackVersion = string(data)
		buildpackVersion = fmt.Sprintf("%s.%s", buildpackVersion, time.Now().Format("20060102150405"))
	}

	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = AfterSuite(func() {
	cutlass.DeleteOrphanedRoutes()
})
