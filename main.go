package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/ergochat/readline"
)

var (
	nf        *NeoFrame
	completer = readline.NewPrefixCompleter(
		readline.PcItem("exit"),
		readline.PcItem("quit"),
		readline.PcItem("clear"),
		readline.PcItem("help"),
		readline.PcItem("passthrough"),
	)
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func runCMD() {
	historyFile := "neoframe.history"
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
			if err == readline.ErrInterrupt {
				fmt.Println("\nExiting NeoFrame...")
				os.Exit(0)
			}
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch line {
		case "exit", "quit", "q", ":q":
			fmt.Println("Exiting NeoFrame...")
			os.Exit(0)
		case "clear", "cls":
			nf.Clear()
		case "passthrough":
			nf.SetMousePassthrough(false)
			fmt.Println("Passthrough disabled.")
			fmt.Println("ESC to reenable passthrough.")
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  exit - Exit the application")
			fmt.Println("  clear - Clear the NeoFrame display")
			fmt.Println("  help - Show this help message")
		default:
			fmt.Printf("Unknown command: %q\n", line)
		}

	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	fmt.Println("Starting NeoFrame...")
	nf = &NeoFrame{}
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

	/*
		go func() {
			time.Sleep(2 * time.Second)
			nf.DrawText(150, 150, 25.0, "Hello NeoFrame!", "#FFFFFF")

			// Círculo preenchido vermelho, raio 50
			nf.DrawCircle(100, 100, 50, 0, true, "FF0000FF")

			// Círculo não preenchido azul, raio 30, espessura 3
			nf.DrawCircle(200, 200, 30, 3, false, "0000FFFF")

			// Círculo não preenchido verde, raio 20, espessura 1 (linha fina)
			nf.DrawCircle(300, 300, 20, 2, false, "00FF00FF")

		}()
	*/

	go runCMD()

	nf.Run()
}
