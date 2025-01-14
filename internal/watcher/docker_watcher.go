package watcher

import (
	"context"
	"log"
	"regexp"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/container"
)

type DockerWatcher struct {
	client   *client.Client
	handlers []EventHandler
}

func NewDockerWatcher(client *client.Client) *DockerWatcher {
	return &DockerWatcher{
		client:   client,
		handlers: make([]EventHandler, 0),
	}
}

func (w *DockerWatcher) AddHandler(handler EventHandler) {
	w.handlers = append(w.handlers, handler)
}

func (w *DockerWatcher) Start(ctx context.Context) {
	w.scanContainers()

	// filterArgs := filters.NewArgs()
	// filterArgs.Add("type", "container")

	// eventOptions := types.EventsOptions{
	// 	Filters: filterArgs,
	// }

	// events, errs := w.client.Events(ctx, eventOptions)

	// for {
	// 	select {
	// 	case event := <-events:
	// 		w.processEvent(event)
	// 	case err := <-errs:
	// 		if err != nil {
	// 			log.Printf("Erreur lors de la réception des événements: %v", err)
	// 		}
	// 	case <-ctx.Done():
	// 		return
	// 	}
	// }
}

func (w *DockerWatcher) scanContainers() {
	containers, err := w.client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		log.Printf("Error while scanning containers: %v", err)
		return
	}
	log.Printf("Containers found: %d", len(containers))
	for _, container := range containers {
		labels := container.Labels
		for label, value := range labels {
			// log.Printf("Label: %+v, Value: %+v", label, value)
			if isTraefikRouterRule(label) || isHomerLabel(label) {
				w.processEvent(DockerEvent{
					label: label,
					value: value,
				})
			}
		}
	}
}

func isTraefikRouterRule(labelName string) bool {
	matched, _ := regexp.MatchString(`^traefik\.http\.routers\.[^.]+\.rule$`, labelName)
	return matched
}

func isHomerLabel(labelName string) bool {
	matched, _ := regexp.MatchString(`^homer\.`, labelName)
	return matched
}	

func (w *DockerWatcher) processEvent(dockerEvent DockerEvent) {
	log.Printf("Event received: %+v", dockerEvent)

	for _, handler := range w.handlers {
		handler.HandleEvent(dockerEvent)
	}
} 