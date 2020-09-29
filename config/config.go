// Package core
package config

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	yaml "gopkg.in/yaml.v2"
)

// BaseConf information struct
type BaseConf struct {
	Port    int    `yaml:"port"`        // Main listen port
	LogPath string `yaml:"log_path"`    // log file path
	Passwd  string `yaml:"requirepass"` // password of redis
	Slaveof string `yaml:"slaveof"`     // slave of other redis
}

var baseConf *BaseConf
var appPath string

func init() {
	baseConf = &BaseConf{}

	appPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	confFilePath := appPath + "/conf/base.yaml"
	yamlFile, err := ioutil.ReadFile(confFilePath)
	if err != nil {
		panic("Fail to load base config file:" + err.Error())
	}

	err = yaml.Unmarshal(yamlFile, baseConf)
	if err != nil {
		panic("Fail to decode config file")
	}

	log.Println("Load base config success")
	log.Println("port: " + strconv.Itoa(baseConf.Port))
}

// GetLogPath : Get the path of log file
func GetLogPath() string {
	return baseConf.LogPath
}

// GetListenPort : Get the listen port
func GetListenPort() int {
	return baseConf.Port
}

// GetAppPath : Get the path of application
func GetAppPath() string {
	return appPath
}

// GetPasswd : Get password of server
func GetPasswd() string {
	return baseConf.Passwd
}

// GetSlave : Get slave of ip:port config
func GetSlave() string {
	return baseConf.Slaveof
}
