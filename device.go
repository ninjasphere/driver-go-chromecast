package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ninjasphere/go-castv2"
	"github.com/ninjasphere/go-castv2/api"
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

	response, err := device.receiver.GetStatus(time.Second * 5)

	device.player.Log().Infof("Status response %v error:%e", response, err)

	return device, nil
}

func (d *MediaPlayer) applyVolume(state *channels.VolumeState) error {
	d.player.Log().Infof("applyVolume called, volume %v", state)

	response, err := d.receiver.SetVolume(&controllers.VolumePayload{
		Volume: state.Level,
		Mute:   state.Muted,
	}, time.Second*5)

	go func() {
		d.updateStatus(response)
	}()
	return err
}

func (d *MediaPlayer) updateStatus(message *api.CastMessage) error {

	// TODO: Read the rest of the status

	var volume controllers.VolumePayload

	err := json.Unmarshal([]byte(*message.PayloadUtf8), &volume)

	if err != nil {
		return err
	}

	return d.player.UpdateVolumeState(&channels.VolumeState{
		Level: volume.Volume,
		Muted: volume.Mute,
	})
}
