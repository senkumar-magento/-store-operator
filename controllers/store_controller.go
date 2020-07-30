/*


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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storesv1 "store.com/api/v1"
)

// StoreReconciler reconciles a Store object
type StoreReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=stores.store.com,resources=stores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=stores.store.com,resources=stores/status,verbs=get;update;patch

func (r *StoreReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	context := context.Background()
	log := r.Log.WithValues("store", req.NamespacedName)

	// your logic here
	// Fetch the App instance.
	app := &storesv1.Store{}
	err := r.Get(context, req.NamespacedName, app)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Store does n't exist or may be deleted ...")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("Provisioning the Store ...")

	return ctrl.Result{}, nil
}

func (r *StoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&storesv1.Store{}).
		Complete(r)
}

func (r *StoreReconciler) deploymentForApp(m *storesv1.Store) *appsv1.Deployment {
	ls := labelsForApp(m.Name)
	var size int32 = 2

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webapp",
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "localhost:5000/demo-webapp",
						Name:  "webapp",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 9000,
							Name:          "webapp",
						}},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *StoreReconciler) deploymentForRedis(m *storesv1.Store) *appsv1.Deployment {
	ls := labelsForApp(m.Name)
	var size int32 = 2

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "localhost:5000/redis",
						Name:  m.Name,
						Ports: []corev1.ContainerPort{{
							ContainerPort: 9000,
							Name:          m.Name,
						}},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// labelsForApp creates a simple set of labels for App.
func labelsForApp(name string) map[string]string {
	return map[string]string{"app_name": "app", "app_cr": name}
}
