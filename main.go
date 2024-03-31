package main

import (
	"flag"
	"fmt"
	"nf/config"
	"nf/screen"
	"path/filepath"
)

var (
	versionTag string = "dev"
)

func usage() {
	fmt.Println("NeoFrame - a frame buffer server")
	fmt.Println("Version:", versionTag)
	fmt.Println("Usage: neoframe [options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func main() {
	const tmpDir = "/tmp"
	uds := filepath.Join(tmpDir, "neoframe.sock")

	flag.BoolVar(&config.CFG.GetScreenInfo, "info", false, "Get screen size")
	flag.BoolVar(&config.CFG.Silent, "silent", false, "Silent mode")
	flag.BoolVar(&config.CFG.ServerMode, "server", false, "Server mode")
	flag.StringVar(&config.CFG.UnixDomainSocket, "uds", uds, "Unix domain socket")

	flag.Usage = usage

	flag.Parse()

	switch {
	case config.CFG.GetScreenInfo:
		width, height := screen.GetScreenSize()
		if config.CFG.Silent {
			fmt.Printf("%d %d\n", width, height)
			return
		}
		fmt.Printf("NeoFrame %s\n", versionTag)
		fmt.Printf("Screen size:\n\t%d\tpx width\n\t%d\tpx height\n", width, height)
		fmt.Printf("Unix domain socket:\n\t%s\n", config.CFG.UnixDomainSocket)
		return
	case config.CFG.ServerMode:
		if !config.CFG.Silent {
			fmt.Println("Server mode")
			fmt.Println("Unix domain socket:", config.CFG.UnixDomainSocket)
		}
	default:
		usage()
	}
}
