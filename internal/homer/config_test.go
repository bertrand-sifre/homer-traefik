package homer

import (
	"os"
	"testing"
	"homer-traefik/internal/watcher"
	"github.com/stretchr/testify/require"
)

func TestConfigHandler(t *testing.T) {
	// Créer un fichier temporaire pour les tests
	tmpFile := "test_config.yaml"
	defer os.Remove(tmpFile)

	// Créer une nouvelle instance de ConfigHandler
	handler := NewConfigHandler()
	handler.filePath = tmpFile

	// Affectation du titre
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.title",
		Value: "Mon Dashboard",
	})

	// Ajout d'un item dans le service apps
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

	// Lire le contenu du fichier
	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	expectedContent := `title: Mon Dashboard
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
