package homer

import (
	"log"
	"os"
	"gopkg.in/yaml.v3"
	"homer-traefik/internal/watcher"
	"strings"
	"regexp"
)

type Config struct {
	Title    string    `yaml:"title,omitempty"`
	Subtitle string    `yaml:"subtitle,omitempty"`
	Logo     string    `yaml:"logo,omitempty"`
	Icon     string    `yaml:"icon,omitempty"`
	Header   bool      `yaml:"header,omitempty"`
	Footer   string    `yaml:"footer,omitempty"`
	Theme    string    `yaml:"theme,omitempty"`
	Services []Service `yaml:"services,omitempty"`
}

type Service struct {
	Name  string    `yaml:"name,omitempty"`
	Icon  string    `yaml:"icon,omitempty"`
	Items []Item    `yaml:"items,omitempty"`
}

type Item struct {
	Name     string `yaml:"name,omitempty"`
	Subtitle string `yaml:"subtitle,omitempty"`
	Tag      string `yaml:"tag,omitempty"`
	Url      string `yaml:"url,omitempty"`
	Icon     string `yaml:"icon,omitempty"`
	Service  string `yaml:"-"` //ignore this field
}

type ConfigHandler struct {
	config    *Config
	services  map[string]Service
	items     map[string]Item
	filePath  string
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		config: &Config{
			Title: "Demo dashboard",
			Services: make([]Service, 0),
		},
		services: make(map[string]Service),
		items:    make(map[string]Item),
		filePath: "config.yml",
	}
}

func (h *ConfigHandler) HandleEvent(event watcher.DockerEvent) {
	h.updateConfig(event)
	h.writeConfig()
}

func (h *ConfigHandler) updateConfig(event watcher.DockerEvent) {
	if event.Label == "homer.title" {
		h.config.Title = event.Value
	}
	// Check if label starts with homer.items.
	if strings.HasPrefix(event.Label, "homer.items.") {
		re := regexp.MustCompile(`^homer\.items\.(.*)\.(.*)`)
		matches := re.FindStringSubmatch(event.Label)
		if len(matches) < 3 {
			return
		}
		
		itemId := matches[1]
		itemField := matches[2]

		// Get existing item or create a new one
		item, exists := h.items[itemId]
		if !exists {
			item = Item{}
		}

		// Update the appropriate field
		switch itemField {
		case "name":
			item.Name = event.Value
		case "subtitle":
			item.Subtitle = event.Value
		case "url":
			item.Url = event.Value
		case "icon":
			item.Icon = event.Value
		case "tag":
			item.Tag = event.Value
		case "service":
			item.Service = event.Value
		}

		// Put the updated item back in the map
		h.items[itemId] = item

		// Update the configuration
		h.updateServices()
	}
}

// Update services from items
func (h *ConfigHandler) updateServices() {
	// Group items by service
	serviceItems := make(map[string][]Item)
	for _, item := range h.items {
		if item.Service != "" {
			serviceItems[item.Service] = append(serviceItems[item.Service], item)
		}
	}

	// Update the services list in the config
	h.config.Services = make([]Service, 0)
	for serviceName, items := range serviceItems {
		h.config.Services = append(h.config.Services, Service{
			Name:  serviceName,
			Items: items,
		})
	}
}

func (h *ConfigHandler) writeConfig() {
	file, err := os.OpenFile(h.filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Printf("Error opening config file: %v", err)
		return
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		log.Printf("Error truncating file: %v", err)
		return
	}
	if _, err := file.Seek(0, 0); err != nil {
		log.Printf("Error seeking file: %v", err)
		return
	}

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(h.config); err != nil {
		log.Printf("Error encoding config: %v", err)
		return
	}
}