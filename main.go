package main

import (
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"net"
	"nf/config"
	"nf/screen"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func shutdown(ret int) {
	if config.CFG.ServerMode {
		if !config.CFG.Silent {
			fmt.Println("Shutdown server")
		}
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

		fmt.Println("Recebido:", string(buf[:n]))

		// remove duplicate spaces
		b := strings.Join(strings.Fields(string(buf[:n])), " ")

		cmd := strings.Split(b, " ")

		switch cmd[0] {
		case "shutdown":
			_, err = conn.Write([]byte("shutdown server"))
			if err != nil {
				fmt.Println("failed to write:", err)
				shutdown(1)
			}
			shutdown(0)
		case "test":
			screen.SetBackgroudImage("./awake.png")
		case "image":
			if len(cmd) != 2 {
				e := "image command requires a file name"
				conn.Write([]byte(e))
				fmt.Println(e)
				continue
			}
			file := cmd[1]
			_, err := os.Stat(file)
			if err != nil {
				if err == os.ErrNotExist {
					e := fmt.Sprintf("File %s does not exist", file)
					log.Println(e)
					conn.Write([]byte(e))
					continue
				}
			}
			screen.SetBackgroudImage(file)
		case "image_at":
			if len(cmd) != 4 {
				e := "image_at command requires a file name, x and y"
				conn.Write([]byte(e))
				fmt.Println(e)
				continue
			}
			file := cmd[1]
			_, err := os.Stat(file)
			if err != nil {
				if err == os.ErrNotExist {
					e := fmt.Sprintf("File %s does not exist", file)
					log.Println(e)
					conn.Write([]byte(e))
					continue
				}
			}
			x := cmd[2]
			y := cmd[3]
			xa, err := strconv.Atoi(x)
			if err != nil {
				e := fmt.Sprintf("Invalid x value: %s", x)
				log.Println(e)
				conn.Write([]byte(e))
				continue
			}
			ya, err := strconv.Atoi(y)
			if err != nil {
				e := fmt.Sprintf("Invalid y value: %s", y)
				log.Println(e)
				conn.Write([]byte(e))
				continue
			}
			err = screen.SetBackgroudImageAt(file, xa, ya)
			if err != nil {
				e := fmt.Sprintf("Failed to set image at %d, %d: %s", xa, ya, err)
				log.Println(e)
				conn.Write([]byte(e))
				continue
			}

		case "clear", "cls", "clean":
			screen.Clean()
		default:
			e := fmt.Sprintf("Unknown command: %s", buf[:n])
			if !config.CFG.Silent {
				fmt.Printf(e)
			}
			_, err = conn.Write([]byte(e))
			if err != nil {
				fmt.Println("failed to write:", err)
				shutdown(1)
			}
			continue
		}

		_, err = conn.Write([]byte("OK"))
		if err != nil {
			fmt.Println("failed to write:", err)
			shutdown(1)
		}
	}
}

func UDSClient() net.Conn {
	conn, err := net.Dial("unix", config.CFG.UnixDomainSocket)
	if err != nil {
		fmt.Println("failed to dial:", err)
		shutdown(1)
	}
	return conn
}

func UDSClientSend(conn net.Conn, msg string) string {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("failed to write:", err)
		shutdown(1)
	}
	// read response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("failed to read:", err)
		shutdown(1)
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
		shutdown(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if !config.CFG.Silent {
				fmt.Println("failed to accept:", err)
			}
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

func main() {
	const tmpDir = "/tmp"
	uds := filepath.Join(tmpDir, "neoframe.sock")
	var cmd string

	flag.BoolVar(&config.CFG.GetScreenInfo, "info", false, "Get screen size")
	flag.BoolVar(&config.CFG.Silent, "silent", false, "Silent mode")
	flag.BoolVar(&config.CFG.ServerMode, "server", false, "Server mode")
	flag.StringVar(&config.CFG.UnixDomainSocket, "uds", uds, "Unix domain socket")
	flag.StringVar(&cmd, "cmd", "", "Command to send to server")

	flag.Usage = usage

	flag.Parse()

	defer func() {
		if config.CFG.ServerMode {
			os.Remove(uds)
		}
	}()

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
		_, err := os.Stat(config.CFG.UnixDomainSocket)
		if err == nil {
			fmt.Printf("Unix domain socket %s already exists, remove the file first\n", config.CFG.UnixDomainSocket)
			shutdown(1)
		}
		if !config.CFG.Silent {
			fmt.Println("Server mode")
			fmt.Println("Unix domain socket:", config.CFG.UnixDomainSocket)
		}
		go UDSListener()
		screen.RunApp()
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
