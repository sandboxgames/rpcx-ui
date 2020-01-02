package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func init() {
	LoadConfig()
}

// Config parameters
var ServerConfig = Configuration{}
var Reg Registry

func LoadConfig() {
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	//fmt.Printf("Loaded Config: \n%s\n", string(file))
	json.Unmarshal(file, &ServerConfig)
	fmt.Println("succeeded to read the config")

	switch ServerConfig.RegistryType {
	case "zookeeper":
		Reg = &ZooKeeperRegistry{}
	case "etcd":
		Reg = &EtcdRegistry{}
	case "consul":
		Reg = &ConsulRegistry{}
	default:
		fmt.Printf("unsupported registry: %s\n", ServerConfig.RegistryType)
		os.Exit(2)
	}

	if !strings.HasSuffix(ServerConfig.ServiceBaseURL, "/") {
		ServerConfig.ServiceBaseURL += "/"
	}
	Reg.InitRegistry()
}

// Configuration is configuration strcut refects the config.json
type Configuration struct {
	RegistryType   string `json:"registry_type"`
	RegistryURL    string `json:"registry_url"`
	ServiceBaseURL string `json:"service_base_url"`
	Host           string `json:"host,omitempty"`
	Port           int    `json:"port,omitempty"`
	User           string `json:"user,omitempty"`
	Password       string `json:"password,omitempty"`
}
