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

		// Special handling for URL field from Traefik routers
		// If we're setting a URL and the itemId doesn't match an existing item with a service,
		// try to find a matching item by checking if itemId contains another item's name
		if itemField == "url" {
			matchedId := h.findMatchingItem(itemId)
			if matchedId != "" {
				itemId = matchedId
				log.Printf("Matched router '%s' to item '%s'", matches[1], itemId)
			}
		}

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

// findMatchingItem tries to find an existing item that matches the router name
// For example, if routerName is "gluetun-torrent", it will find "torrent"
func (h *ConfigHandler) findMatchingItem(routerName string) string {
	// First check if there's an exact match
	if item, exists := h.items[routerName]; exists && item.Service != "" {
		return routerName
	}

	// Try to find a partial match
	// Check if routerName contains any existing item name or vice versa
	for existingId, item := range h.items {
		// Skip if the existing item doesn't have a service (it's probably from Traefik)
		if item.Service == "" {
			continue
		}

		// Check if router name contains the item name (e.g., "gluetun-torrent" contains "torrent")
		if strings.Contains(routerName, existingId) {
			return existingId
		}

		// Check if item name contains the router name (less common but possible)
		if strings.Contains(existingId, routerName) {
			return existingId
		}
	}

	return ""
}

// Update services from items
func (h *ConfigHandler) updateServices() {
	// First, merge items where a router URL exists without a service
	// with items that have a service but no URL
	h.mergeMatchingItems()

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

// mergeMatchingItems merges items where router names don't match item names
// For example, merges "gluetun-torrent" (has URL) with "torrent" (has name/service)
func (h *ConfigHandler) mergeMatchingItems() {
	// Find items that have a URL but no service (these are from Traefik routers)
	routerItems := make(map[string]Item)
	for id, item := range h.items {
		if item.Url != "" && item.Service == "" {
			routerItems[id] = item
		}
	}

	// For each router item, try to find a matching item with a service
	for routerId, routerItem := range routerItems {
		// Try to find a matching item
		for itemId, item := range h.items {
			// Skip the router item itself
			if itemId == routerId {
				continue
			}

			// Skip items without a service
			if item.Service == "" {
				continue
			}

			// Check if there's a match (router name contains item name or vice versa)
			if strings.Contains(routerId, itemId) || strings.Contains(itemId, routerId) {
				log.Printf("Merging router '%s' URL into item '%s'", routerId, itemId)
				// Merge the URL from the router item into the service item
				item.Url = routerItem.Url
				h.items[itemId] = item
				// Remove the router item since it's been merged
				delete(h.items, routerId)
				break
			}
		}
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

func (h *ConfigHandler) Reset() {
	h.config = &Config{
		Title:    "Demo dashboard",
		Services: make([]Service, 0),
	}
	h.services = make(map[string]Service)
	h.items = make(map[string]Item)
	h.writeConfig()  // Write empty config to file
}