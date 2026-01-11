// Copyright 2026 The Casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	casbinpkg "github.com/casbin/policywall/pkg/casbin"
	"github.com/casbin/policywall/pkg/controller"
	"github.com/casbin/policywall/pkg/metrics"
	"github.com/casbin/policywall/pkg/webhook"
)

var (
	kubeconfig  string
	certFile    string
	keyFile     string
	port        int
	metricsPort int
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	flag.StringVar(&certFile, "tls-cert-file", "", "Path to TLS certificate file")
	flag.StringVar(&keyFile, "tls-key-file", "", "Path to TLS key file")
	flag.IntVar(&port, "port", 8443, "Webhook server port")
	flag.IntVar(&metricsPort, "metrics-port", 9090, "Metrics server port")
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	klog.Info("Starting PolicyWall admission controller")

	// Create Kubernetes client
	config, err := getKubeConfig()
	if err != nil {
		klog.Fatalf("Failed to get kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create clientset: %v", err)
	}

	// Create policy enforcer
	enforcer := casbinpkg.NewPolicyEnforcer()

	// Create metrics collector
	metricsCollector := metrics.NewCollector()

	// Start metrics server
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", metricsPort),
		Handler: metricsCollector.Handler(),
	}

	go func() {
		klog.Infof("Starting metrics server on :%d", metricsPort)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Create and start policy controller
	policyController := controller.NewPolicyController(clientset, enforcer)
	go func() {
		if err := policyController.Run(); err != nil {
			klog.Fatalf("Failed to run policy controller: %v", err)
		}
	}()

	// Create webhook server
	webhookServer := webhook.NewServer(enforcer, metricsCollector)

	// Setup HTTP handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", webhookServer.HandleAdmission)
	mux.HandleFunc("/health", webhookServer.HandleHealth)
	mux.HandleFunc("/ready", webhookServer.HandleReady)

	// Create HTTPS server
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// Start webhook server
	go func() {
		klog.Infof("Starting webhook server on :%d", port)
		if certFile != "" && keyFile != "" {
			if err := server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				klog.Fatalf("Failed to start webhook server: %v", err)
			}
		} else {
			klog.Warning("TLS cert/key not provided, running in insecure mode")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				klog.Fatalf("Failed to start webhook server: %v", err)
			}
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	klog.Info("Shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		klog.Errorf("Webhook server shutdown error: %v", err)
	}

	if err := metricsServer.Shutdown(ctx); err != nil {
		klog.Errorf("Metrics server shutdown error: %v", err)
	}

	policyController.Stop()

	klog.Info("Shutdown complete")
}

func getKubeConfig() (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
