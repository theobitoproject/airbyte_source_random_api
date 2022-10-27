package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/theobitoproject/airbyte_source_random_api/streams"
	"github.com/theobitoproject/kankuro/pkg/messenger"
	"github.com/theobitoproject/kankuro/pkg/protocol"
)

// MaxLimitAllowed is the max amount of objects that can be fecthed from random API platform
const (
	MinLimitAllowed = 2
	MaxLimitAllowed = 100
)

// RandomAPISource defines the source from which data will come
// from random api platform
// See: https://random-data-api.com/
type RandomAPISource struct {
	url string
}

type sourceConfiguration struct {
	Limit int `json:"limit"`
}

// NewRandomAPISource creates a new instance of RandomAPISource
func NewRandomAPISource(url string) *RandomAPISource {
	return &RandomAPISource{url}
}

// Spec returns the schema which described how the source connector can be configured
func (s *RandomAPISource) Spec(
	mw messenger.MessageWriter,
	cp messenger.ConfigParser,
) (*protocol.ConnectorSpecification, error) {
	return &protocol.ConnectorSpecification{
		DocumentationURL:      "https://random-data-api.com/",
		ChangeLogURL:          "https://random-data-api.com/",
		SupportsIncremental:   false,
		SupportsNormalization: true,
		SupportsDBT:           true,
		SupportedDestinationSyncModes: []protocol.DestinationSyncMode{
			protocol.DestinationSyncModeOverwrite,
		},
		ConnectionSpecification: protocol.ConnectionSpecification{
			Title:       "Random Data API",
			Description: "Random Data Source API",
			Type:        "object",
			Required:    []protocol.PropertyName{"limit"},
			Properties: protocol.Properties{
				Properties: map[protocol.PropertyName]protocol.PropertySpec{
					"limit": {
						Description: fmt.Sprintf(
							"max number of element to pull per instance. Allowed values between %d and %d",
							MinLimitAllowed,
							MaxLimitAllowed,
						),
						PropertyType: protocol.PropertyType{
							Type: []protocol.PropType{
								protocol.Integer,
							},
						},
					},
				},
			},
		},
	}, nil
}

// Check verifies that, given a configuration, data can be accessed properly
func (s *RandomAPISource) Check(
	mw messenger.MessageWriter,
	cp messenger.ConfigParser,
) error {
	err := mw.WriteLog(protocol.LogLevelInfo, "checking random api source")
	if err != nil {
		return err
	}

	apiNames := []string{
		streams.AppliancesName,
		streams.BeersName,
	}

	for _, apiName := range apiNames {
		uri := fmt.Sprintf("%s/%s", s.url, apiName)

		res, err := http.Get(uri)
		if err != nil {
			return err

		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("response error on checking random api source: %d", res.StatusCode)
		}

		// prevent throttling
		time.Sleep(200 * time.Millisecond)
	}

	var sc sourceConfiguration
	err = cp.UnmarshalSourceConfigPath(&sc)
	if err != nil {
		return err
	}

	if sc.Limit < MinLimitAllowed {
		return fmt.Errorf("limit configuration value must be greater than or equal to %d", MinLimitAllowed)
	}

	if sc.Limit > MaxLimitAllowed {
		return fmt.Errorf("limit configuration value must be less than or equal to %d", MaxLimitAllowed)
	}

	return nil
}

// Discover returns the schema which describes the structure of the data
// that can be extracted from the source
func (s *RandomAPISource) Discover(
	mw messenger.MessageWriter,
	cp messenger.ConfigParser,
) (*protocol.Catalog, error) {
	return &protocol.Catalog{Streams: []protocol.Stream{
		streams.GetBeersStream(),
		streams.GetAppliancesStream(),
	}}, nil
}

// Read fetches data from the source and
// communicates all records to the record channel
// Note: To stop execution, do not use Close method inside the implementation
// Instead, send a value to the done channel (doneChannel <- true)
func (s *RandomAPISource) Read(
	cc *protocol.ConfiguredCatalog,
	mw messenger.MessageWriter,
	cp messenger.ConfigParser,
	hub messenger.ChannelHub,
) {
	err := mw.WriteLog(protocol.LogLevelInfo, "running read")
	if err != nil {
		hub.GetErrorChannel() <- err
	}

	var sc sourceConfiguration
	err = cp.UnmarshalSourceConfigPath(&sc)
	if err != nil {
		hub.GetErrorChannel() <- err
		return
	}

	doneStreamChannel := make(chan bool)

	go func() {
		for i := 0; i < len(cc.Streams); i++ {
			<-doneStreamChannel
		}

		hub.GetClosingChannel() <- true

		close(hub.GetRecordChannel())
		close(hub.GetErrorChannel())
		close(hub.GetClosingChannel())
	}()

	for _, stream := range cc.Streams {

		switch stream.Stream.Name {
		case streams.AppliancesName:
			go s.fetchAppliances(stream, sc.Limit, hub, doneStreamChannel)

		case streams.BeersName:
			go s.fetchBeers(stream, sc.Limit, hub, doneStreamChannel)

		default:
			hub.GetErrorChannel() <- fmt.Errorf("stream not supported: %s", stream.Stream.Name)
		}
	}
}

func (s *RandomAPISource) fetchAppliances(
	stream protocol.ConfiguredStream,
	limit int,
	hub messenger.ChannelHub,
	doneStreamChannel chan bool,
) {
	var appliances []streams.Appliance
	err := s.fetchDataForStream(
		stream,
		limit,
		&appliances,
	)
	if err != nil {
		hub.GetErrorChannel() <- err
		doneStreamChannel <- true
		return
	}

	for _, appliance := range appliances {
		select {
		case <-hub.GetClosingChannel():
			doneStreamChannel <- true
			return
		default:
			rec, err := marshalRecord(stream, appliance)
			if err != nil {
				hub.GetErrorChannel() <- err
				doneStreamChannel <- true
				return
			}

			hub.GetRecordChannel() <- rec
		}
	}

	doneStreamChannel <- true
}

func (s *RandomAPISource) fetchBeers(
	stream protocol.ConfiguredStream,
	limit int,
	hub messenger.ChannelHub,
	doneStreamChannel chan bool,
) {
	var beers []streams.Beer
	err := s.fetchDataForStream(
		stream,
		limit,
		&beers,
	)
	if err != nil {
		hub.GetErrorChannel() <- err
		doneStreamChannel <- true
		return
	}

	for _, beer := range beers {
		select {
		case <-hub.GetClosingChannel():
			doneStreamChannel <- true
			return
		default:
			rec, err := marshalRecord(stream, beer)
			if err != nil {
				hub.GetErrorChannel() <- err
				doneStreamChannel <- true
				return
			}

			hub.GetRecordChannel() <- rec
		}
	}

	doneStreamChannel <- true
}

func (s *RandomAPISource) fetchDataForStream(
	stream protocol.ConfiguredStream,
	limit int,
	records interface{},
) error {
	uri := fmt.Sprintf("%s/%s?size=%d", s.url, stream.Stream.Name, limit)

	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO: check status code

	return json.NewDecoder(resp.Body).Decode(records)
}

func marshalRecord(
	stream protocol.ConfiguredStream,
	rec interface{},
) (*protocol.Record, error) {
	var recData *protocol.RecordData

	data, err := json.Marshal(rec)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &recData)
	if err != nil {
		return nil, err
	}

	return &protocol.Record{
		Namespace: stream.Stream.Namespace,
		Data:      recData,
		Stream:    stream.Stream.Name,
	}, nil
}
