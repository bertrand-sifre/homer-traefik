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
	w.firstScan()

	filterArgs := filters.NewArgs()
	filterArgs.Add("type", "container")

	eventOptions := types.EventsOptions{
		Filters: filterArgs,
	}

	dockerEvents, errs := w.client.Events(ctx, eventOptions)

	for {
		select {
		case event := <-dockerEvents:
			if event.Type == events.ContainerEventType {
				labels := event.Actor.Attributes
				for label, value := range labels {
					if isTraefikRouterRule(label) {
						w.processTraefikHostEvent(DockerEvent{
							Label: label,
							Value: value,
						})
					}
					if isHomerLabel(label) {
						w.processEvent(DockerEvent{
							Label: label,
							Value: value,
						})
					}
				}
			}
		case err := <-errs:
			if err != nil {
				log.Printf("Erreur lors de la réception des événements: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (w *DockerWatcher) firstScan() {
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
		return "homer.services." + matches[1] + ".url"
	}
	return ""
}

func convertTraefikValueHost(value string) string {
	re := regexp.MustCompile(`Host\(\"([^"]+)\"\)`)
	matches := re.FindStringSubmatch(value)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}