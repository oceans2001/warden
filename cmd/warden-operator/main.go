package main

import (
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	wardenv1 "github.com/oceans2001/warden/internal/apis/v1"
	"github.com/oceans2001/warden/internal/controller"
	"github.com/oceans2001/warden/internal/k8s"
)

func main() {
	ctrl.SetLogger(zap.New())

	clientset, err := k8s.NewClient()
	if err != nil {
		ctrl.Log.Error(err, "unable to create clientset")
		os.Exit(1)
	}

	dynamicClient, err := k8s.NewDynamicClient()
	if err != nil {
		ctrl.Log.Error(err, "unable to create dynamic client")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		ctrl.Log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := wardenv1.AddToScheme(mgr.GetScheme()); err != nil {
		ctrl.Log.Error(err, "unable to add scheme")
		os.Exit(1)
	}

	reconciler := &controller.AgentSandboxReconciler{
		Client:        mgr.GetClient(),
		Clientset:     clientset,
		DynamicClient: dynamicClient,
	}

	if err := reconciler.SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to setup controller")
		os.Exit(1)
	}

	ctrl.Log.Info("starting warden-operator")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctrl.Log.Error(err, "manager exited with error")
		os.Exit(1)
	}
}
