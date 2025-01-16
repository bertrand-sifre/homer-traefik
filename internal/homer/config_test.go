package homer

import (
	"os"
	"testing"
	"homer-traefik/internal/watcher"
	"github.com/stretchr/testify/require"
)

func TestConfigHandler(t *testing.T) {
	// Create a temporary file for testing
	tmpFile := "test_config.yaml"
	defer os.Remove(tmpFile)

	// Create a new instance of ConfigHandler
	handler := NewConfigHandler()
	handler.filePath = tmpFile

	// Set the title
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.title",
		Value: "My Dashboard",
	})

	// Add an item to the apps service
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.items.test1.name",
		Value: "Test Service",
	})
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.items.test1.url",
		Value: "http://test.com",
	})
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.items.test1.service",
		Value: "apps",
	})

	// Read the file content
	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	expectedContent := `title: My Dashboard
services:
  - name: apps
    icon: fas fa-rocket
    items:
      - name: Test Service
        url: http://test.com
        service: apps
`
	require.Equal(t, expectedContent, string(content))
} 
