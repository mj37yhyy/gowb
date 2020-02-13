package utils

import (
	"github.com/spf13/viper"
	"os"
)

type ConfigUtils struct {
}

type Config struct {
	viper *viper.Viper
}

func (c *ConfigUtils) New(configName string, configType string) (*Config, error) {
	//获取项目的执行路径
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	config := viper.New()

	config.AddConfigPath(path)       //设置读取的文件路径
	config.SetConfigName(configName) //设置读取的文件名
	config.SetConfigType(configType) //设置文件的类型
	//尝试进行配置读取
	if err := config.ReadInConfig(); err != nil {
		return nil, err
	}
	return &Config{viper: config}, nil
}

func (c *Config) Get(name string) interface{} {
	return c.viper.Get(name)
}

func (c *Config) Unmarshal(pojo interface{}) error {
	return c.viper.Unmarshal(&pojo)
}
