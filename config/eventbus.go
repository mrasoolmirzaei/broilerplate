package config

import "github.com/leandro-lugaresi/hub"

type ApplicationEvent struct {
	Type    string
	Payload interface{}
}

const (
	TopicUser            = "user.*"
	EventUserUpdate      = "user.update"
	EventHeartbeatCreate = "heartbeat.create"
	FieldPayload         = "payload"
	FieldUser            = "user"
	FieldUserId          = "user.id"
)

var eventHub *hub.Hub

func init() {
	eventHub = hub.New()
}

func EventBus() *hub.Hub {
	return eventHub
}
