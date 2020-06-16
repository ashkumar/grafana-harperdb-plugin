package main

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type DatasourceSettings struct {
	URL      string
	Path     string
	Schema   string
	Table    string
	Password string
}

func LoadSettings(settings backend.DataSourceInstanceSettings) (*DatasourceSettings, error) {
	model := &DatasourceSettings{}
	err := json.Unmarshal(settings.JSONData, &model)
	if err != nil {
		return nil, fmt.Errorf("error reading settings: %s", err.Error())
	}
	model.Password = settings.DecryptedSecureJSONData["apiKey"]
	return model, nil
}
