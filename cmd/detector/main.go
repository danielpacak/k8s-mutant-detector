package main

import (
	"log"

	"github.com/danielpacak/k8s-mutant-detector/pkg/controller/replicaset"

	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %s", err.Error())
	}
}

func run() error {
	// Instantiate controllers' manager
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Namespace: corev1.NamespaceDefault,
	})
	if err != nil {
		return err
	}

	// Setup a new controller to reconcile ReplicaSets
	c, err := controller.New("mutant-detector", mgr, controller.Options{
		Reconciler: replicaset.NewReconciler(mgr.GetClient()),
	})

	if err != nil {
		return nil
	}

	// Watch ReplicaSets and enqueue ReplicaSet object key
	if err := c.Watch(&source.Kind{Type: &appsv1.ReplicaSet{}}, &handler.EnqueueRequestForObject{}); err != nil {
		return err
	}

	// Watch Pods and enqueue owning ReplicaSet key
	if err := c.Watch(&source.Kind{Type: &corev1.Pod{}},
		&handler.EnqueueRequestForOwner{OwnerType: &appsv1.ReplicaSet{}, IsController: true}); err != nil {
		return err
	}

	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		return err
	}

	return nil
}
