package main

import (
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/huin/goupnp"
	"github.com/jonaz/mdns"
	"github.com/ninjasphere/go-castv2"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/logger"
	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")
var log = logger.GetLogger(info.Name)

type Driver struct {
	support.DriverSupport
	devices map[string]*MediaPlayer
	addLock sync.Mutex
}

func NewDriver() (*Driver, error) {

	driver := &Driver{
		devices: make(map[string]*MediaPlayer),
	}

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
	log.Infof("Driver Starting")

	d.startMDNSDiscovery()
	d.startUPNPDiscovery()

	return nil
}

func (d *Driver) startMDNSDiscovery() {

	castService := "_googlecast._tcp"

	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {

			log.Debugf("Found mdns service: %v", entry)

			if !strings.Contains(entry.Name, castService) {
				return
			}

			err := d.add(entry.Addr, entry.Port, parseMdnsInfo(entry.Info))

			if err != nil {
				log.Fatalf("Failed to connect to chromecast %s", entry.Addr)
			}
		}
	}()

	go func() {
		// Start the lookup
		mdns.Lookup(castService, entriesCh)
	}()

}

func (d *Driver) startUPNPDiscovery() {

	go func() {
		for {
			devices, err := goupnp.DiscoverDevices("urn:dial-multiscreen-org:service:dial:1")

			if err == nil {
				for _, dev := range devices {

					info := parseUpnpInfo(dev.Root)
					ip := strings.TrimSuffix(strings.TrimPrefix(dev.Root.URLBaseStr, "http://"), ":8008")

					d.add(net.ParseIP(ip), 8009, info)
				}
			}

			time.Sleep(time.Second * 30)
		}
	}()

}

func (d *Driver) add(host net.IP, port int, info map[string]string) error {
	d.addLock.Lock()
	defer d.addLock.Unlock()

	if _, ok := d.devices[info["id"]]; ok {
		log.Infof("Already found id:%s name:%s", info["id"], info["fn"])
		// We already have this one.
		// TODO: Update IP
		return nil
	}

	log.Infof("Found new chromecast id:%s name:%s", info["id"], info["fn"])

	client, err := castv2.NewClient(host, port)

	if err != nil {
		return err
	}

	device, err := NewMediaPlayer(d, d.Conn, info, client)

	if err != nil {
		client.Close()
		return err
	}

	d.devices[info["id"]] = device
	return nil
}

func parseMdnsInfo(field string) map[string]string {
	vals := make(map[string]string)

	for _, part := range strings.Split(field, "|") {
		chunks := strings.Split(part, "=")
		if len(chunks) == 2 {
			vals[chunks[0]] = chunks[1]
		}
	}
	vals["discovered"] = "MDNS"
	return vals
}
func parseUpnpInfo(device *goupnp.RootDevice) map[string]string {

	reg, _ := regexp.Compile("[^a-f0-9]+")

	id := strings.TrimPrefix(device.Device.UDN, "uuid:")
	id = reg.ReplaceAllString(id, "")

	vals := map[string]string{
		"fn":         device.Device.FriendlyName,
		"id":         id,
		"discovered": "UPNP",
	}

	return vals
}
