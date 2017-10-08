package cflocal

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/cloudfoundry/libbuildpack/cutlass/models"
)

type local struct {
	Language            string
	DefaultMemory       string
	DefaultDisk         string
	DefaultStdoutStderr io.Writer
	buildpackPath       string
}

var _ = models.Cf(&local{})

func New(language, buildpackPath, memory, disk string, out io.Writer) models.Cf {
	return &local{
		Language:            language,
		DefaultMemory:       memory,
		DefaultDisk:         disk,
		DefaultStdoutStderr: out,
		buildpackPath:       buildpackPath,
	}
}

type app struct {
	Name                string
	Path                string
	Stack               string
	Buildpacks          []string
	Memory              string
	Disk                string
	DefaultStdoutStderr io.Writer
	stdout              *bytes.Buffer
	env                 map[string]string
	logCmd              *exec.Cmd
	tmpDir              string
	port                string
}

var _ = models.CfApp(&app{})

func (c *local) New(fixture string) (models.CfApp, error) {
	tmpDir, err := ioutil.TempDir("", "cutlass.cflocal.")
	if err != nil {
		return nil, err
	}
	return &app{
		Name:                filepath.Base(fixture) + "-" + cutlass.RandStringRunes(20),
		Path:                fixture,
		Stack:               "",
		Buildpacks:          []string{c.buildpackPath},
		Memory:              c.DefaultMemory,
		Disk:                c.DefaultDisk,
		DefaultStdoutStderr: c.DefaultStdoutStderr,
		env:                 map[string]string{},
		logCmd:              nil,
		tmpDir:              tmpDir,
	}, nil
}

func (c *local) HasTask() (bool, error) {
	return false, nil
}
func (c *local) HasMultiBuildpack() (bool, error) {
	return true, nil
}
func (c *local) Buildpack(path string) error {
	c.buildpackPath = path
	return nil
}
func (c *local) Cleanup() error {
	return nil
}

func (a *app) RunTask(command string) ([]byte, error) {
	return nil, fmt.Errorf("Tasks can not be run on cf local")
}
func (a *app) SetEnv(key, value string) {
	a.env[key] = value
}
func (a *app) generateLocalYML() error {
	app := struct {
		Name       string
		Buildpacks []string `json:"buildpacks"`
		// Command    string            `json:"command"`
		Memory string            `json:"memory"`
		Disk   string            `json:"disk_quota"`
		ENV    map[string]string `json:"env"`
	}{
		a.Name,
		a.Buildpacks,
		// "",
		a.Memory,
		a.Disk,
		a.env,
	}
	cfg := map[string][]interface{}{"applications": []interface{}{app}}
	fmt.Println("TmpDir: ", a.tmpDir)
	return libbuildpack.NewYAML().Write(filepath.Join(a.tmpDir, "local.yml"), cfg)
}
func (a *app) Push() error {
	a.stdout = bytes.NewBuffer(nil)
	if err := a.generateLocalYML(); err != nil {
		return err
	}

	// FIXME -- Should add "-e" for default buildpack case
	cmd := exec.Command("cf", "local", "stage", a.Name, "-p", a.Path)
	cmd.Dir = a.tmpDir
	cmd.Stderr = a.DefaultStdoutStderr
	cmd.Stdout = a.stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	return a.start()
}
func (a *app) start() error {
	var buf bytes.Buffer
	w := io.MultiWriter(&buf, a.stdout)

	a.logCmd = exec.Command("cf", "local", "run", a.Name)
	a.logCmd.Dir = a.tmpDir
	a.logCmd.Stderr = a.DefaultStdoutStderr
	a.logCmd.Stdout = w
	if err := a.logCmd.Start(); err != nil {
		return err
	}

	timeout := time.After(5 * time.Second)
	tick := time.Tick(100 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("Could not determine start port")
		case <-tick:
			if line, err := buf.ReadString("\n"[0]); err == nil {
				fmt.Println(line, err)
				if m := regexp.MustCompile(`on port (\d+)`).FindStringSubmatch(line); len(m) == 2 {
					a.port = m[1]
					return nil
				}
			}
		}
	}

	return fmt.Errorf("Could not determine start port")
}
func (a *app) stop() error {
	if a.logCmd != nil {
		if err := a.logCmd.Process.Kill(); err != nil {
			return err
		}
		a.logCmd = nil
	}
	return nil
}
func (a *app) Restart() error {
	if err := a.stop(); err != nil {
		return err
	}
	return a.start()
}
func (a *app) IsRunning() bool {
	_, err := net.Dial("tcp", "localhost:"+a.port)
	return err == nil
}
func (a *app) Stdout() string {
	return a.stdout.String()
}
func (a *app) GetUrl(path string) (string, error) {
	return fmt.Sprintf("http://localhost:%s%s", a.port, path), nil
}
func (a *app) Files(path string) ([]string, error) {
	return nil, nil
}
func (a *app) Destroy() error {
	if err := a.stop(); err != nil {
		return err
	}
	return os.RemoveAll(a.tmpDir)
}
