package homer

import (
	"os"
	"testing"
	"homer-traefik/internal/watcher"
)

func TestConfigHandler(t *testing.T) {
	// Créer un fichier temporaire pour les tests
	tmpFile := "test_config.yaml"
	defer os.Remove(tmpFile)

	// Créer une nouvelle instance de ConfigHandler
	handler := NewConfigHandler()
	handler.filePath = tmpFile

	// Test 1: Vérifier le titre par défaut
	if handler.config.Title != "Demo dashboard" {
		t.Errorf("Titre par défaut incorrect, obtenu: %s, attendu: %s", handler.config.Title, "Demo dashboard")
	}

	// Test 2: Modifier le titre
	handler.HandleEvent(watcher.DockerEvent{
		Label: "homer.title",
		Value: "Mon Dashboard",
	})

	// Vérifier que le titre a été mis à jour
	if handler.config.Title != "Mon Dashboard" {
		t.Errorf("Titre non mis à jour, obtenu: %s, attendu: %s", handler.config.Title, "Mon Dashboard")
	}

	// Test 3: Ajouter un item
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

	// Vérifier que l'item a été ajouté correctement
	if item, exists := handler.items["test1"]; !exists {
		t.Error("Item non trouvé dans la map")
	} else {
		if item.Name != "Test Service" {
			t.Errorf("Nom de l'item incorrect, obtenu: %s, attendu: %s", item.Name, "Test Service")
		}
		if item.Url != "http://test.com" {
			t.Errorf("URL de l'item incorrect, obtenu: %s, attendu: %s", item.Url, "http://test.com")
		}
		if item.Service != "apps" {
			t.Errorf("Service de l'item incorrect, obtenu: %s, attendu: %s", item.Service, "apps")
		}
	}

	// Test 4: Vérifier que le service a été créé
	if len(handler.config.Services) != 1 {
		t.Errorf("Nombre de services incorrect, obtenu: %d, attendu: 1", len(handler.config.Services))
	}

	if len(handler.config.Services) > 0 {
		service := handler.config.Services[0]
		if service.Name != "apps" {
			t.Errorf("Nom du service incorrect, obtenu: %s, attendu: %s", service.Name, "apps")
		}
		if len(service.Items) != 1 {
			t.Errorf("Nombre d'items dans le service incorrect, obtenu: %d, attendu: 1", len(service.Items))
		}
	}

	// Test 5: Vérifier que le fichier a été créé
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("Le fichier de configuration n'a pas été créé")
	}
} 