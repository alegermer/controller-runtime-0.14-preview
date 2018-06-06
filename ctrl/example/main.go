/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"log"

	"github.com/kubernetes-sigs/kubebuilder/pkg/client"
	"github.com/kubernetes-sigs/kubebuilder/pkg/config"
	"github.com/kubernetes-sigs/kubebuilder/pkg/ctrl"
	"github.com/kubernetes-sigs/kubebuilder/pkg/ctrl/eventhandler"
	"github.com/kubernetes-sigs/kubebuilder/pkg/ctrl/inject"
	"github.com/kubernetes-sigs/kubebuilder/pkg/ctrl/reconcile"
	"github.com/kubernetes-sigs/kubebuilder/pkg/ctrl/source"
	"github.com/kubernetes-sigs/kubebuilder/pkg/signals"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func main() {
	flag.Parse()

	// Create the ControllerManager and Controller
	cm := ctrl.ControllerManager{Config: config.GetConfigOrDie()}
	c := &ctrl.Controller{Reconcile: &ReconcileReplicaSet{}}

	// Watch Pods and ReplicaSets
	cm.AddController(c, func() {
		c.Watch(
			&source.KindSource{Type: &appsv1.ReplicaSet{}}, // Watch ReplicaSets
			&eventhandler.EnqueueHandler{})                 // Enqueue ReplicaSet object key
		c.Watch(
			&source.KindSource{Type: &corev1.Pod{}}, // Watch Pods
			&eventhandler.EnqueueOwnerHandler{ // Enqueue Owning ReplicaSet object key
				OwnerType:    &appsv1.ReplicaSet{},
				IsController: true,
			})
	})

	// Start the Controllers and block until we get a shutdown signal
	cm.Start(signals.SetupSignalHandler())
}

// ReconcileReplicaSet reconciles ReplicaSets
type ReconcileReplicaSet struct {
	client client.Interface
}

// Implement inject.Client so the Controller can inject a client
var _ inject.Client = &ReconcileReplicaSet{}

func (r *ReconcileReplicaSet) InjectClient(c client.Interface) { r.client = c }

// Implement reconcile.Reconcile so the Controller can reconcile objects
var _ reconcile.Reconcile = &ReconcileReplicaSet{}

func (r *ReconcileReplicaSet) Reconcile(request reconcile.ReconcileRequest) (reconcile.ReconcileResult, error) {

	// Fetch the ReplicaSet from the cache
	rs := &appsv1.ReplicaSet{}
	err := r.client.Get(context.TODO(), request.NamespacedName, rs)
	if err != nil {
		log.Printf("Could not fetch ReplicaSet %v for %+v\n", err, request)
		return reconcile.ReconcileResult{}, err
	}

	// Print the ReplicaSet
	log.Printf("ReplicaSet Name %s Namespace %s, Pod Name: %s\n",
		rs.Name, rs.Namespace, rs.Spec.Template.Spec.Containers[0].Name)
	return reconcile.ReconcileResult{}, nil
}
