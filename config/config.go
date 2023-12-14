package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type Config struct {
	Mysql []MysqlConfig `json:"mysql"`
	Redis RedisConfig   `json:"redis"`
}

type MysqlConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Addr     string `json:"addr"`
	Dbname   string `json:"dbname"`
	Option   string `json:"option"`
	UqName   string `json:"uqName"`
}

type RedisConfig struct {
	Database string `json:"database"`
	Host     string `json:"host"`
	Port     int64  `json:"port"`
	Default  bool   `json:"default"`
}

var config Config
var configOnce sync.Once

func GetConfig() *Config {
	configOnce.Do(func() {
		v := viper.New()
		v.SetConfigName("config")
		v.AddConfigPath("./config")
		v.AddConfigPath("../config")
		v.AddConfigPath("../../config")
		v.AddConfigPath("../")
		v.SetConfigType("json")
		err := v.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("fatal error config file: %s \n", err))
		}

		if err := v.Unmarshal(&config); err != nil {
			panic("parse config file failed")
		}

	})

	return &config
}

func LoadConfigByFile() *Config {
	fmt.Println(GetCurrentAbPath())
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return &config
	}

	//file, err := os.Open( /*GetCurrentAbPath() +*/ "./config/config.json")
	configFilePath := filepath.Join(currentDir /*"..",*/, "config", "config.json")

	file, err := os.Open(configFilePath)

	if err != nil {
		fmt.Println("Error opening config file:", err)

		return &config
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config file:", err)

	}

	fmt.Println("加载config完成：", &config)

	return &config
}

func GetCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}

	return dir
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}
