package integration_test

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/cloudfoundry/libbuildpack/packager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var bpDir string
var buildpackVersion string
var UpdateBuildpack bool

func init() {
	flag.BoolVar(&cutlass.Cached, "cached", true, "cached buildpack")
	flag.BoolVar(&UpdateBuildpack, "update-buildpack", true, "build buildpack and update to buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "256M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
	flag.Parse()
	fmt.Println("cutlass.Cached", cutlass.Cached)
}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func PushAppAndConfirm(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Expect(app.InstanceStates()).To(Equal([]string{"RUNNING"}))
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
	data, err := ioutil.ReadFile(filepath.Join(bpDir, "VERSION"))
	Expect(err).NotTo(HaveOccurred())
	buildpackVersion = string(data)

	if UpdateBuildpack {
		buildpackVersion = fmt.Sprintf("%s.%s", buildpackVersion, time.Now().Format("20060102150405"))
		file, err := packager.Package(bpDir, packager.CacheDir, buildpackVersion, cutlass.Cached)
		Expect(err).To(BeNil())

		var manifest struct {
			Language string `yaml:"language"`
		}
		Expect(libbuildpack.NewYAML().Load(filepath.Join(bpDir, "manifest.yml"), &manifest)).To(Succeed())
		Expect(cutlass.UpdateBuildpack(manifest.Language, file)).To(Succeed())

		os.Remove(file)
	}
})

var _ = AfterSuite(func() {
	cutlass.DeleteOrphanedRoutes()
})
