package main

import (
    "context"
    "log"
    "github.com/docker/docker/client"
    "homer-traefik/internal/watcher"
    "homer-traefik/internal/homer"
)

func main() {
    ctx := context.Background()
    
    // Création du client Docker
    dockerClient, err := client.NewClientWithOpts(client.FromEnv)
    if err != nil {
        log.Fatalf("Erreur lors de la création du client Docker: %v", err)
    }
    
    // Création du watcher
    watcher := watcher.NewDockerWatcher(dockerClient)
    
    // Création et ajout du handler Homer
    homerHandler := homer.NewConfigHandler()
    watcher.AddHandler(homerHandler)
    
    log.Println("Watcher started")
    watcher.Start(ctx)
} 