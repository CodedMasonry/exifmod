package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/crypto/acme/autocert"
)

func handlerMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, TLS user! Your config: %+v", r.TLS)
	})

	return mux
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	email := os.Getenv("EXIFMOD_EMAIL")
	var hosts []string
	err = json.Unmarshal([]byte(os.Getenv("EXIFMOD_HOSTS")), &hosts)
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	log.Printf("Hosts: %+v", hosts)
	certManager := &autocert.Manager{
		Cache:      autocert.DirCache("exifmod_certs"),
		Prompt:     autocert.AcceptTOS,
		Email:      email,
		HostPolicy: autocert.HostWhitelist(hosts...),
	}
	handler := handlerMux()

	quic_server := &http3.Server{
		Addr:    ":https",
		Handler: handler,
		TLSConfig: http3.ConfigureTLSConfig(&tls.Config{
			GetCertificate: certManager.GetCertificate,
		}),
	}

	// Requires special handler to set QUIC Alt-Svc headers
	http_server := &http.Server{
		Addr: ":https",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			quic_server.SetQUICHeaders(w.Header())
			handler.ServeHTTP(w, r)
		}),
		TLSConfig: certManager.TLSConfig(),
	}

	log.Printf("Starting Servers")
	hErr := make(chan error, 1)
	qErr := make(chan error, 1)
	go func() {
		hErr <- http_server.ListenAndServeTLS("", "")
	}()
	go func() {
		qErr <- quic_server.ListenAndServe()
	}()

	select {
	case err := <-hErr:
		log.Fatalf("HTTP Server Error: %v", err)
	case err := <-qErr:
		log.Fatalf("QUIC Server Error: %v", err)
	}
}
