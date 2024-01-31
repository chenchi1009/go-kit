package config

import (
	"github.com/spf13/viper"
)

// Loader 是一个配置加载器的接口
type Loader interface {
	Load(config interface{}) error
}

// ViperLoader 实现了 Loader 接口，使用 Viper 加载配置
type viperLoader struct {
	ConfigFile string
}

// NewLoader 创建一个新的配置加载器
func NewLoader(configFile string) Loader {
	loader := &viperLoader{ConfigFile: configFile}
	return loader
}

// Load 使用 Viper 加载配置到指定的结构体
func (vl *viperLoader) Load(config interface{}) error {
	v := viper.New()
	v.SetConfigFile(vl.ConfigFile)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	if err := v.Unmarshal(config); err != nil {
		return err
	}
	return nil
}
