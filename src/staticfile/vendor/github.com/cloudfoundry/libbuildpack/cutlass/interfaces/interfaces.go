package interfaces

type Cf interface {
	HasTask() (bool, error)
	HasMultiBuildpack() (bool, error)
	Buildpack(file string) error
	Cleanup() error

	New(fixture string) (CfApp, error)
}

type CfApp interface {
	RunTask(command string) ([]byte, error)
	SetEnv(key, value string)
	Push() error
	Restart() error
	IsRunning() bool
	Stdout() string
	GetBody(path string) (string, error)
	Files(path string) ([]string, error)
	Destroy() error
}
