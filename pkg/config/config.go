package config

import "time"

type Config struct {
	Log   Log   `mapstructure:"log"`
	Web   Web   `mapstructure:"web"`
	Trace Trace `mapstructure:"trace"`
	Mysql Mysql `mapstructure:"mysql"`
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

type Mysql struct {
	UserName        string        `mapstructure:"userName"`
	Password        string        `mapstructure:"password"`
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	Database        string        `mapstructure:"database"`
	Params          string        `mapstructure:"params"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
}
