package config

type AboutSoftware struct {
	Version    string   `json:"version" yaml:"version"`
	Leader     string   `json:"leader" yaml:"leader"`
	Developers []string `json:"developers" yaml:"developers"`
}
