package replicaset

import (
	"context"
	"fmt"
	"strconv"

	"github.com/danielpacak/k8s-mutant-detector/pkg/mutant"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	labelIsMutant          = "is-mutant"
	annotationMutantStatus = "mutant/status"
)

// reconciler reconciles ReplicaSets
type reconciler struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
}

// NewReconciler constructs a new ReplicaSet reconciler.
//
// In each cycle the reconciler calculates the mutant.Status
// of the ReplicaSets and saves it as JSON value of the mutant/state
// annotation.
func NewReconciler(client client.Client) reconcile.Reconciler {
	return &reconciler{
		client: client,
	}
}

func (r *reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch ReplicaSet from cache
	rs := &appsv1.ReplicaSet{}
	err := r.client.Get(context.TODO(), request.NamespacedName, rs)
	if errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch ReplicaSet: %+v", err)
	}

	// List pods managed by ReplicaSet
	managedPods, err := r.getPodsControllerBy(rs)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch Pods managed by ReplicaSet: %+v", err)
	}

	fmt.Println("managed pods", len(managedPods))
	// Create struct to determine whether ReplicaSet is mutant or not
	mutantStatus := mutant.GetStatus(managedPods)

	// Update ReplicaSet with label and annotation so other tools can
	// see if it's mutant or not
	err = r.updateReplicaSet(rs, mutantStatus)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not update ReplicaSet: %+v", err)
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) getPodsControllerBy(rs *appsv1.ReplicaSet) ([]corev1.Pod, error) {
	selectorLabels, err := metav1.LabelSelectorAsMap(rs.Spec.Selector)
	if err != nil {
		return nil, err
	}

	var list = &corev1.PodList{}
	err = r.client.List(context.TODO(), list, client.MatchingLabels(selectorLabels))
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (r *reconciler) updateReplicaSet(rs *appsv1.ReplicaSet, status mutant.Status) error {
	mutantStatusAsJson, err := status.AsJson()
	if err != nil {
		return err
	}

	if rs.Annotations == nil {
		rs.Annotations = map[string]string{}
	}
	rs.Annotations[annotationMutantStatus] = mutantStatusAsJson

	if rs.Labels == nil {
		rs.Labels = map[string]string{}
	}

	rs.Labels[labelIsMutant] = strconv.FormatBool(status.IsMutant())

	err = r.client.Update(context.TODO(), rs)
	if err != nil {
		return err
	}

	return nil
}
