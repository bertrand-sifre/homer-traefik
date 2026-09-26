package homer

import (
	"math/rand"
	"os"
	"testing"

	"homer-traefik/internal/watcher"
)

// Labels réels de la stack : des routers dont le nom en contient un autre
// (trackarr / trackarr-discovery, navidrome / navidrome-decouvertes, sonarr / sonarr-api)
// et des items dont aucun router ne porte le nom (torrent, slskd).
var serverLabels = []watcher.DockerEvent{
	{Label: "homer.items.trackarr.url", Value: "https://trackarr.bertrand-sifre.com"},
	{Label: "homer.items.trackarr.name", Value: "Trackarr"},
	{Label: "homer.items.trackarr.service", Value: "warez"},
	{Label: "homer.items.trackarr-discovery.url", Value: "https://trackarr-discovery.bertrand-sifre.com"},
	{Label: "homer.items.trackarr-discovery.name", Value: "Trackarr (découvertes)"},
	{Label: "homer.items.trackarr-discovery.service", Value: "warez"},
	{Label: "homer.items.navidrome.url", Value: "https://navidrome.bertrand-sifre.com"},
	{Label: "homer.items.navidrome.name", Value: "Navidrome"},
	{Label: "homer.items.navidrome.service", Value: "mediacenter"},
	{Label: "homer.items.navidrome-decouvertes.url", Value: "https://navidrome-decouvertes.bertrand-sifre.com"},
	{Label: "homer.items.navidrome-decouvertes.name", Value: "Navidrome (Daily Mix)"},
	{Label: "homer.items.navidrome-decouvertes.service", Value: "mediacenter"},
	{Label: "homer.items.sonarr.url", Value: "https://sonarr.bertrand-sifre.com"},
	{Label: "homer.items.sonarr-api.url", Value: "https://sonarr-api.bertrand-sifre.com"},
	{Label: "homer.items.sonarr.name", Value: "Sonarr"},
	{Label: "homer.items.sonarr.service", Value: "warez"},
	{Label: "homer.items.gluetun-torrent.url", Value: "https://torrent.bertrand-sifre.com"},
	{Label: "homer.items.torrent.name", Value: "qbittorrent"},
	{Label: "homer.items.torrent.service", Value: "warez"},
	{Label: "homer.items.gluetun-slskd.url", Value: "https://slskd.bertrand-sifre.com"},
	{Label: "homer.items.slskd.name", Value: "Slskd"},
	{Label: "homer.items.slskd.service", Value: "warez"},
}

var expectedUrls = map[string]string{
	"Trackarr":               "https://trackarr.bertrand-sifre.com",
	"Trackarr (découvertes)": "https://trackarr-discovery.bertrand-sifre.com",
	"Navidrome":              "https://navidrome.bertrand-sifre.com",
	"Navidrome (Daily Mix)":  "https://navidrome-decouvertes.bertrand-sifre.com",
	"Sonarr":                 "https://sonarr.bertrand-sifre.com",
	"qbittorrent":            "https://torrent.bertrand-sifre.com",
	"Slskd":                  "https://slskd.bertrand-sifre.com",
}

// L'ordre des labels d'un conteneur est celui d'une map Go, donc aléatoire : le résultat
// ne doit pas en dépendre. L'ancienne fusion par sous-chaîne donnait à « trackarr »
// l'URL de « trackarr-discovery » (ou l'inverse) selon l'ordre d'arrivée.
func TestRouterUrlsDoNotLeakBetweenSimilarNames(t *testing.T) {
	tmpFile := "test_router_urls.yaml"
	defer os.Remove(tmpFile)

	rng := rand.New(rand.NewSource(1))
	for run := 0; run < 200; run++ {
		labels := append([]watcher.DockerEvent(nil), serverLabels...)
		rng.Shuffle(len(labels), func(i, j int) { labels[i], labels[j] = labels[j], labels[i] })

		handler := NewConfigHandler()
		handler.filePath = tmpFile
		for _, event := range labels {
			handler.HandleEvent(event)
		}

		got := map[string]string{}
		for _, service := range handler.config.Services {
			for _, item := range service.Items {
				got[item.Name] = item.Url
			}
		}
		if len(got) != len(expectedUrls) {
			t.Fatalf("run %d: %d items, attendu %d : %v", run, len(got), len(expectedUrls), got)
		}
		for name, want := range expectedUrls {
			if got[name] != want {
				t.Fatalf("run %d: %s → %q, attendu %q", run, name, got[name], want)
			}
		}
	}
}
