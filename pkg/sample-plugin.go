package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const variableFilter = `(?m)([a-zA-Z]+)\.([a-zA-Z]+)`

type reading struct {
	Reading int32
	Created int64
}

const millisInSecond = 1000
const nsInSecond = 1000000

func getValuesFromHarperDB(url string, schema string, table string, auth string, startPeriod int64, endPeriod int64) []byte {

	sqlStatment := fmt.Sprintf(`{"operation": "sql","sql":"select reading, __createdtime__ as created from %s.%s where __createdtime__ <= %d and __createdtime__ > %d order by __createdtime__ ASC"}`, schema, table, endPeriod, startPeriod)

	log.DefaultLogger.Info(sqlStatment)
	authParam := fmt.Sprintf("Basic %s", auth)

	data := []byte(sqlStatment)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.DefaultLogger.Error("Error reading request. " + err.Error())
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", authParam)

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.DefaultLogger.Error("Error reading response. " + err.Error())
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.DefaultLogger.Error(err.Error())
		}
		return bodyBytes
	}
	return nil
}

func getChartFields(url string, schema string, table string, auth string, startPeriod int64, endPeriod int64) ([]int32, []time.Time) {
	body := getValuesFromHarperDB(url, schema, table, auth, startPeriod, endPeriod)
	var readings []reading
	err1 := json.Unmarshal([]byte(body), &readings)
	if err1 != nil {
		fmt.Println("error:", err1)
	}
	var sendReading []int32
	var sendTime []time.Time

	for _, o := range readings {
		sendReading = append(sendReading, o.Reading)
		sendTime = append(sendTime, time.Unix(o.Created, 0))
	}

	return sendReading, sendTime
}

// newDatasource returns datasource.ServeOpts.
func newDatasource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)
	ds := &SampleDatasource{
		im: im,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

// SampleDatasource is an example datasource used to scaffold
// new datasource plugins with an backend.
type SampleDatasource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	//log.DefaultLogger.Info("QueryData", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()
	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		qmodel, err := GetQueryModel(q)
		if err != nil {
			response.Responses[q.RefID] = backend.DataResponse{
				Error: err,
			}
		} else {
			response.Responses[q.RefID] = td.query(ctx, q, qmodel.Options.URL, qmodel.Options.Schema, qmodel.Options.Table)
		}
	}

	return response, nil
}

type queryModel struct {
	Format string `json:"format"`
}

func (td *SampleDatasource) query(ctx context.Context, query backend.DataQuery, url string, schema string, table string) backend.DataResponse {

	// Unmarshal the json into our queryModel
	var qm queryModel

	response := backend.DataResponse{}

	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return response
	}

	// Log a warning if `Format` is empty.
	if qm.Format == "" {
		log.DefaultLogger.Warn("format is empty. defaulting to time series")
	}

	// create data frame response
	frame := data.NewFrame("response")

	auth := "your authorization(base64) goes here"
	readings, times := getChartFields(url, schema, table, auth, query.TimeRange.From.UnixNano()/1000000, query.TimeRange.To.UnixNano()/1000000)

	for _, v := range readings {
		fmtReading := fmt.Sprintf("%d", v)
		log.DefaultLogger.Info(fmtReading)
	}

	// add the time dimension
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, times),
	)

	// add values
	frame.Fields = append(frame.Fields,
		data.NewField("values1", nil, readings),
	)

	response.Frames = append(response.Frames, frame)

	return response
}

func (ds *SampleDatasource) getInstance(ctx backend.PluginContext) (*instanceSettings, error) {
	s, err := ds.im.Get(ctx)
	if err != nil {
		return nil, err
	}
	return s.(*instanceSettings), nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *SampleDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = "Data source is working"

	if rand.Int()%2 == 0 {
		status = backend.HealthStatusError
		message = "randomized error"
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

type instanceSettings struct {
	url        string
	password   string
	path       string
	httpClient *http.Client
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := LoadSettings(setting)
	if err != nil {
		return nil, fmt.Errorf("error reading settings: %s", err.Error())
	}

	return &instanceSettings{
		url:        settings.URL,
		path:       settings.Path,
		password:   settings.Password,
		httpClient: &http.Client{},
	}, nil
}

func (s *instanceSettings) Dispose() {
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
}
