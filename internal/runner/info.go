package runner

type info struct {
	Name string
}

func newInfo() *info {
	return &info{
		Name: "local-bbp",
	}
}
