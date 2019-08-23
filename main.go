package main

import (
	"crypto/tls"
	"flag"
	"github.com/golang/glog"
	"time"
)

const (
	gracefulShutdownDefault = 10 * time.Second
)

func main() {

	var port int
	var shutdownGracePeriod time.Duration
	var tlsCertFilePath string
	var tlsKeyFilePath string

	var err error

	flag.IntVar(&port, "port", 8443, "Server tls port")
	flag.DurationVar(&shutdownGracePeriod, "grace-period", gracefulShutdownDefault, "Graceful shutdown time period")
	flag.StringVar(&tlsCertFilePath, "tlsCertFilePath", "certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&tlsKeyFilePath, "tlsKeyFilePath", "certs/key.pem", "File containing the x509 private key to --tlsCertFile.")

	flag.Parse()

	// Load certificates
	pair, err := tls.LoadX509KeyPair(tlsCertFilePath, tlsKeyFilePath)
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}

	// Create NewServer and add handler
	server := NewServer(port, shutdownGracePeriod)
	server.mux.HandleFunc("/validate", validateHandler)

	// Start server
	err = server.Start(pair)
	if err != nil {
		glog.Fatalf("Error: %v", err)
	}
}
