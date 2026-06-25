package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	fmt.Println("Starting NeoFrame...")

	nf := &NeoFrame{}
	nf.CFG = &Config{}
	nf.CFG.MousePassthrough = true
	nf.ConfigureMonitorSize()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nExiting NeoFrame...")
		os.Exit(0)
	}()

	nf.Run()
}
