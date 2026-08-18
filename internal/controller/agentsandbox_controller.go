package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wardenv1 "github.com/oceans2001/warden/internal/apis/v1"
	"github.com/oceans2001/warden/internal/k8s"
	"github.com/oceans2001/warden/internal/policy"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type AgentSandboxReconciler struct {
	client.Client
	Clientset     *kubernetes.Clientset
	DynamicClient dynamic.Interface
}

func (r *AgentSandboxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var sandbox wardenv1.AgentSandbox
	if err := r.Get(ctx, req.NamespacedName, &sandbox); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	command := sandbox.Spec.Command
	if len(command) == 0 {
		command = []string{"sh", "-c", "sleep 3600"}
	}

	err := k8s.CreatePod(r.Clientset, sandbox.Name, sandbox.Spec.Image, command)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}

	pol := policy.BuildFQDNPolicy(sandbox.Name, sandbox.Spec.AllowedDomains)
	err = policy.ApplyPolicy(r.DynamicClient, pol)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}

	sandbox.Status.Phase = "Ready"
	if err := r.Status().Update(ctx, &sandbox); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AgentSandboxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wardenv1.AgentSandbox{}).
		Complete(r)
}
