package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/cloudfoundry/libbuildpack/cutlass/cflocal"
	"github.com/cloudfoundry/libbuildpack/cutlass/models"
)

func main() {
	bpPath, _ := filepath.Abs("../../staticfile_buildpack-cached-v1.4.16.zip")
	// bpPath := "https://github.com/cloudfoundry/staticfile-buildpack/releases/download/v1.4.16/staticfile-buildpack-v1.4.16.zip"
	// fixPath, _ := filepath.Abs("../../fixtures/staticfile_app")
	fixPath, _ := filepath.Abs("../../fixtures/include_headers")
	cf := cflocal.New("staticfile", bpPath, "1Gb", "1Gb", os.Stdout)
	app, err := cf.New(fixPath)
	if err != nil {
		panic(err)
	}
	app.SetEnv("BP_DEBUG", "1")
	if err := app.Push(); err != nil {
		fmt.Println(app.Stdout())
		panic(err)
	}

	waitForRunning(app)

	if body, err := cutlass.GetBody(app, ""); err != nil {
		panic(err)
	} else {
		fmt.Println(body)
	}

	time.Sleep(1 * time.Second)
	fmt.Println(app.Stdout())

	if err := app.Destroy(); err != nil {
		panic(err)
	}
}

func waitForRunning(app models.CfApp) {
	timeout := time.After(5 * time.Second)
	tick := time.Tick(100 * time.Millisecond)
Loop:
	for {
		select {
		case <-timeout:
			panic("not running")
		case <-tick:
			if app.IsRunning() {
				break Loop
			}
		}
	}
}
