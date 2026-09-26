package homer

import (
	"log"
	"os"
	"gopkg.in/yaml.v3"
	"homer-traefik/internal/watcher"
	"regexp"
	"sort"
	"strings"
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
//
// Les labels arrivent un par un, dans l'ordre (aléatoire) d'itération des maps de labels
// Docker : aucune décision ne peut donc être prise à l'arrivée d'un label. h.items garde
// chaque entrée sous son propre id, et la résolution des URLs se refait entièrement ici,
// sans rien modifier dans h.items.
func (h *ConfigHandler) updateServices() {
	serviceItems := make(map[string][]Item)
	for id, item := range h.items {
		if item.Service == "" {
			continue
		}
		if item.Url == "" {
			item.Url = h.routerUrlFor(id)
		}
		serviceItems[item.Service] = append(serviceItems[item.Service], item)
	}

	// Ordre stable : sans tri, chaque réécriture du fichier mélangeait services et items.
	names := make([]string, 0, len(serviceItems))
	for name := range serviceItems {
		names = append(names, name)
	}
	sort.Strings(names)

	h.config.Services = make([]Service, 0, len(names))
	for _, name := range names {
		items := serviceItems[name]
		sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
		h.config.Services = append(h.config.Services, Service{Name: name, Items: items})
	}
}

// routerUrlFor cherche l'URL d'un item dont aucun router Traefik ne porte le nom, parmi
// les routers orphelins (URL sans item Homer du même nom) : « gluetun-torrent » pour
// l'item « torrent ».
//
// N'est appelé que pour un item **sans URL** : un item qui a son propre router
// (« trackarr ») ne reçoit jamais celle d'un voisin qui le contient
// (« trackarr-discovery »), ni « sonarr » celle de « sonarr-api ». Parmi plusieurs
// candidats, le nom le plus court l'emporte, puis l'ordre alphabétique : le résultat ne
// dépend pas de l'ordre d'arrivée des labels.
func (h *ConfigHandler) routerUrlFor(itemId string) string {
	best := ""
	for routerId, router := range h.items {
		if routerId == itemId || router.Service != "" || router.Url == "" {
			continue
		}
		if !strings.Contains(routerId, itemId) && !strings.Contains(itemId, routerId) {
			continue
		}
		if best == "" || len(routerId) < len(best) || (len(routerId) == len(best) && routerId < best) {
			best = routerId
		}
	}
	if best == "" {
		return ""
	}
	return h.items[best].Url
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