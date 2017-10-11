package models

type Cf interface {
	HasTask() (bool, error)
	HasMultiBuildpack() (bool, error)
	Buildpack(file string) error
	Cleanup() error

	New(fixture string) (CfApp, error)
}

type CfApp interface {
	Name() string
	SetEnv(key, value string)
	Buildpacks(paths ...string)
	Push() error
	Restart() error
	IsRunning() bool
	Stdout() string
	GetUrl(path string) (string, error)
	Files(path string) ([]string, error)
	RunTask(command string) ([]byte, error)
	Destroy() error
}
