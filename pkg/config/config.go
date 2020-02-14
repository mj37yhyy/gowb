package config

type Config struct {
	Log   Log   `mapstructure:"log"`
	Web   Web   `mapstructure:"web"`
	Trace Trace `mapstructure:"trace"`
}

type Fields struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
	Ref   string `mapstructure:"ref"`
}

type Log struct {
	Level       string   `mapstructure:"level"`
	Formatter   string   `mapstructure:"formatter"`
	PrintMethod bool     `mapstructure:"printMethod"`
	Fields      []Fields `mapstructure:"fields"`
}

type Web struct {
	Port    int    `mapstructure:"port"`
	RunMode string `mapstructure:"runMode"`
}

type Trace struct {
	Fields []string `mapstructure:"fields"`
}
