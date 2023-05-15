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
	"bytes"
	"context"
	"fmt"

	"github.com/go-ini/ini"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fcorev1 "github.com/Faione/frp-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceProxyReconciler reconciles a ServiceProxy object
type ServiceProxyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=core.faione.frp,resources=serviceproxies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core.faione.frp,resources=serviceproxies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core.faione.frp,resources=serviceproxies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ServiceProxy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *ServiceProxyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var serviceProxy fcorev1.ServiceProxy
	if err := r.Get(ctx, req.NamespacedName, &serviceProxy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	frps := &fcorev1.FrpService{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: serviceProxy.Spec.Frps.NameSpace,
		Name:      serviceProxy.Spec.Frps.Name,
	}, frps); err != nil {
		return ctrl.Result{}, err
	}

	configMap := &corev1.ConfigMap{}
	if err := r.Get(ctx, req.NamespacedName, configMap); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}

		log.Info("create frpc configMap")
		if err := r.Create(ctx, defaultConfigMap(&serviceProxy, frps)); err != nil {
			return ctrl.Result{}, err
		}
	}

	deploy := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deploy); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}

		log.Info("create frpc deploy")
		if err := r.Create(ctx, defaultDeploy(&serviceProxy)); err != nil {
			return ctrl.Result{}, err
		}

	}

	return ctrl.Result{}, nil
}

func defaultConfigMap(sp *fcorev1.ServiceProxy, frps *fcorev1.FrpService) *corev1.ConfigMap {

	cfg := ini.Empty()

	cfg.Section("common").Key("tls_enable").SetValue("true")
	cfg.Section("common").Key("server_addr").SetValue(frps.Spec.Address)
	cfg.Section("common").Key("server_port").SetValue(fmt.Sprintf("%d", frps.Spec.Port))
	cfg.Section("common").Key("token").SetValue(frps.Spec.Token)

	for _, proxy := range sp.Spec.Proxies {
		cfg.Section(proxy.Name).Key("local_ip").SetValue(proxy.LocalIp)
		cfg.Section(proxy.Name).Key("local_port").SetValue(proxy.LocalPort)
		cfg.Section(proxy.Name).Key("remote_port").SetValue(proxy.RemotePort)
	}

	bf := new(bytes.Buffer)
	cfg.WriteTo(bf)
	config := bf.String()

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sp.Name,
			Namespace: sp.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(sp, sp.GroupVersionKind()),
			},
		},
		Data: map[string]string{"frpc.ini": config},
	}
}

func defaultDeploy(sp *fcorev1.ServiceProxy) *appsv1.Deployment {
	image := "snowdreamtech/frpc"
	mountPath := "/etc/frp/frpc.ini"

	if sp.Spec.Image != "" {
		image = sp.Spec.Image
	}

	if sp.Spec.MountPath != "" {
		mountPath = sp.Spec.MountPath
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sp.Name,
			Namespace: sp.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(sp, sp.GroupVersionKind()),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": sp.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   sp.Name,
					Labels: map[string]string{"app": sp.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  sp.Name,
							Image: image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: mountPath,
									SubPath:   "frpc.ini",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: sp.Name,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceProxyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fcorev1.ServiceProxy{}).
		Complete(r)
}
