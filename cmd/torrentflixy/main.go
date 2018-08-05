package main

import "github.com/Charles-Catta/torrent-flixy/pkg/api"

func main() {
	s := api.New()
	s.Serve()
}
