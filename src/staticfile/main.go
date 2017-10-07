package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass/cflocal"
	"github.com/cloudfoundry/libbuildpack/cutlass/interfaces"
)

func main() {
	bpPath := "/home/dgodd/workspace/staticfile-buildpack/staticfile_buildpack-cached-v1.4.16.zip"
	cf := cflocal.New("staticfile", bpPath, "1Gb", "1Gb", os.Stdout)
	app, err := cf.New("../../fixtures/staticfile_app")
	if err != nil {
		panic(err)
	}
	if err := app.Push(); err != nil {
		fmt.Println(app.Stdout())
		panic(err)
	}

	waitForRunning(app)
	time.Sleep(1 * time.Second)

	if body, err := app.GetBody("/"); err != nil {
		panic(err)
	} else {
		fmt.Println(body)
	}

	fmt.Println(app.Stdout())
}

func waitForRunning(app interfaces.CfApp) {
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
