package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	// Packages
)

var (
	flagDevice = flag.Int("device", -1, "Audio Device")
)

func main() {
	flag.Parse()

	// Create a new SDL instance
	sdl, err := NewSDL(40000)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer sdl.Close()

	// If there is no device set, then print devices and exit
	if *flagDevice < 0 {
		fmt.Println("Use -device flag to use a specific capture device:")
		for i := 0; i < sdl.NumDevices(); i++ {
			fmt.Printf("  -device %d: '%s'\n", i, sdl.DeviceName(i))
		}
		os.Exit(0)
	}

	// Open device
	if err := sdl.Open(*flagDevice); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Repeat until cancelled
	ctx := ContextWithCancel(os.Interrupt)
	fmt.Println("[speak now]")

	if err := sdl.Capture(ctx, func() {
		fmt.Println("Captured", len(sdl.Bytes()))
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func ContextWithCancel(sigs ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, sigs...)
		<-c
		cancel()
	}()
	return ctx
}
