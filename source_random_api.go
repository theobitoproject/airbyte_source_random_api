package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/theobitoproject/airbyte_source_random_api/streams"
	"github.com/theobitoproject/kankuro/pkg/messenger"
	"github.com/theobitoproject/kankuro/pkg/protocol"
)

const (
	minLimitAllowed = 2
	maxLimitAllowed = 100
)

type sourceRandomApi struct {
	url string
}

type sourceConfiguration struct {
	Limit int `json:"limit"`
}

func newSourceRandomApi(url string) *sourceRandomApi {
	return &sourceRandomApi{url}
}

// Spec returns the schema which described how the source connector can be configured
func (s *sourceRandomApi) Spec(
	mw messenger.MessageWriter,
	cp messenger.ConfigParser,
) (*protocol.ConnectorSpecification, error) {
	return &protocol.ConnectorSpecification{
		DocumentationURL: "https://random-data-api.com/",
		ChangeLogURL:     "https://random-data-api.com/",
		SupportedDestinationSyncModes: []protocol.DestinationSyncMode{
			protocol.DestinationSyncModeOverwrite,
		},
		ConnectionSpecification: protocol.ConnectionSpecification{
			Title:       "Random Data API",
			Description: "This source extracts data from Random Data Generator",
			Type:        "object",
			Required:    []protocol.PropertyName{"limit"},
			Properties: protocol.Properties{
				Properties: map[protocol.PropertyName]protocol.PropertySpec{
					"limit": {
						Description: fmt.Sprintf(
							"Max number of element to pull per instance. Allowed values between %d and %d",
							minLimitAllowed,
							maxLimitAllowed,
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
func (s *sourceRandomApi) Check(
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

	if sc.Limit < minLimitAllowed {
		return fmt.Errorf(
			"limit configuration value must be greater than or equal to %d",
			minLimitAllowed,
		)
	}

	if sc.Limit > maxLimitAllowed {
		return fmt.Errorf(
			"limit configuration value must be less than or equal to %d",
			maxLimitAllowed,
		)
	}

	return nil
}

// Discover returns the schema which describes the structure of the data
// that can be extracted from the source
func (s *sourceRandomApi) Discover(
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
func (s *sourceRandomApi) Read(
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

	configuredStreamChan := make(chan protocol.ConfiguredStream)
	workersDoneChan := make(chan bool)

	streamExtractor := newStreamExtractor(
		configuredStreamChan,
		hub,
		sc.Limit,
		s.url,
		workersDoneChan,
	)

	for _, stream := range cc.Streams {
		streamExtractor.addWorker(stream)
	}

	for i := 0; i < len(cc.Streams); i++ {
		<-workersDoneChan
	}

	close(workersDoneChan)
	close(configuredStreamChan)

	close(hub.GetRecordChannel())
	close(hub.GetErrorChannel())
}
