package cflocal

// type Cf interface {
// 	HasTask() (bool, error)
// 	HasMultiBuildpack() (bool, error)
// 	Buildpack(file string) error
// 	Cleanup() error

// 	New(fixture string) CfApp
// }

// type CfApp interface {
// 	ConfirmBuildpack(version string) error
// 	RunTask(command string) ([]byte, error)
// 	SetEnv(key, value string)
// 	Push() error
// 	Restart() error
// 	IsRunning(max int) bool
// 	GetBody(path string) (string, error)
// 	Files(path string) ([]string, error)
// 	Destroy() error
// }

type local struct {
	DefaultMemory       string
	DefaultDisk         string
	DefaultStdoutStderr io.Writer
	buildpackPath		string
}

func New(language, memory, disk string, out io.Writer) interfaces.Cf {
	return &local{
		Language: language
		DefaultMemory: memory
		DefaultDisk: disk
		DefaultStdoutStderr: out
	}
}

type app struct {
	Name       string
	Path       string
	Stack      string
	Buildpacks []string
	Memory     string
	Disk       string
	DefaultStdoutStderr io.Writer
	stdout     *bytes.Buffer
	env        map[string]string
	logCmd     *exec.Cmd
}

func (c *local) New(fixture string) interfaces.CfApp {
	return &app{
		Name:       filepath.Base(fixture) + "-" + RandStringRunes(20),
		Path:       fixture,
		Stack:      "",
		Buildpacks: []string{c.buildpackPath},
		Memory:     c.DefaultMemory,
		Disk:       c.DefaultDisk,
		DefaultStdoutStderr: c.DefaultStdoutStderr,
		env:        map[string]string{},
		logCmd:     nil,
	}
}

func (c *local)	HasTask() (bool, error){
	return false, nil
}
func (c *local)	HasMultiBuildpack() (bool, error){
	return true, nil
}
func (c *local)	Buildpack(path string) error{
	c.buildpackPath = path
	return nil
}
func (c *local)	Cleanup() error{
	return nil
}

func (a *app) RunTask(command string) ([]byte, error) {
	return nil, fmt.Errorf("Tasks can not be run on cf local")
}
func (a *app) SetEnv(key, value string) {
	a.env[key] = value
}
func (a *app) Push() error {
	a.stdout = bytes.NewBuffer(nil)

	// FIXME -- needs buildpack etc... CURRENTLY added -e and first -b
	cmd := exec.Command("cf", "local", "stage", a.Name, "-e", "-b", )
	cmd.Stderr = a.DefaultStdoutStderr
	cmd.Stdout := a.stdout
	if err := cmd.Run(); err != nil {
		return err
	}

	a.logCmd = exec.Command("cf", "local", "run", a.Name)
	a.logCmd.Stderr = a.DefaultStdoutStderr
	a.logCmd.Stdout = a.stdout
	if err := a.logCmd.Start(); err != nil {
		return err
	}

	return nil
}
func (a *app) Restart() error {
	return nil
}
func (a *app) IsRunning(max int) bool {
	return nil
}
func (a *app) Stdout() string {
	return a.stdout.String()
}
func (a *app) GetBody(path string) (string, error) {
	return nil, nil
}
func (a *app) Files(path string) ([]string, error) {
	return nil, nil
}
func (a *app) Destroy() error {
	return nil
}
