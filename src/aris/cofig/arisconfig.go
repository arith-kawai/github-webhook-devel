package arisconf

import (
	"io/ioutil"
	"fmt"
	"gopkg.in/yaml.v2"
)

const APPLICATION_CONFIG_FILEPATH = "./application.yaml"

type Config struct {
	Webserver struct {
			  Port int
		  }
	Github    struct {
			  Secret string
		  }
	Googleapi struct {
			  ClientSecretFilePath string `yaml:"clientSecretFilePath"`
		  }
	Database  struct {
			  Type           string
			  Protocol       string
			  Host           string
			  User           string
			  Password       string
			  DbName         string `yaml:"dbName"`
			  CaCertFilePath string `yaml:"caCertFilePath"`
		  }
}

var sharedInstance *Config = NewConfig()

func newSingle() *Config {
	// 何かしらの初期化処理
	return &Config{/* 初期化 */
	}
}
func NewConfig() *Config {
	appdata, err := ioutil.ReadFile(APPLICATION_CONFIG_FILEPATH)
	if err != nil {
		fmt.Printf("Unable to read application.yaml: %v", err)
		panic(err)
	}
	application := &Config{}

	err = yaml.Unmarshal([]byte(appdata), &application)
	if err != nil {
		fmt.Printf("error: %v", err)
		panic(err)
	}
	return application
}

func GetSharedConfig() *Config {
	return sharedInstance
}