package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"stache/internal/engine"
	"stache/internal/server"
	"syscall"
	"time"
)

func main() { 
	socketPath := flag.String("socket", "/tmp/stache.sock", "Path to Unix Domain Socket")
	maxMem := flag.Int64("max-memory", 256*1024*1024, "Max memory in bytes (Default - 256MB)")
	maxItem := flag.Int64("max-item-size", 4*1024*1024, "Max single item size in bytes (Default - 4MB)")

	flag.Parse()

	log.Printf("Starting stache daemon (socket: %s, max-memory: %dMB)", *socketPath, *maxMem)

	cacheEngine := engine.New(
		engine.LimitsConfig{
			MaxMemoryBytes: *maxMem,
			MaxItemBytes: *maxItem,
			CleanUpInterval: 1 * time.Minute,
		},
	)
	defer cacheEngine.Close()

	s := server.New(
		*socketPath,
		cacheEngine,
	)
	if err := s.Start(); err != nil { 
		log.Fatalf("Failed to start the server: %v", err)
	}
	log.Printf("stache daemon is ready for connections")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	log.Println("Shutting the stache daemon now...")

	if err := s.Stop(); err != nil { 
		log.Printf("Can't shut the daemon: %v", err)
	}

	log.Println("Daemon stopped succesfully")
}