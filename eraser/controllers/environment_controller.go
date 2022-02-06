/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"

	learningk8scrdsv1beta2 "github.com/igarridot/learning-k8s-crds/eraser/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// EnvironmentReconciler reconciles a Environment object
type EnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=learning-k8s-crds.learning-k8s-crds,resources=environments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=learning-k8s-crds.learning-k8s-crds,resources=environments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=learning-k8s-crds.learning-k8s-crds,resources=environments/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources="namespaces",verbs="*"

func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var environment learningk8scrdsv1beta2.Environment
	if err := r.Get(ctx, req.NamespacedName, &environment); err != nil {
		fmt.Println(err, "unable to fetch environment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var childnamespaces corev1.NamespaceList
	if err := r.List(ctx, &childnamespaces, client.InNamespace(req.Namespace), client.MatchingFields{namespaceOwnerKey: req.Name}); err != nil {
		fmt.Println(err, "unable to list child namespaces")
		return ctrl.Result{}, err
	}

	constructNamespaceForEnvironment := func(environment *learningk8scrdsv1beta2.Environment) (*corev1.Namespace, error) {

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: environment.Name,
			},
		}

		return namespace, nil
	}

	namespace, err := constructNamespaceForEnvironment(&environment)
	if err != nil {
		fmt.Println(err, "unable to construct namespace from template")
		return ctrl.Result{}, nil
	}
	if err := r.Create(ctx, namespace); err != nil {
		fmt.Println(err, "unable to create namespace for environment", "namespace", namespace)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

var (
	namespaceOwnerKey = ".metadata.controller"
	apiGVStr          = learningk8scrdsv1beta2.GroupVersion.String()
)

func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Namespace{}, namespaceOwnerKey, func(rawObj client.Object) []string {
		namespace := rawObj.(*corev1.Namespace)
		owner := metav1.GetControllerOf(namespace)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Environment" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&learningk8scrdsv1beta2.Environment{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}
