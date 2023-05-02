package utils

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Config struct {
	IPAddress []string `yaml:"dataNodeServerAddr"`
}

func ReadFromUtil() []string {
	yamlFile, err := ioutil.ReadFile("./conf/conf.yaml")
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	// 解析YAML文件
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	// 输出IP地址数组
	//fmt.Println(config.IPAddr)
	return config.IPAddress
}
