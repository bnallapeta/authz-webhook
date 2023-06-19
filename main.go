package main

import (
	"github.com/bnallapeta/authz-webhook/pkg/webhook"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func main() {
	logger := zap.New(zap.UseDevMode(true))
	ctrl.SetLogger(logger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		// Port to listen for webhook requests
		Port:    9443,
		CertDir: "./cert",
	})
	if err != nil {
		logger.Error(err, "unable to start manager")
	}

	validator, err := webhook.NewUserValidator()
	if err != nil {
		logger.Error(err, "unable to initialize user validator")
	}

	hookServer := mgr.GetWebhookServer()
	hookServer.Register("/validate", &admission.Webhook{
		Handler: validator,
	})

	logger.Info("Testing if this is getting printed 0")
	logger.Info("Starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		logger.Error(err, "Error starting manager")
	}

	logger.Info("Testing if this is getting printed 1")
}
