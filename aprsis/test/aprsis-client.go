// GoBalloon
// aprsis-client.go - An APRS-IS client.  Connects to APRS-IS and parses shit.

// (c) 2014, Christopher Snell

package main

import (
	"flag"
	"fmt"
	"github.com/chrissnell/GoBalloon/aprsis"
	"log"
	"os"
)

var call, pass, filter, server, rawlog string

func init() {
	flag.StringVar(&server, "server", "second.aprs.net:14580", "APRS-IS upstream")
	flag.StringVar(&call, "call", "", "Your callsign (for APRS-IS)")
	flag.StringVar(&pass, "pass", "", "Your call pass (for APRS-IS)")
	flag.StringVar(&filter, "filter", "", "Optional filter for APRS-IS server")
	flag.StringVar(&rawlog, "rawlog", "", "Path to log raw messages")

}

func main() {
	flag.Parse()

	is, err := aprsis.Dial("tcp", server)
	if err != nil {
		log.Fatalln(err)
	}

	is.Auth(call, pass, filter)

	if rawlog != "" {
		logWriter, err := os.OpenFile(rawlog,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatalln(err)
		}
		is.SetRawLog(logWriter)
	}

	for {
		fmt.Print("\n------------------------------------------\n")

		msg, err := is.Next()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Printf("%+v\n", msg)
		fmt.Print("\n------------------------------------------\n")

	}

}
