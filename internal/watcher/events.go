package watcher

type DockerEvent struct {
	Label   string
	Value   string
}

type EventHandler interface {
	HandleEvent(event DockerEvent)
	Reset()
} 