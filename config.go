package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Device struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

type Config struct {
	ListenIP  string   `yaml:"listen_ip"`
	Port      string   `yaml:"port"`
	VerifyURL string   `yaml:"verify_url"`
	McShopID  int64    `yaml:"mc_shop_id"`
	SecretKey string   `yaml:"secret_key"`
	Devices   []Device `yaml:"devices"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
