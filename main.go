package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/signalfx/signalflow-client-go/v2/signalflow"
)

var ErrInvalidConfig = errors.New("could not find splunk access token or ream")
var ErrCreatingClient = errors.New("error creating signalflow client")

type SplunkConfig struct {
	realm       string
	accessToken string
}

func NewSplunkConnection(ctx context.Context, config *SplunkConfig) (*signalflow.Client, error) {
	if config.realm == "" || config.accessToken == "" {
		return nil, ErrInvalidConfig
	}

	apiClient, err := signalflow.NewClient(
		signalflow.StreamURLForRealm(config.realm),
		signalflow.AccessToken(config.accessToken),
		signalflow.OnError(func(err error) {
			fmt.Printf("Error in signalflow client: %v\n", err)
		}),
	)

	if err != nil {
		return nil, ErrCreatingClient
	}

	return apiClient, nil
}

func main() {
	args := os.Args[1:]
	query := args[0]

	if query == "" {
		panic("No query provided.")
	}

	config := SplunkConfig{
		realm:       os.Getenv("SPLUNK_REALM"),
		accessToken: os.Getenv("SPLUNK_TOKEN"),
	}

	ctx := context.Background()
	client, err := NewSplunkConnection(ctx, &config)

	if err != nil {
		panic("Cannot establishing connection to signalflow")
	}

	start := time.Now().Add(-1 * time.Minute)
	stop := time.Now()

	comp, err := client.Execute(ctx, &signalflow.ExecuteRequest{
		Program: query,
		Start:   start,
		Stop:    stop,
	})

	if err != nil {
		panic("Problem executing query")
	}

	max := math.Inf(-1)
	min := math.Inf(1)

	valueSum := 0.0
	valueCount := 0

	go func() {
		time.Sleep(10 * time.Second)

		if err := comp.Stop(context.Background()); err != nil {
			fmt.Printf("Failed to stop computation")
		}
	}()

	for msg := range comp.Data() {
		if len(msg.Payloads) == 0 {
			continue
		}

		for _, pl := range msg.Payloads {
			value := pl.Float64()

			if value > max {
				max = value
			}

			if value < min {
				min = value
			}

			valueSum += value
			valueCount++
		}

		comp.Stop(ctx)
	}

	if err := comp.Err(); err != nil {
		log.Fatalf("computation error: %v", err)
	}

	fmt.Print("Values for the last minute:\n")
	fmt.Printf(" - avg: %f\n", valueSum/float64(valueCount))
	fmt.Printf(" - max: %f\n", max)
	fmt.Printf(" - min: %f\n", min)
}
