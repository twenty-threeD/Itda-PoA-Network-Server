// Command itdad is a placeholder entrypoint for the Itda PoA node.
//
// It only parses the command line and blocks until the process is signalled,
// which is enough for the deploy pipeline and the systemd unit to be exercised
// end to end. The real node implementation replaces this.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	home := fs.String("home", defaultHome(), "node home directory")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(2)
	}

	switch cmd {
	case "start":
		start(*home)
	case "version":
		fmt.Println("itdad (placeholder build)")
	default:
		log.Printf("unknown command %q", cmd)
		usage()
		os.Exit(2)
	}
}

func start(home string) {
	log.Printf("itdad starting, home=%s", home)

	if err := os.MkdirAll(home, 0o750); err != nil {
		log.Fatalf("cannot create home directory: %v", err)
	}

	// systemd reads this line to confirm the unit came up cleanly.
	log.Print("itdad ready (placeholder: no consensus engine running)")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("received %s, shutting down", <-sig)
}

func defaultHome() string {
	if dir, err := os.UserHomeDir(); err == nil {
		return dir + "/.itda"
	}
	return ".itda"
}

func usage() {
	log.Print("usage: itdad <start|version> [--home dir]")
}
