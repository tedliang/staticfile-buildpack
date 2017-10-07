package cutlass

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass/interfaces"
	"github.com/cloudfoundry/libbuildpack/packager"
)

type VersionedBuildpackPackage struct {
	Version string
	File    string
}

func FindRoot() (string, error) {
	file := "VERSION"
	for {
		files, err := filepath.Glob(file)
		if err != nil {
			return "", err
		}
		if len(files) == 1 {
			file, err = filepath.Abs(filepath.Dir(file))
			if err != nil {
				return "", err
			}
			return file, nil
		}
		file = filepath.Join("..", file)
	}
}

func PackageUniquelyVersionedBuildpack() (VersionedBuildpackPackage, error) {
	bpDir, err := FindRoot()
	if err != nil {
		return VersionedBuildpackPackage{}, err
	}

	data, err := ioutil.ReadFile(filepath.Join(bpDir, "VERSION"))
	if err != nil {
		return VersionedBuildpackPackage{}, err
	}
	buildpackVersion := string(data)
	buildpackVersion = fmt.Sprintf("%s.%s", buildpackVersion, time.Now().Format("20060102150405"))

	file, err := packager.Package(bpDir, packager.CacheDir, buildpackVersion, Cached)
	if err != nil {
		return VersionedBuildpackPackage{}, err
	}

	var manifest struct {
		Language string `yaml:"language"`
	}
	err = libbuildpack.NewYAML().Load(filepath.Join(bpDir, "manifest.yml"), &manifest)
	if err != nil {
		return VersionedBuildpackPackage{}, err
	}

	err = CreateOrUpdateBuildpack(manifest.Language, file)
	if err != nil {
		return VersionedBuildpackPackage{}, err
	}

	return VersionedBuildpackPackage{
		Version: buildpackVersion,
		File:    file,
	}, nil
}

func ConfirmBuildpack(a interfaces.CfApp, version string) error {
	if !strings.Contains(a.Stdout(), fmt.Sprintf("Buildpack version %s\n", version)) {
		var versionLine string
		for _, line := range strings.Split(a.Stdout(), "\n") {
			if versionLine == "" && strings.Contains(line, " Buildpack version ") {
				versionLine = line
			}
		}
		return fmt.Errorf("Wrong buildpack version. Expected '%s', but this was logged: %s", version, versionLine)
	}
	return nil
}

func SeedRandom() {
	seed := int64(time.Now().Nanosecond() + os.Getpid())
	rand.Seed(seed)
}

func RemovePackagedBuildpack(buildpack VersionedBuildpackPackage) error {
	return os.Remove(buildpack.File)
}
