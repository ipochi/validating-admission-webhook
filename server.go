package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/ipochi/psp-validating-admission-webhook/webhook"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// initalize deserializers
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

// Server contains the configuration for the HTTP server
type Server struct {
	address             string
	shutdownGracePeriod time.Duration
	mux                 *http.ServeMux
}

// NewServer creates a new Server definition with an empty ServeMux
func NewServer(port int, shutdownGracePeriod time.Duration) *Server {
	return &Server{
		address:             fmt.Sprintf(":%d", port),
		shutdownGracePeriod: shutdownGracePeriod,
		mux:                 http.NewServeMux(),
	}
}

func (s *Server) Start(pair tls.Certificate) error {
	httpServer := &http.Server{
		Addr:      s.address,
		Handler:   s.mux,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
	}
	go func() {
		var err error
		err = httpServer.ListenAndServeTLS("", "")
		if err != nil {
			glog.Fatalf("Failed to start HTTPS server: %s", err.Error())
		}
	}()
	glog.Infof("HTTPS server started, listening on: %s", httpServer.Addr)
	return gracefulShutdown(httpServer, s.shutdownGracePeriod)
}

func gracefulShutdown(s *http.Server, timeout time.Duration) error {
	signals := make(chan os.Signal, 1)
	// SIGTERM is used by Kubernetes to gracefully stop pods.
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	glog.Infof("Shutdown starting with a grace period of %s", timeout)
	return s.Shutdown(ctx)
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		glog.Errorf("Invalid content type: %s. Expected `application/json`", contentType)
		http.Error(w, "Invalid content type. Expected `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	requestedAdmissionReview := &v1beta1.AdmissionReview{}
	responseAdmissionReview := &v1beta1.AdmissionReview{}

	_, _, err := deserializer.Decode(body, nil, requestedAdmissionReview)

	if err != nil {
		glog.Errorf("Cannot decode request body: %v", err)
		responseAdmissionReview.Response = webhook.ToAdmissionResponse(err.Error())
	}

	glog.Info("Coming here to validate")
	responseAdmissionReview.Response = webhook.Validate(requestedAdmissionReview)

	// Encode admissionreview response
	resp, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		glog.Errorf("Cannot encode response: %v", err)
		http.Error(w, fmt.Sprintf("Cannot encode response: %v", err), http.StatusInternalServerError)
	}

	// Send AdmissionReview Response
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Cannot write response: %v", err)
		http.Error(w, fmt.Sprintf("Cannot write response: %v", err), http.StatusInternalServerError)
	}
}
