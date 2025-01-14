package main

import (
    "context"
    "log"
    "github.com/docker/docker/client"
    "homer-traefik/internal/watcher"
)

func main() {
    ctx := context.Background()
    
    // Création du client Docker
    dockerClient, err := client.NewClientWithOpts(client.FromEnv)
    if err != nil {
        log.Fatalf("Erreur lors de la création du client Docker: %v", err)
    }
    
    // Démarrage du watcher
		log.Println("Watcher started")
    watcher := watcher.NewDockerWatcher(dockerClient)
    watcher.Start(ctx)
} 