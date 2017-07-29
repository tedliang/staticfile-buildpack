package integration_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var bpDir string
var buildpackVersion string
var UpdateBuildpack bool

func init() {
	flag.BoolVar(&cutlass.Cached, "cached", false, "cached buildpack")
	flag.BoolVar(&UpdateBuildpack, "update-buildpack", true, "build buildpack and update to buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "256M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "memory", "256M", "default disk for pushed apps")
}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
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
	data, err := ioutil.ReadFile(filepath.Join(bpDir, "VERSION"))
	Expect(err).NotTo(HaveOccurred())
	buildpackVersion = string(data)

	fmt.Printf("UpdateBuildpack:", UpdateBuildpack)
	panic("hi")
})

var _ = AfterSuite(func() {
	cutlass.DeleteOrphanedRoutes()
})
