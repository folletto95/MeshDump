package meshdump

import (
	"context"
	"log"

	mqtt "github.com/mochi-co/mqtt/server"
	"github.com/mochi-co/mqtt/server/listeners"
)

type mqttAuth struct {
	user string
	pass string
}

func (a *mqttAuth) Authenticate(u, p []byte) bool {
	return string(u) == a.user && string(p) == a.pass
}

func (a *mqttAuth) ACL(user []byte, topic string, write bool) bool {
	return true
}

// StartMQTTServer starts an embedded MQTT broker listening on addr.
func StartMQTTServer(ctx context.Context, addr, user, pass string) error {
	srv := mqtt.NewServer(nil)
	tcp := listeners.NewTCP("tcp", addr)
	err := srv.AddListener(tcp, &listeners.Config{Auth: &mqttAuth{user: user, pass: pass}})
	if err != nil {
		return err
	}
	go func() {
		if err := srv.Serve(); err != nil {
			log.Printf("mqtt server: %v", err)
		}
	}()
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	log.Printf("mqtt server started on %s", addr)
	return nil
}
