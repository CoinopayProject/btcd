package Services

import (
	"encoding/json"
	"os"
)

type IConfigHelper interface {
	GetSection(key string) (interface{}, error)
}

type ConfigurationHelper struct {
	Filename string
	Config   string
}

func NewConfigHelper(filename string) (*ConfigurationHelper, error) {
	if filename == "" {
		filename = "appSettings.json"
	}
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &ConfigurationHelper{
		Filename: filename,
		Config:   string(content),
	}, nil
}

func (ch *ConfigurationHelper) GetSection(key string) (interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(ch.Config), &data)
	if err != nil {
		return nil, err
	}
	return data[key], nil
}
