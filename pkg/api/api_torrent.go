package api

import (
	"log"

	"github.com/anacrolix/torrent"
)

const datadir = "/tmp/torrents/"

func createEngine() *torrent.Client {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = datadir

	c, err := torrent.NewClient(cfg)
	if err != nil {
		log.Fatal("An error occured while setting up the torrent engine", err)
	}
	return c
}
