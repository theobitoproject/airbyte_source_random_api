package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/theobitoproject/airbyte_source_random_api/streams"
	"github.com/theobitoproject/kankuro/protocol"
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
func NewRandomAPISource(url string) RandomAPISource {
	return RandomAPISource{url}
}

// Spec returns the input "form" spec needed for your source
func (s RandomAPISource) Spec(
	messenger protocol.Messenger,
	configParser protocol.ConfigParser,
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

// Check verifies the source - usually verify creds/connection etc.
func (s RandomAPISource) Check(
	messenger protocol.Messenger,
	configParser protocol.ConfigParser,
) error {
	err := messenger.WriteLog(protocol.LogLevelInfo, "checking random api source")
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
	err = configParser.UnmarshalSourceConfigPath(&sc)
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

// Discover returns the schema of the data you want to sync
func (s RandomAPISource) Discover(
	messenger protocol.Messenger,
	configParser protocol.ConfigParser,
) (*protocol.Catalog, error) {
	return &protocol.Catalog{Streams: []protocol.Stream{
		streams.GetBeersStream(),
		streams.GetAppliancesStream(),
	}}, nil
}

// Read will read the actual data from your source and use
// tracker.Record(), tracker.State() and tracker.Log() to sync data
// with airbyte/destinations
// MessageTracker is thread-safe and so it is completely find to
// spin off goroutines to sync your data (just don't forget your waitgroups :))
// returning an error from this will cancel the sync and returning a nil
// from this will successfully end the sync
func (s RandomAPISource) Read(
	configuredCat *protocol.ConfiguredCatalog,
	messenger protocol.Messenger,
	configParser protocol.ConfigParser,
) error {
	err := messenger.WriteLog(protocol.LogLevelInfo, "running read")
	if err != nil {
		return err
	}

	var sc sourceConfiguration
	err = configParser.UnmarshalSourceConfigPath(&sc)
	if err != nil {
		return err
	}

	record := func(
		rec interface{},
		stream protocol.ConfiguredStream,
	) error {
		return messenger.WriteRecord(rec, stream.Stream.Name, stream.Stream.Namespace)
	}

	// TODO: use goroutines to fetch and record faster for every stream
	// for loop is not very efficient
	for _, stream := range configuredCat.Streams {

		switch stream.Stream.Name {
		case streams.AppliancesName:
			var appliances []streams.Appliance
			err = s.fetch(stream.Stream.Name, sc.Limit, &appliances)
			if err != nil {
				return err
			}
			for _, appliance := range appliances {
				err = record(appliance, stream)
				if err != nil {
					return err
				}
			}

		case streams.BeersName:
			var beers []streams.Beer
			err = s.fetch(stream.Stream.Name, sc.Limit, &beers)
			if err != nil {
				return err
			}
			for _, beer := range beers {
				err = record(beer, stream)
				if err != nil {
					return err
				}
			}

		default:
			return fmt.Errorf("stream not supported: %s", stream.Stream.Name)
		}
	}

	return nil
}

func (s RandomAPISource) fetch(streamName string, limit int, decode interface{}) error {
	uri := fmt.Sprintf("%s/%s?size=%d", s.url, streamName, limit)

	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO: check status code

	return json.NewDecoder(resp.Body).Decode(decode)
}
