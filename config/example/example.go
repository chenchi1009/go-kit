package main

import (
	"fmt"

	"github.com/chenchi1009/go-kit/config"
)

// AppConfig 是应用程序的配置结构体
type AppConfig struct {
	// 定义你的配置结构体
	AppName string `mapstructure:"app_name"`
}

func main() {
	// 初始化配置加载器
	loader := config.NewLoader("config.yml")

	// 创建你自己的配置结构体
	var appConfig AppConfig

	// 加载配置到你的配置结构体
	if err := loader.Load(&appConfig); err != nil {
		panic(err)
	}

	// 输出配置
	fmt.Println("App Name:", appConfig.AppName)
}
