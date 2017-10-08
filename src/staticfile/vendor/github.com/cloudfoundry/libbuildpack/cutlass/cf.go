package cutlass

import (
	"io"

	"github.com/cloudfoundry/libbuildpack/cutlass/models"
)

var Cf models.Cf
var DefaultStdoutStderr io.Writer
var DefaultMemory string
var DefaultDisk string
var Cached bool

type App struct {
	app models.CfApp
}

func New(fixture string) *App {
	app, err := Cf.New(fixture)
	if err != nil {
		panic(err)
	}
	return &App{
		app: app,
	}
}

func DeleteOrphanedRoutes() error {
	return Cf.Cleanup()
}

func DeleteBuildpack(language string) error {
	return Cf.Cleanup()
}

func UpdateBuildpack(language, file string) error {
	return CreateOrUpdateBuildpack(language, file)
}

func CreateOrUpdateBuildpack(language, file string) error {
	// FIXME
	// if language != Cf.Language() {
	// 	return fmt.Errorf("Language does not match: %s != %s", language, Cf.Language)
	// }
	return Cf.Buildpack(file)
}

func (a *App) ConfirmBuildpack(version string) error {
	return ConfirmBuildpack(a.app, version)
}

func (a *App) RunTask(command string) ([]byte, error) {
	return a.app.RunTask(command)
}

func (a *App) Restart() error {
	return a.app.Restart()
}

func (a *App) IsRunning() bool {
	return a.app.IsRunning()
}

func (a *App) SetEnv(key, value string) {
	a.app.SetEnv(key, value)
}

func (a *App) Push() error {
	return a.app.Push()
}

func (a *App) GetUrl(path string) (string, error) {
	return a.app.GetUrl(path)
}

func (a *App) Get(path string, headers map[string]string) (string, map[string][]string, error) {
	return Get(a.app, path, headers)
}

func (a *App) GetBody(path string) (string, error) {
	return GetBody(a.app, path)
}

func (a *App) Files(path string) ([]string, error) {
	return a.app.Files(path)
}

func (a *App) Destroy() error {
	return a.app.Destroy()
}
