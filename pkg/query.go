package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// QueryOptions represents datasource configuration options
type QueryOptions struct {
	URL      string `json:"url"`
	Schema   string `json:"schema"`
	Table    string `json:"table"`
	Password string `json:"apikey"`
}

// QueryModel represents a spreadsheet query.
type QueryModel struct {
	RawQuery string       `json:"query"`
	Options  QueryOptions `json:"options"`

	// Not from JSON
	TimeRange     backend.TimeRange `json:"-"`
	MaxDataPoints int64             `json:"-"`
	Interval      time.Duration     `json:"-"`
}

func GetQueryModel(query backend.DataQuery) (*QueryModel, error) {
	model := &QueryModel{}
	err := json.Unmarshal(query.JSON, &model)
	if err != nil {
		return nil, fmt.Errorf("error reading query: %s", err.Error())
	}
	// Copy directly from the well typed query
	model.TimeRange = query.TimeRange
	model.MaxDataPoints = query.MaxDataPoints
	model.Interval = query.Interval
	return model, nil
}
