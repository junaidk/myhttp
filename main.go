package main

import (
	"context"
	"flag"
	"fmt"
	"myhttp/lib"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	parallelCount := flag.Int("parallel", 10, "number of parallel requests")
	flag.Parse()

	urlList := flag.Args()

	proc := lib.NewProcessor(
		&http.Client{Timeout: time.Second * 5},
		*parallelCount,
	)

	ctx, cancel := context.WithCancel(context.Background())
	setupContextCancelHandler(cancel)
	proc.Run(ctx, urlList, os.Stdout)
}

func setupContextCancelHandler(cancel context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Ctrl+C pressed. Cancelling...")
		cancel()
		os.Exit(0)
	}()
}
