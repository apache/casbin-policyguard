package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/casbin/policywall/api/v1alpha1"
	"github.com/casbin/policywall/pkg/controller"
	"github.com/casbin/policywall/pkg/webhook"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
}

func main() {
	var (
		metricsAddr          string
		healthProbeAddr      string
		webhookPort          int
		webhookCertDir       string
		enableLeaderElection bool
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&healthProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.IntVar(&webhookPort, "webhook-port", 9443, "The port the webhook server binds to.")
	flag.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/k8s-webhook-server/serving-certs", "The directory containing webhook certificates.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")

	klog.InitFlags(nil)
	flag.Parse()

	// Set up manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: healthProbeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "policywall.casbin.org",
	})
	if err != nil {
		klog.Errorf("Unable to start manager: %v", err)
		os.Exit(1)
	}

	// Create webhook server
	webhookServer := webhook.NewWebhookServer()

	// Set up controller
	if err = (&controller.AdmissionPolicyReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		WebhookServer: webhookServer,
	}).SetupWithManager(mgr); err != nil {
		klog.Errorf("Unable to create controller: %v", err)
		os.Exit(1)
	}

	// Add health check
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Errorf("Unable to set up health check: %v", err)
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.Errorf("Unable to set up ready check: %v", err)
		os.Exit(1)
	}

	// Start webhook server in a separate goroutine
	go startWebhookServer(webhookServer, webhookPort, webhookCertDir)

	// Start manager
	klog.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.Errorf("Problem running manager: %v", err)
		os.Exit(1)
	}
}

func startWebhookServer(server *webhook.WebhookServer, port int, certDir string) {
	mux := http.NewServeMux()
	mux.Handle("/validate", server)
	mux.HandleFunc("/healthz", server.HealthCheck)

	// Set up HTTPS server
	tlsCert := fmt.Sprintf("%s/tls.crt", certDir)
	tlsKey := fmt.Sprintf("%s/tls.key", certDir)

	// Check if certificates exist
	if _, err := os.Stat(tlsCert); os.IsNotExist(err) {
		klog.Warningf("Certificate file not found at %s, will wait for it", tlsCert)
		waitForCertificates(tlsCert, tlsKey)
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	klog.Infof("Starting webhook server on port %d", port)
	if err := httpServer.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
		klog.Errorf("Webhook server failed: %v", err)
		os.Exit(1)
	}
}

func waitForCertificates(certFile, keyFile string) {
	klog.Info("Waiting for certificates to be available...")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.After(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat(certFile); err == nil {
				if _, err := os.Stat(keyFile); err == nil {
					klog.Info("Certificates are now available")
					return
				}
			}
		case <-timeout:
			klog.Error("Timeout waiting for certificates")
			os.Exit(1)
		}
	}
}
