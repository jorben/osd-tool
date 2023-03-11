package config

import "gopkg.in/yaml.v3"

// TransferConfig 同步配置
type TransferConfig struct {
	Storage string `yaml:"storage"`
	Upload  struct {
		List   []Path   `yaml:"list"`
		Ignore []string `yaml:"ignore"`
	} `yaml:"upload"`
	Download struct {
		List   []Path   `yaml:"list"`
		Ignore []string `yaml:"ignore,omitempty"`
	} `yaml:"download"`
	Osd struct {
		SecretId  string `yaml:"secret_id"`
		SecretKey string `yaml:"secret_key"`
		Bucket    string `yaml:"bucket"`
		Region    string `yaml:"region"`
		Timeout   int    `yaml:"timeout"`
	} `yaml:"osd"`
}

type Path struct {
	Source string `yaml:"source"`
	Dest   string `yaml:"dest"`
}

func GetConfigDemo() []byte {
	buf, _ := yaml.Marshal(TransferConfig{
		Storage: "",
		Upload: struct {
			List   []Path   `yaml:"list"`
			Ignore []string `yaml:"ignore"`
		}{
			List: []Path{
				{
					Source: "",
					Dest:   "",
				},
			},
			Ignore: []string{".git", ".idea"},
		},
		Download: struct {
			List   []Path   `yaml:"list"`
			Ignore []string `yaml:"ignore,omitempty"`
		}{
			List: []Path{
				{
					Source: "",
					Dest:   "",
				},
			},
		},
		Osd: struct {
			SecretId  string `yaml:"secret_id"`
			SecretKey string `yaml:"secret_key"`
			Bucket    string `yaml:"bucket"`
			Region    string `yaml:"region"`
			Timeout   int    `yaml:"timeout"`
		}{},
	})
	return buf
}
