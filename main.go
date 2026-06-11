package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	fmt.Println("ID as a Service")

	exitWithError := atomic.Bool{}
	wg := sync.WaitGroup{}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	wg.Go(func() { <-ctx.Done(); stop() }) // Default signal behaviour after first catch.

	//
	// Init HTTP server.

	server := &http.Server{
		Addr:        ":8080",
		Handler:     http.HandlerFunc(handleRequest),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	//
	// Run server.

	wg.Go(func() {
		defer stop()

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Println("server stopped:", err)
			exitWithError.Store(true)
		}
	})

	//
	// Stop server.

	wg.Go(func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			fmt.Println("failed to shut down server:", err)
			exitWithError.Store(true)
		}
	})

	//
	// Exit program.

	wg.Wait()
	if exitWithError.Load() {
		os.Exit(1)
	}
}

// Stores the last assigned ID as timestamp where one unit is equal to 1 ns.
// "Overflows" on 2262-04-11 23:47:16.854775807 UTC.
var usedID atomic.Int64

func handleRequest(w http.ResponseWriter, _ *http.Request) {
	prevID := usedID.Load()
	nextID := max(prevID+1, time.Now().UnixNano())

	if !usedID.CompareAndSwap(prevID, nextID) {
		// CAS failed. Thus, the ID was updated just now (between `Load()` and
		// `CompareAndSwap()`). We can therefore assume that the ID's timestamp
		// is within an acceptable range and we can just increase it by 1.
		nextID = usedID.Add(1)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strconv.FormatInt(nextID, 10)))
}
