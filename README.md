# Homer Traefik Integration

⚠️ **Important Notice**

This project was created as an experiment to test [Cursor](https://cursor.sh/), an AI-powered IDE. 100% of the code (readme included) was generated through interactions with Cursor's AI assistant. I have no prior experience with Go programming, and this project serves as a demonstration of AI-assisted development capabilities.

Use this code at your own risk, as it may not follow Go best practices or contain optimal implementations.

## Description

This project automatically generates and maintains your Homer dashboard configuration based on Docker container labels. It monitors Docker events and updates the Homer configuration in real-time.

## Features

- Automatic service detection via Docker labels
- Real-time synchronization with Homer
- Traefik labels support for service URLs
- Customizable configuration through Homer labels

## Installation


### Using Docker Compose

```yml
version: '3'
services:
  homer-traefik:
    image: ghcr.io/bertrandsifre/homer-traefik:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./homer:/app/config
  homer:
    image: b4rt0/homer:latest
    ports:
      - 8080:80
    volumes:
      - ./homer:/app/config
```

### Configuration

Add labels to your Docker containers to make them appear in Homer:

```yaml
version: '3'
services:
  my-app:
    image: nginx
    labels:
      - "traefik.http.routers.my-app.rule=Host(`app.example.com`)"
      - "homer.items.my-app.name=My Application"
      - "homer.items.my-app.subtitle=A great application"
      - "homer.items.my-app.service=apps"
      - "homer.items.my-app.icon=fas fa-rocket"
```

## Supported Labels

- `homer.items.<id>.name`: Service name
- `homer.items.<id>.subtitle`: Service subtitle
- `homer.items.<id>.url`: Service URL (optional if Traefik label is present)
- `homer.items.<id>.icon`: Service icon (Font Awesome format)
- `homer.items.<id>.service`: <id-service> Service category in Homer
- `homer.services.<id-service>.name`: Service name

## Development

### Prerequisites

- Go 1.23.4 or higher
- Docker

### Local Setup

```bash
git clone https://github.com/bertrandsifre/homer-traefik.git
cd homer-traefik

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build
```

## License

[MIT](LICENSE)
