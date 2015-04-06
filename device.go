package main

import (
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
	media    *controllers.MediaController
}

func NewMediaPlayer(driver ninja.Driver, conn *ninja.Connection, info map[string]string, client *castv2.Client) (*MediaPlayer, error) {

	name := info["fn"]
	sigs := map[string]string{
		"ninja:manufacturer": "Google",
		"ninja:productName":  "Chromecast",
		"ninja:thingType":    "mediaplayer",
	}

	for k, v := range info {
		sigs["chromecast:"+k] = v
	}

	player, err := devices.CreateMediaPlayerDevice(driver, &model.Device{
		NaturalID:     info["id"],
		NaturalIDType: "chromecast",
		Name:          &name,
		Signatures:    &sigs,
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

	player.ApplyPlayPause = device.applyPlayPause
	if err := player.EnableControlChannel([]string{"playing", "paused", "stopped", "buffering", "busy", "idle", "inactive"}); err != nil {
		player.Log().Fatalf("Failed to enable control channel: %s", err)
	}

	//_ = controllers.NewHeartbeatController(client, "Tr@n$p0rt-0", "Tr@n$p0rt-0")

	heartbeat := controllers.NewHeartbeatController(client, "sender-0", "receiver-0")
	heartbeat.Start()

	connection := controllers.NewConnectionController(client, "sender-0", "receiver-0")
	connection.Connect()

	device.receiver = controllers.NewReceiverController(client, "sender-0", "receiver-0")

	go func() {
		for msg := range device.receiver.Incoming {
			if err := device.onReceiverStatus(msg); err != nil {
				device.player.Log().Warningf("Failed to update status: %s", err)
			}
		}
	}()

	_, _ = device.receiver.GetStatus(time.Second * 5)

	//spew.Dump("Status response", status, err)

	//spew.Dump("Media namespace?", status.GetSessionByNamespace("urn:x-cast:com.google.cast.media"))

	device.media = controllers.NewMediaController(device.client, "sender-0", "")

	go func() {
		for {
			mediaStatus := <-device.media.Incoming
			if len(mediaStatus) == 0 {
				break
			}
			device.media.MediaSessionID = mediaStatus[0].MediaSessionID
		}
	}()

	return device, nil
}

func (d *MediaPlayer) applyPlayPause(play bool) error {
	if play {
		_, err := d.media.Play(time.Second * 3)
		//spew.Dump("Play response", response)
		return err
	}
	_, err := d.media.Pause(time.Second * 3)
	//spew.Dump("Pause response", response)
	return err
}

func (d *MediaPlayer) applyVolume(state *channels.VolumeState) error {
	d.player.Log().Infof("applyVolume called, volume %v", state)

	var err error

	// Chromecast doesn't like getting muted=true and a level at the same time.
	if state.Muted != nil && *state.Muted == true {
		_, err = d.receiver.SetVolume(&controllers.Volume{
			Muted: state.Muted,
		}, time.Second*5)
	} else {
		_, err = d.receiver.SetVolume(&controllers.Volume{
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

func (d *MediaPlayer) onReceiverStatus(status *controllers.ReceiverStatus) error {

	//spew.Dump("Got status", status)

	mediaSession := status.GetSessionByNamespace("urn:x-cast:com.google.cast.media")

	if mediaSession != nil && d.media.DestinationID != *mediaSession.TransportId {
		connection := controllers.NewConnectionController(d.client, "sender-0", *mediaSession.TransportId)
		connection.Connect()

		log.Infof("Connected to media session %s", *mediaSession.TransportId)

		d.media.SetDestinationID(*mediaSession.TransportId)
		d.media.GetStatus(time.Second * 5)
	}

	if mediaSession == nil {
		//d.media = nil
	}

	return d.player.UpdateVolumeState(&channels.VolumeState{
		Level: status.Volume.Level,
		Muted: status.Volume.Muted,
	})
}

/*

status, err := receiver.GetStatus(time.Second * 5)

spew.Dump("Status response", status, err)

spew.Dump("Media namespace?", status.GetSessionByNamespace("urn:x-cast:com.google.cast.media"))

connection2 := controllers.NewConnectionController(client, "sender-0", "web-6")
connection2.Connect()

media := controllers.NewMediaController(client, "sender-0", "receiver-0")
mediaResponse, err := media.GetStatus(time.Second * 5)
spew.Dump("Media status response", mediaResponse, err)

*/
