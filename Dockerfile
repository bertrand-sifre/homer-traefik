# Étape de build
FROM golang:1.23.4-alpine3.21 AS builder

WORKDIR /app

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le code source
COPY . .

# Compiler l'application
RUN CGO_ENABLED=0 GOOS=linux go build -o homer-traefik

# Étape finale avec une image minimale
FROM alpine:3.21

WORKDIR /app

# Copier l'exécutable depuis l'étape de build
COPY --from=builder /app/homer-traefik .

# Créer un volume pour la configuration Homer
VOLUME /app/config

# Définir le répertoire de travail comme répertoire de configuration
WORKDIR /app/config

# Exécuter l'application
CMD ["/app/homer-traefik"]