package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/satoshi-u/go-microservices/handlers"
)

// go mod init github.com/satoshi-u/go-microservices
// go run main.go
// curl -v  localhost:9090 -d sarthak
// GET     -> curl localhost:9090 | jq
// POST    -> curl localhost:9090 -d '{"name": "Indian Tea", "description": "nice cup of tea", "price": 3.14, "sku": "prod-bev-003"}'| jq
// PUT   	 -> curl localhost:9090/1 -XPUT -d '{"name": "Cappuccino", "description": "steamed milk foam", "price": 5.00, "sku": "prod-bev-001"}'| jq
// POST(2) -> curl localhost:9090 -d '{"name": "coffee $1", "description": "cheap coffee", "price": 1.00, "sku": "prod-bev-004"}'| jq
func main() {
	// logger dependency injection
	l := log.New(os.Stdout, "product-api", log.LstdFlags)
	// handler instantiate
	// hh := handlers.NewHello(l)
	ph := handlers.NewProduct(l)

	// new std lib mux : create mux and register handlers
	// sm := http.NewServeMux()
	// sm.Handle("/", hh)
	// sm.Handle("/", ph)

	// gorilla mux : create mux and register handlers
	sm := mux.NewRouter()
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/", ph.GetProducts)

	putRouter := sm.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/{id}", ph.UpdateProducts)
	putRouter.Use(ph.MiddlewareValidateProduct)

	postRouter := sm.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/", ph.AddProducts)
	postRouter.Use(ph.MiddlewareValidateProduct)

	// new server- address, handler, tls, timeouts
	s := &http.Server{
		Addr:         ":9090",
		Handler:      sm,
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	// start server listen as a non-blocking separate go routine
	go func() {
		log.Printf("Started http server at 9090...")
		err := s.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// graceful shutdown with os signal -> set signal notification on our sig channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// graceful shutdown when recieve input in sigChan, blocking operation
	sig := <-sigChan
	log.Println("Recieved terminate in sigChan, initiating graceful shutdown... sig:", sig)
	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// Even though ctx will be expired, it is good practice to call its
	// cancellation function in any case. Failure to do so may keep the
	// context and its parent alive longer than necessary.
	defer cancel()
	s.Shutdown(tc)
}
