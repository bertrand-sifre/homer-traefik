package watcher

import (
	"context"
	"log"
	"regexp"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/events"
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
	// Initial scan
	w.scanEverything()

	filterArgs := filters.NewArgs()
	filterArgs.Add("type", "container")
	filterArgs.Add("type", "service")  // Add service events

	eventOptions := types.EventsOptions{
		Filters: filterArgs,
	}

	dockerEvents, errs := w.client.Events(ctx, eventOptions)

	for {
		select {
		case event := <-dockerEvents:
			// Rescan everything on any container or service event
			if event.Type == events.ContainerEventType || event.Type == events.ServiceEventType {
				name := event.Actor.Attributes["name"]
				if name == "" {
					name = event.Actor.ID[:12] // Use short ID if name not available
				}
				log.Printf("Docker event received: type=%s action=%s name=%s", event.Type, event.Action, name)
				w.scanEverything()
			}
		case err := <-errs:
			if err != nil {
				log.Printf("Error receiving events: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (w *DockerWatcher) scanEverything() {
	// Reset all handlers
	for _, handler := range w.handlers {
		handler.Reset()
	}

	// Scan standalone containers
	containers, err := w.client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		log.Printf("Error while scanning containers: %v", err)
		return
	}
	log.Printf("Containers found: %d", len(containers))
	for _, container := range containers {
		labels := container.Labels
		w.processLabels(labels)
	}

	// Scan Swarm services
	services, err := w.client.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		log.Printf("Error while scanning services: %v", err)
		return
	}
	log.Printf("Services found: %d", len(services))
	for _, service := range services {
		labels := service.Spec.Labels
		w.processLabels(labels)
	}
}

// Helper function to process labels
func (w *DockerWatcher) processLabels(labels map[string]string) {
	for label, value := range labels {
		if isTraefikRouterRule(label) {
			log.Printf("Traefik router rule: %s", label)
			w.processTraefikHostEvent(DockerEvent{
				Label: label,
				Value: value,
			})
		}
		if isHomerLabel(label) {
			log.Printf("Homer label: %s", label)
			w.processEvent(DockerEvent{
				Label: label,
				Value: value,
			})
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
	for _, handler := range w.handlers {
		handler.HandleEvent(dockerEvent)
	}
}

func (w *DockerWatcher) processTraefikHostEvent(dockerEvent DockerEvent) {
	label := convertTraefikLabelHost(dockerEvent.Label)
	value := convertTraefikValueHost(dockerEvent.Value)
	log.Printf("Traefik host event: %s -> %s", label, value)
	for _, handler := range w.handlers {
		handler.HandleEvent(DockerEvent{
			Label: label,
			Value: value,
		})
	}
}

func convertTraefikLabelHost(labelName string) string {
	re := regexp.MustCompile(`^traefik\.http\.routers\.([^.]+)\.rule$`)
	matches := re.FindStringSubmatch(labelName)
	if len(matches) > 1 {
		return "homer.items." + matches[1] + ".url"
	}
	return ""
}

func convertTraefikValueHost(value string) string {
	// Try first format: Host("example.com")
	re := regexp.MustCompile(`Host\(\"([^"]+)\"\)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) > 1 {
		return "https://" + matches[1]
	}

	// Try second format: Host(`example.com`)
	re = regexp.MustCompile(`Host\(\x60([^\x60]+)\x60\)`)
	matches = re.FindStringSubmatch(value)
	if len(matches) > 1 {
		return "https://" + matches[1]
	}

	log.Printf("No host found in value: %s", value)
	return ""
}