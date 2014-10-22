package main

import (
	"fmt"
	"time"

	"github.com/ninjasphere/go-castv2"
	"github.com/ninjasphere/go-castv2/controllers"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/channels"
	"github.com/ninjasphere/go-ninja/devices"
	"github.com/ninjasphere/go-ninja/model"
)

type MediaPlayer struct {
	player   *devices.MediaPlayerDevice
	client   *castv2.Client
	receiver *controllers.ReceiverController
}

func NewMediaPlayer(driver ninja.Driver, conn *ninja.Connection, id string, client *castv2.Client) (*MediaPlayer, error) {
	name := fmt.Sprintf("Chromecast %s", id)

	player, err := devices.CreateMediaPlayerDevice(driver, &model.Device{
		NaturalID:     id,
		NaturalIDType: "mdns",
		Name:          &name,
		Signatures: &map[string]string{
			"ninja:manufacturer": "Google",
			"ninja:productName":  "Chromecast",
			"ninja:thingType":    "mediaplayer",
		},
	}, conn)

	if err != nil {
		return nil, err
	}

	device := &MediaPlayer{
		player: player,
		client: client,
	}

	player.ApplyVolume = device.applyVolume
	if err := player.EnableVolumeChannel(true); err != nil {
		player.Log().Fatalf("Failed to enable volume channel: %s", err)
	}

	//_ = controllers.NewHeartbeatController(client, "Tr@n$p0rt-0", "Tr@n$p0rt-0")

	heartbeat := controllers.NewHeartbeatController(client, "sender-0", "receiver-0")
	heartbeat.Start()

	connection := controllers.NewConnectionController(client, "sender-0", "receiver-0")
	connection.Connect()

	device.receiver = controllers.NewReceiverController(client, "sender-0", "receiver-0")

	go func() {
		for {
			if err := device.onReceiverStatus(<-device.receiver.Incoming); err != nil {
				device.player.Log().Warningf("Failed to update volume status: %s", err)
			}
		}
	}()

	return device, nil
}

func (d *MediaPlayer) applyVolume(state *channels.VolumeState) error {
	d.player.Log().Infof("applyVolume called, volume %v", state)

	var err error

	// Chromecast doesn't like getting muted=true and a level at the same time.
	if state.Muted != nil && *state.Muted == true {
		_, err = d.receiver.SetVolume(&controllers.VolumePayload{
			Muted: state.Muted,
		}, time.Second*5)
	} else {
		_, err = d.receiver.SetVolume(&controllers.VolumePayload{
			Level: state.Level,
			Muted: state.Muted,
		}, time.Second*5)
	}

	/*
		  //TODO: Read the response? We always get updated instantly anyway?
		  go func() {
				d.updateStatus(response)
			}()*/
	return err
}

func (d *MediaPlayer) onReceiverStatus(status *controllers.StatusResponse) error {
	return d.player.UpdateVolumeState(&channels.VolumeState{
		Level: status.Status.Volume.Level,
		Muted: status.Status.Volume.Muted,
	})
}
