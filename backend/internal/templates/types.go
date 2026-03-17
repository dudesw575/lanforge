package templates

type Template struct {
	ID          string            `yaml:"id" json:"id"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Image       string            `yaml:"image" json:"image"`
	Env         map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Ports       []Port            `yaml:"ports,omitempty" json:"ports,omitempty"`
	Volumes     []Volume          `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Restart     string            `yaml:"restart,omitempty" json:"restart,omitempty"`
}

type Port struct {
	Host      int `yaml:"host" json:"host"`
	Container int `yaml:"container" json:"container"`
}

type Volume struct {
	Name      string `yaml:"name" json:"name"`
	Container string `yaml:"container" json:"container"`
}
