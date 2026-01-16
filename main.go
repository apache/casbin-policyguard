package main

import (
	"flag"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	policyv1alpha1 "github.com/casbin/policywall/api/v1alpha1"
	webhookpkg "github.com/casbin/policywall/pkg/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(policyv1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool
	var certDir string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&certDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "Directory containing TLS certificates")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "policywall.casbin.org",
		WebhookServer: webhook.NewServer(webhook.Options{
			Port:    9443,
			CertDir: certDir,
		}),
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create policy enforcer with cached reader for better performance
	enforcer, err := webhookpkg.NewPolicyEnforcer(mgr.GetClient(), mgr.GetCache())
	if err != nil {
		setupLog.Error(err, "unable to create policy enforcer")
		os.Exit(1)
	}

	// Load policies initially
	ctx := ctrl.SetupSignalHandler()
	if err := enforcer.LoadPolicies(ctx); err != nil {
		setupLog.Error(err, "unable to load policies")
		os.Exit(1)
	}

	// Create webhook handlers
	mutatingWebhook := webhookpkg.NewMutatingWebhook(mgr.GetClient(), enforcer)
	validatingWebhook := webhookpkg.NewValidatingWebhook(mgr.GetClient(), enforcer)

	// Register mutating webhook
	mgr.GetWebhookServer().Register("/mutate", &webhook.Admission{
		Handler: mutatingWebhook,
	})

	// Register validating webhook
	mgr.GetWebhookServer().Register("/validate", &webhook.Admission{
		Handler: validatingWebhook,
	})

	// Setup health checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// Index pods by namespace for efficient queries
	if err := mgr.GetFieldIndexer().IndexField(ctx, &corev1.Pod{}, "metadata.namespace", func(rawObj client.Object) []string {
		pod := rawObj.(*corev1.Pod)
		return []string{pod.Namespace}
	}); err != nil {
		setupLog.Error(err, "unable to create field indexer")
		os.Exit(1)
	}

	setupLog.Info("starting webhook server")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
