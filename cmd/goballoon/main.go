package main

import (
	"context"
	"os"

	"github.com/namsral/flag"
	logging "github.com/op/go-logging"

	"github.com/webflow/stagehand/pkg/database"
)

var log = logging.MustGetLogger("kubekite")
var dynamoClient *database.DynamoDBClient

func init() {
	var format = logging.MustStringFormatter(
		`%{color}%{time:15:04:05.000} %{shortfile} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	logBackend := logging.NewLogBackend(os.Stderr, "", 0)
	logBackendFormatter := logging.NewBackendFormatter(logBackend, format)
	logging.SetBackend(logBackendFormatter)

}

func main() {

	var debug bool
	var remoteGPS, remoteTNC, localTNC, balloonCallsign, ownerCallsign string
	var beaconInterval int

	flag.BoolVar(&debug, "debug", false, "Turn on debugging")
	flag.StringVar(&remoteGPS, "remote-gps", "", "Remote gpsd service hostname:port")
	flag.StringVar(&remoteTNC, "remote-tnc", "", "Remote TNC server hostname:port")
	flag.StringVar(&localTNC, "local-tnc", "", "Serial port device for local TNC")
	flag.StringVar(&balloonCallsign, "balloon-callsign", "", "Balloon callsign with optional SSID (e.g. NW5W, NW5W-10, etc.")
	flag.StringVar(&ownerCallsign, "owner-callsign", "", "Owner callsign with optional SSID (e.g. NW5W-4, etc.")
	flag.IntVar(&beaconInterval, "beacon-interval", 60, "APRS position beacon interval in seconds (default: 60)")

	flag.Parse()

	if remoteGPS == "" {
		log.Fatal("Must provide a remote gpsd server witih -remote-gps")
	}
	if remoteTNC == "" && localTNC == "" {
		log.Fatal("Must provide a TNC via -remote-tnc or -local-tnc")
	}
	if balloonCallsign == "" {
		log.Fatal("Must provide a callsign for the balloon payload via -balloon-callsign")
	}
	if ownerCallsign == "" {
		log.Fatal("Must provide a callsign for the balloon owner via -owner-callsign")
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan struct{}, 1)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	//wg := new(sync.WaitGroup)

	go func(cancel context.CancelFunc) {
		// If we get a SIGINT or SIGTERM, cancel the context and unblock 'done'
		// to trigger a program shutdown
		<-sigs
		cancel()
		close(done)
	}(cancel)

	// // Set up a client for DynamoDB
	// dynamoClient, _ = database.NewDynamoDBClient(dynamoDBRegion, dynamoDBTableName, kubeClusterName)

	// // Start a Kubernetes API client
	// kubeClient, err := kubernetes.NewClient(kubeConfig, kubeNamespace, kubeTimeout, debug)
	// if err != nil {
	// 	log.Critical("Could not connect to Kubernetes API: %v", err)
	// }

	// // Create a new mux.Router
	// r := mux.NewRouter()

	// // Create a new API service
	// a := api.NewService(r, dynamoClient, kubeClient, debug)

	// // Set up handlers for our mux.Router
	// a.InitializeHandlers()

	// Block until we get something on the "done" channel
	<-done

	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.

	srv.Shutdown(ctx)
	log.Info("shutting down")

}
