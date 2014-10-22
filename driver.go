package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/armon/mdns"
	"github.com/davecgh/go-spew/spew"
	"github.com/ninjasphere/go-castv2"
	"github.com/ninjasphere/go-castv2/controllers"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")

type Driver struct {
	support.DriverSupport
}

func NewDriver() (*Driver, error) {

	driver := &Driver{}

	err := driver.Init(info)
	if err != nil {
		log.Fatalf("Failed to initialize driver: %s", err)
	}

	err = driver.Export(driver)
	if err != nil {
		log.Fatalf("Failed to export driver: %s", err)
	}

	return driver, nil
}

func (d *Driver) Start(_ interface{}) error {
	log.Printf("Driver Starting")

	castService := "_googlecast._tcp.local"

	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {

			if !strings.Contains(entry.Name, castService) {
				return
			}

			fmt.Printf("Got new chromecast: %v\n", entry)

			client, err := castv2.NewClient(entry.Addr, entry.Port)

			NewMediaPlayer(d, d.Conn, entry.Name, client)

			if err != nil {
				log.Fatalf("Failed to connect to chromecast %s", entry.Addr)
			}

			//_ = controllers.NewHeartbeatController(client, "Tr@n$p0rt-0", "Tr@n$p0rt-0")

			heartbeat := controllers.NewHeartbeatController(client, "sender-0", "receiver-0")
			heartbeat.Start()

			connection := controllers.NewConnectionController(client, "sender-0", "receiver-0")
			connection.Connect()

			receiver := controllers.NewReceiverController(client, "sender-0", "receiver-0")

			response, err := receiver.GetStatus(time.Second * 5)

			spew.Dump("Status response", response, err)
		}
	}()

	go func() {
		mdns.Query(&mdns.QueryParam{
			Service: castService,
			Domain:  "local",
			Timeout: time.Second * 30,
			Entries: entriesCh,
		})
	}()

	return nil
}
