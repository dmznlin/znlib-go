package test

import (
	"testing"
	"time"

	"github.com/dmznlin/znlib-go/znlib/mqtt"
	mt "github.com/eclipse/paho.mqtt.golang"
)

func TestMqtt(t *testing.T) {
	if err := mqtt.Client.Start(func(client mt.Client, message mt.Message) {
		t.Log(message.Topic(), string(message.Payload()))
	}); err != nil {
		t.Fatal(err)
	}

	defer mqtt.Client.Stop()
	if err := mqtt.Client.Subscribe("test/cmd", mqtt.Qos0); err != nil {
		t.Fatal(err)
	}

	if err := mqtt.Client.Publish("test/cmd", mqtt.Qos0, []byte("hello")); err != nil {
		t.Fatal(err)
	}

	if err := mqtt.Client.Publish("by_name", mqtt.Qos0, []byte("hello")); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)
}
