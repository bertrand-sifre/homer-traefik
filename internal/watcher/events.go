package watcher

type DockerEvent struct {
	label   string
	value   string
}

type EventHandler interface {
	HandleEvent(event DockerEvent)
} 