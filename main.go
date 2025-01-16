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
    
    // Create Docker client
    dockerClient, err := client.NewClientWithOpts(client.FromEnv)
    if err != nil {
        log.Fatalf("Error creating Docker client: %v", err)
    }
    
    // Create watcher
    watcher := watcher.NewDockerWatcher(dockerClient)
    
    // Create and add Homer handler
    homerHandler := homer.NewConfigHandler()
    watcher.AddHandler(homerHandler)
    
    log.Println("Watcher started")
    watcher.Start(ctx)
} 