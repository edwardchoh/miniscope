package main

import (
	"flag"

	"github.com/edwardchoh/miniscope/cscope"
	"github.com/go-playground/log"
	"github.com/go-playground/log/handlers/console"
)

func main() {
	listenAddr := flag.String("listen", ":8080", "Listen Address (default *:8080)")
	cscopeOutPath := flag.String("path", ".", "Specify path to cscope.out")
	assetsPath := flag.String("assets", "", "Debug mode to load assets from directory")
	flag.Parse()

	cLog := console.New()
	log.RegisterHandler(cLog, log.AllLevels...)
	h, err := cscope.NewHttpServer(*listenAddr, *cscopeOutPath, *assetsPath)

	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Listening on ", *listenAddr, " using database at ", *cscopeOutPath)
	if err = h.ListenAndServe(); err != nil {
		log.Error(err)
		return
	}
}
