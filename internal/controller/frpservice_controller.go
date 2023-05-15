/*
Copyright 2023.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fcorev1 "github.com/Faione/frp-operator/api/v1"
)

// FrpServiceReconciler reconciles a FrpService object
type FrpServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.faione.frp,resources=frpservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.faione.frp,resources=frpservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.faione.frp,resources=frpservices/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FrpService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *FrpServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var frps fcorev1.FrpService

	if err := r.Get(ctx, req.NamespacedName, &frps); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	svc := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {

		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}

		log.Info("create frps svc")
		if err := r.Create(ctx, defaultService(&frps)); err != nil {
			return ctrl.Result{}, err
		}
	}

	endpoints := &corev1.Endpoints{}
	if err := r.Get(ctx, req.NamespacedName, endpoints); err != nil {

		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}

		log.Info("create frps endpoints")
		if err := r.Create(ctx, defaultEndpoints(svc, &frps)); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil

	}

	// update
	log.Info("todo update frps svc")

	// using OwnerReferences there is need to delete svc or endpoints

	return ctrl.Result{}, nil
}

func defaultService(frps *fcorev1.FrpService) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      frps.Name,
			Namespace: frps.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(frps, frps.GroupVersionKind()),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       frps.Spec.Port,
					TargetPort: intstr.FromInt(int(frps.Spec.Port)),
				},
			},
		},
	}
	return svc
}

func defaultEndpoints(svc *corev1.Service, frps *fcorev1.FrpService) *corev1.Endpoints {
	return &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(svc, svc.GroupVersionKind()),
			},
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: frps.Spec.Address,
					},
				},
				Ports: []corev1.EndpointPort{
					{
						Port: frps.Spec.Port,
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *FrpServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fcorev1.FrpService{}).
		Complete(r)
}
