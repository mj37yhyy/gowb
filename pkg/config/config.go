package config

import "time"

type Config struct {
	Log   Log   `mapstructure:"log" yaml:"log" json:"log"`
	Web   Web   `mapstructure:"web" yaml:"web" json:"web"`
	Trace Trace `mapstructure:"trace" yaml:"trace" json:"trace"`
	Mysql Mysql `mapstructure:"mysql" yaml:"mysql" json:"mysql"`
}

type Fields struct {
	Name  string `mapstructure:"name" yaml:"name" json:"name"`
	Value string `mapstructure:"value" yaml:"value" json:"value"`
	Ref   string `mapstructure:"ref" yaml:"ref" json:"ref"`
}

type Log struct {
	Level       string   `mapstructure:"level" yaml:"level" json:"level"`
	Formatter   string   `mapstructure:"formatter" yaml:"formatter" json:"formatter"`
	PrintMethod bool     `mapstructure:"printMethod" yaml:"printMethod" json:"printMethod"`
	Fields      []Fields `mapstructure:"fields" yaml:"fields" json:"fields"`
}

type Web struct {
	Port    int    `mapstructure:"port" yaml:"port" json:"port"`
	RunMode string `mapstructure:"runMode" yaml:"runMode" json:"runMode"`
	LogSkipPath []string `mapstructure:"logSkipPath" yaml:"logSkipPath" json:"logSkipPath"`
}

type Trace struct {
	Fields []string `mapstructure:"fields" yaml:"fields" json:"fields"`
}

type Mysql struct {
	Enabled         bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	UserName        string        `mapstructure:"userName" yaml:"userName" json:"userName"`
	Password        string        `mapstructure:"password" yaml:"password" json:"password"`
	Host            string        `mapstructure:"host" yaml:"host" json:"host"`
	Port            string        `mapstructure:"port" yaml:"port" json:"port"`
	Database        string        `mapstructure:"database" yaml:"database" json:"database"`
	Params          string        `mapstructure:"params" yaml:"params" json:"params"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns" yaml:"maxOpenConns" json:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns" yaml:"maxIdleConns" json:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime" yaml:"connMaxLifetime" json:"connMaxLifetime"`
}
