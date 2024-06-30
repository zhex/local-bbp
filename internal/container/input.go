package container

type Input struct {
	Name    string
	Image   string
	WorkDir string
	HostDir string
	Envs    map[string]string
}
