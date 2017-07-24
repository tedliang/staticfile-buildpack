package main

import (
	"flag"
	"fmt"
	"github.com/cloudfoundry/libbuildpack"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Manifest struct {
	Dependencies []struct {
		Name string `yaml:"name"`
		Url  string `yaml:"url"`
		MD5  string `yaml:"md5"`
	} `yaml:"dependencies"`
}

func main() {
	cached := flag.Bool("cached", false, "build a cached buildpack")

	manifest := Manifest{}
	err := libbuildpack.NewYAML().Load("manifest.yml", manifest)
	checkErr(err)

	dir, err := copyBuildpack("~/workspace/staticfile-buildpack")
	checkErr(err)

	if *cached {
		cacheDir = filepath.Join(os.ENV["HOME"], ".buildpack-packager", "cache")
		err := cacheDependencies(dir, cacheDir, &manifest)
		checkErr(err)
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func copyBuildpack(srcDir string) (string, error) {
	destDir, err := ioutil.TempDir("", "cutlass-fixture-copy")
	if err != nil {
		return "", err
	}
	if err := libbuildpack.CopyDirectory(srcDir, destDir); err != nil {
		return "", err
	}
	return destDir, nil
}

func cacheDependencies(dir, cacheDir string, manifest *Manifest) error {
	for _, d := range manifest.Dependencies {
		fmt.Println(d)
	}
	return nil
}

func downloadFromUrl(dir, url string) error {
	err := os.MkdirAll(filepath.Join(dir, "dependencies"), 0755)
	if err != nil {
		return err
	}

	r := strings.NewReplacer("/", "_", ":", "_", "?", "_", "&", "_")
	fileName := filepath.Join(dir, "dependencies", r.Replace(url))
	fmt.Println("Downloading", url, "to", fileName)

	output, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		return err
	}

	return nil
}
