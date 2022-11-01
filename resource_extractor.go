package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/theobitoproject/kankuro/pkg/messenger"
	"github.com/theobitoproject/kankuro/pkg/protocol"
)

type streamExtractor struct {
	configuredStreamChan chan protocol.ConfiguredStream
	hub                  messenger.ChannelHub
	limit                int
	url                  string
	workersDoneChan      chan bool
}

func newStreamExtractor(
	configuredStreamChan chan protocol.ConfiguredStream,
	hub messenger.ChannelHub,
	limit int,
	url string,
	workersDoneChan chan bool,
) *streamExtractor {
	return &streamExtractor{
		configuredStreamChan,
		hub,
		limit,
		url,
		workersDoneChan,
	}
}

func (se *streamExtractor) addWorker(
	configuredStream protocol.ConfiguredStream,
) {
	go func() {
		defer se.removeWorker()

		rcds, err := se.fetchData(configuredStream)
		if err != nil {
			se.hub.GetErrorChannel() <- err
			return
		}

		err = se.sendRecords(configuredStream, rcds)
		if err != nil {
			se.hub.GetErrorChannel() <- err
			return
		}
	}()
}

func (se *streamExtractor) sendRecords(
	configuredStream protocol.ConfiguredStream,
	rcds []*protocol.RecordData,
) error {
	for _, rcd := range rcds {
		var recData *protocol.RecordData

		data, err := json.Marshal(rcd)
		if err != nil {
			return err
		}

		err = json.Unmarshal(data, &recData)
		if err != nil {
			return err
		}

		rec := &protocol.Record{
			Namespace: configuredStream.Stream.Namespace,
			Data:      recData,
			Stream:    configuredStream.Stream.Name,
		}

		se.hub.GetRecordChannel() <- rec
	}

	return nil
}

func (se *streamExtractor) fetchData(
	configuredStream protocol.ConfiguredStream,
) ([]*protocol.RecordData, error) {
	uri := fmt.Sprintf(
		"%s/%s?size=%d",
		se.url,
		configuredStream.Stream.Name,
		se.limit,
	)

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: check status code

	var rcds []*protocol.RecordData
	err = json.NewDecoder(resp.Body).Decode(&rcds)
	if err != nil {
		return nil, err
	}

	return rcds, nil
}

func (se *streamExtractor) removeWorker() {
	se.workersDoneChan <- true
}
