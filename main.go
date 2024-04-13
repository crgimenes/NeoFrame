package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"

	"nf/config"
	"nf/luaengine"
	"nf/neoframe"

	"github.com/ergochat/readline"
)

const (
	tmpDir = "/tmp"
)

var (
	versionTag string = "dev"
	le         *luaengine.LuaExtender
	nf         *neoframe.NeoFrame
	ac         *AppCtrl = &AppCtrl{}
)

func usage() {
	fmt.Println("NeoFrame - a frame buffer server")
	fmt.Println("Version:", versionTag)
	fmt.Println("Usage: neoframe [options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

type AppCtrl struct{}

func (ac *AppCtrl) Shutdown(ret int) {
	if config.CFG.ServerMode {
		fmt.Println("\r\nShutdown server")
		os.Remove(config.CFG.UnixDomainSocket)
	}
	os.Exit(ret)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return
			}
			fmt.Println("failed to read:", err)
			return
		}

		cmd := string(buf[:n])

		err = le.Run(cmd)
		if err != nil {
			conn.Write([]byte(fmt.Sprintf("failed to run command: %s\n", err.Error())))
			fmt.Println("failed to run command:", err)
			continue
		}

		conn.Write([]byte("OK\n"))
	}
}

func UDSClient() net.Conn {
	conn, err := net.Dial("unix", config.CFG.UnixDomainSocket)
	if err != nil {
		fmt.Println("failed to dial:", err)
		ac.Shutdown(1)
	}
	return conn
}

func UDSClientSend(conn net.Conn, msg string) string {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("failed to write:", err)
		ac.Shutdown(1)
	}
	// read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("failed to read:", err)
		ac.Shutdown(1)
	}

	return string(buf[:n])
}

func UDSCientClose(conn net.Conn) {
	conn.Close()
}

func UDSListener() {
	listener, err := net.Listen("unix", config.CFG.UnixDomainSocket)
	if err != nil {
		fmt.Println("failed to listen:", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("failed to accept:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func RGBAImageToBytes(img *image.RGBA) []byte {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	bytes := make([]byte, 0, w*h*4)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			bytes = append(bytes, byte(r>>8), byte(g>>8), byte(b>>8), byte(a>>8))
		}
	}
	return bytes
}

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

var completer = readline.NewPrefixCompleter()

func runCMD() {
	historyFile := filepath.Join(tmpDir, "neoframe.history")
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     historyFile,
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
		Stdin:               os.Stdin,
		Stdout:              os.Stdout,
		Stderr:              os.Stderr,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()
	log.SetOutput(rl.Stderr()) // redraw the prompt correctly after log output

	for {
		line, err := rl.ReadLine()
		if err != nil {
			ac.Shutdown(1)
			return
		}

		err = le.Run(line)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func main() {
	uds := filepath.Join(tmpDir, "neoframe.sock")
	var cmd string

	flag.BoolVar(&config.CFG.GetScreenInfo, "info", false, "Get screen size")
	flag.BoolVar(&config.CFG.MousePassthrough, "mouse_passthrough", true, "Mouse passthrough")
	flag.BoolVar(&config.CFG.ServerMode, "server", false, "Server mode")
	flag.BoolVar(&config.CFG.WindowDecorated, "window_decorated", false, "Window decorated")
	flag.IntVar(&config.CFG.WindowHeight, "height", 0, "Window height")
	flag.IntVar(&config.CFG.WindowWidth, "width", 0, "Window width")
	flag.IntVar(&config.CFG.WindowX, "x", 0, "Window x position")
	flag.IntVar(&config.CFG.WindowY, "y", 0, "Window y position")
	flag.StringVar(&config.CFG.RunLuaScript, "run_file", "", "Run lua script")
	flag.StringVar(&config.CFG.UnixDomainSocket, "uds", uds, "Unix domain socket")
	flag.StringVar(&config.CFG.WindowBgColor, "bgcolor", "00000000", "Window background color (RGBA) in hex")
	flag.StringVar(&config.CFG.WindowTitle, "title", "NeoFrame", "Window title")
	flag.StringVar(&cmd, "cmd", "", "Command to send to server")

	flag.Usage = usage
	flag.Parse()

	if config.CFG.ServerMode {
		go func() {
			sc := make(chan os.Signal, 1)
			signal.Notify(sc, os.Interrupt)
			<-sc
			ac.Shutdown(0)
		}()
	}

	switch {
	case config.CFG.ServerMode:
		_, err := os.Stat(config.CFG.UnixDomainSocket)
		if err == nil {
			fmt.Printf("Unix domain socket %s already exists, remove the file first\n", config.CFG.UnixDomainSocket)
			ac.Shutdown(1)
		}
		fmt.Println("Server mode")
		fmt.Println("Unix domain socket:", config.CFG.UnixDomainSocket)

		nf = neoframe.New()
		le = luaengine.New(nf, ac)

		hasInitFile := func() bool {
			_, err = os.Stat("init.lua")
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return false
				}
				log.Fatal(err)
			}
			return true
		}()

		if hasInitFile {
			le.Proto, err = le.Compile("init.lua")
			if err != nil {
				fmt.Println("failed to compile init.lua:", err)
				ac.Shutdown(1)
			}

			err = le.InitStateWithProto()
			if err != nil {
				fmt.Println("failed to init lua state:", err)
				ac.Shutdown(1)
			}
		}

		go UDSListener()
		go runCMD()
		nf.Run()
	case cmd != "":
		conn := UDSClient()
		s := UDSClientSend(conn, cmd)
		if s != "" {
			fmt.Println(s)
		}
		UDSCientClose(conn)
	default:
		usage()
	}
}
