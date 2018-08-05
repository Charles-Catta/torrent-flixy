package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/gin-gonic/gin"
)

const (
	assetsFolder = "../../assets"
	portenv      = "PORT"
	Kilobit      = 1 << 10
	Megabit      = Kilobit * 1000
)

// API implements the web API which communicates with the torrent engine
type API struct {
	Router        *gin.Engine
	server        *http.Server
	torrentEngine *torrent.Client
	torrentMap    map[string]*torrent.Torrent
}

// New API, will serve on port 8080 by default
// You can define your port by setting the environment variable PORT
func New() *API {
	addr := ""
	if p := os.Getenv(portenv); len(p) > 0 {
		addr = ":" + p
	} else {
		addr = ":8080"
	}

	router := gin.Default()

	s := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	a := &API{
		Router: router,
		server: s,
	}
	a.registerRoutes()
	a.torrentEngine = createEngine()
	a.torrentMap = make(map[string]*torrent.Torrent)
	return a
}

// Serve the API over HTTP
func (a *API) Serve() {
	go a.Router.Run()

	// Exit cleanly
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting Down API Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Stop torrent engine
	a.torrentEngine.Close()
	// Stop API server
	if err := a.server.Shutdown(ctx); err != nil {
		log.Fatal("API Server Unexpectedly Shutdown:", err)
	}
	log.Println("Server exiting")
}

type addMagnetRequest struct {
	Magnet string `json:"magnet" binding:"required"`
}

type deleteRequest struct {
	ID string `json:"ID" binding:"required"`
}

func (a *API) registerRoutes() {
	a.Router.GET("/torrents", func(c *gin.Context) {
		c.JSON(200, a.torrentMap)
	})

	a.Router.GET("/torrent/:id", func(c *gin.Context) {
		id := c.Param("id")
		if t := a.torrentMap[id]; t != nil {
			// Wait to have torrent info before requesting stream
			if t.Info() == nil {
				<-t.GotInfo()
			}
			r := t.NewReader()
			defer r.Close()
			r.SetResponsive()
			http.ServeContent(c.Writer, c.Request, t.String(), time.Now(), r)
		} else {
			c.String(404, "Torrent not found")
		}
	})

	a.Router.GET("/torrent/:id/stats", func(c *gin.Context) {
		id := c.Param("id")
		if t := a.torrentMap[id]; t != nil {
			c.JSON(200, t.Stats())
		} else {
			c.String(404, "Torrent not found")
		}
	})

	a.Router.GET("/torrent/:id/metadata", func(c *gin.Context) {
		id := c.Param("id")
		if t := a.torrentMap[id]; t != nil {
			c.JSON(200, t.Metainfo())
		} else {
			c.String(404, "Torrent not found")
		}
	})

	a.Router.POST("/torrent", func(c *gin.Context) {
		var request addMagnetRequest
		c.BindJSON(&request)
		t, err := a.torrentEngine.AddMagnet(request.Magnet)
		if err != nil {
			c.String(500, "An error occured while adding the torrent "+err.Error())
		} else {
			a.torrentMap[t.String()] = t
			go func() {
				<-t.GotInfo()
				t.DownloadAll()
			}()
			c.String(200, "Torrent added %v", t.String())
		}
	})

	a.Router.DELETE("/torrent", func(c *gin.Context) {
		var request deleteRequest
		c.BindJSON(&request)
		if t := a.torrentMap[request.ID]; t != nil {
			var removalError error
			for _, file := range t.Files() {
				fmt.Println(datadir + file.Path())
				err := os.Remove(datadir + file.Path())
				if os.IsNotExist(err) {
					removalError = nil
				} else {
					removalError = err
				}
			}
			if removalError != nil {
				c.String(500, "An error occured while removing the torrent files "+removalError.Error())
			} else {
				delete(a.torrentMap, request.ID)
				c.String(200, t.String()+" deleted")
			}
		} else {
			c.String(404, "Torrent not found")
		}
	})
}

// type stats struct {
// 	BytesWritten uint8 `json:"BytesWritten"`,

// }

// func getJsonableTorrentStats(stats torrent.TorrentStats) {

// }
