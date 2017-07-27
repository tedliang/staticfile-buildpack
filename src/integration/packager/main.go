package main

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"integration/cutlass"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type Manifest struct {
	Language     string   `yaml:"language"`
	IncludeFiles []string `yaml:"include_files"`
	PrePackage   string   `yaml:"pre_package"`
	Dependencies []struct {
		URI string `yaml:"uri"`
		MD5 string `yaml:"md5"`
	} `yaml:"dependencies"`
}

type File struct {
	Name, Path string
}

func main() {
	var cached bool
	var cacheDir, version string
	flag.BoolVar(&cached, "cached", false, "include dependencies")
	flag.StringVar(&cacheDir, "cachedir", filepath.Join(os.Getenv("HOME"), ".buildpack-packager", "cache"), "cache dir")
	flag.StringVar(&version, "version", "", "version")
	flag.Parse()

	if version == "" {
		v, err := ioutil.ReadFile("VERSION")
		if err != nil {
			log.Fatalf("error: Could not read VERSION file: %v", err)
		}
		version = string(v)
	}

	dir, err := cutlass.CopyFixture(".")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(dir, "VERSION"), []byte(version), 0644)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	manifest := Manifest{}
	data, err := ioutil.ReadFile(filepath.Join(dir, "manifest.yml"))
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		log.Fatalf("error: %v", err)
	}

	if manifest.PrePackage != "" {
		cmd := exec.Command(manifest.PrePackage)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(string(out))
			log.Fatalf("error: %v", err)
		}
	}

	files := []File{}
	for _, name := range manifest.IncludeFiles {
		files = append(files, File{name, filepath.Join(dir, name)})
	}

	if cached {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			log.Fatalf("error: %v", err)
		}
		r := strings.NewReplacer("/", "_", ":", "_", "?", "_", "&", "_")
		for _, d := range manifest.Dependencies {
			// TODO filteredUri make user/pass safe
			// TODO set file value instead and make libbuildpack use it
			name := r.Replace(d.URI)
			dest := filepath.Join(cacheDir, name)
			if _, err := os.Stat(dest); err != nil {
				if os.IsNotExist(err) {
					err = downloadFromUrl(d.URI, dest)
				}
				if err != nil {
					log.Fatalf("error: %v", err)
				}
			}
			files = append(files, File{filepath.Join("dependencies", name), dest})
		}
	}

	zipFile := fmt.Sprintf("%s_buildpack-v%s.zip", manifest.Language, version)
	buildpackType := "uncached"
	if cached {
		zipFile = fmt.Sprintf("%s_buildpack-cached-v%s.zip", manifest.Language, version)
		buildpackType = "cached"
	}

	ZipFiles(zipFile, files)

	stat, err := os.Stat(zipFile)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("%s buildpack created and saved as %s with a size of %dMB\n", buildpackType, zipFile, stat.Size()/1024/1024)
}

func downloadFromUrl(url, fileName string) error {
	// TODO: check file existence first with io.IsExist
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

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("could not download: %d", response.StatusCode)
	}

	if _, err := io.Copy(output, response.Body); err != nil {
		return err
	}
	return nil
}

func checkMD5(filePath, expectedMD5 string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	hashInBytes := hash.Sum(nil)[:16]
	actualMD5 := hex.EncodeToString(hashInBytes)

	if actualMD5 != expectedMD5 {
		return fmt.Errorf("dependency md5 mismatch: expected md5 %s, actual md5 %s", expectedMD5, actualMD5)
	}
	return nil
}

func ZipFiles(filename string, files []File) error {
	newfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newfile.Close()

	zipWriter := zip.NewWriter(newfile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {

		zipfile, err := os.Open(file.Path)
		if err != nil {
			return err
		}
		defer zipfile.Close()

		// Get the file information
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate
		header.Name = file.Name

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, zipfile)
		if err != nil {
			return err
		}
	}
	return nil
}
