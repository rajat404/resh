package signalhandler

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func sendSignals(sig os.Signal, subscribers []chan os.Signal, done chan string) {
	for _, sub := range subscribers {
		sub <- sig
	}
	chanCount := len(subscribers)
	start := time.Now()
	delay := time.Millisecond * 100
	timeout := time.Millisecond * 2000

	for {
		select {
		case _ = <-done:
			chanCount--
			if chanCount == 0 {
				log.Println("signalhandler: All boxes shut down successfully")
				return
			}
		default:
			time.Sleep(delay)
		}
		if time.Since(start) > timeout {
			log.Println("signalhandler: Timouted while waiting for proper shutdown - " + strconv.Itoa(chanCount) + " boxes are up after " + timeout.String())
			return
		}
	}
}

// Run catches and handles signals
func Run(subscribers []chan os.Signal, done chan string, server *http.Server) {
	signals := make(chan os.Signal, 1)

	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	sig := <-signals
	log.Println("signalhandler: Got signal " + sig.String())

	log.Println("signalhandler: Sending signals to Subscribers")
	sendSignals(sig, subscribers, done)

	log.Println("signalhandler: Shutting down the server")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}
}
