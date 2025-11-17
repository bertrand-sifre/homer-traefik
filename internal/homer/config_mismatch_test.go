package homer

import (
	"fmt"
	"os"
	"testing"
	"homer-traefik/internal/watcher"
)

// TestMismatchScenario tests the scenario where router name doesn't match item name
func TestMismatchScenario(t *testing.T) {
	tmpFile := "test_mismatch.yaml"
	defer os.Remove(tmpFile)

	handler := NewConfigHandler()
	handler.filePath = tmpFile

	// Simulate the labels from the user's example
	// Router: gluetun-torrent → creates homer.items.gluetun-torrent.url
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.items.gluetun-torrent.url",
		Value: "https://torrent.bertrand-sifre.com",
	})
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.items.torrent.name",
		Value: "qbittorrent",
	})
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.items.torrent.service",
		Value: "warez",
	})

	// Router: gluetun-slskd → creates homer.items.gluetun-slskd.url
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.items.gluetun-slskd.url",
		Value: "https://slskd.bertrand-sifre.com",
	})
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.items.slskd.name",
		Value: "Slskd",
	})
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.items.slskd.service",
		Value: "warez",
	})

	// Print the items map to see what's happening
	fmt.Println("\n=== Items in the map ===")
	for key, item := range handler.items {
		fmt.Printf("Key: %s\n", key)
		fmt.Printf("  Name: %s\n", item.Name)
		fmt.Printf("  URL: %s\n", item.Url)
		fmt.Printf("  Service: %s\n", item.Service)
		fmt.Println()
	}

	// Read and print the config file
	content, _ := os.ReadFile(tmpFile)
	fmt.Println("=== Config file content ===")
	fmt.Println(string(content))
}
