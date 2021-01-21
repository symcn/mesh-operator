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

package meshconfig

import (
	"context"

	"github.com/go-logr/logr"
	meshv1alpha1 "github.com/symcn/meshach/api/v1alpha1"
	"github.com/symcn/meshach/pkg/option"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var log = logf.Log.WithName("controller_meshconfig")

// Reconciler reconciles a MeshConfig object
type Reconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Opt    *option.ControllerOption
}

// +kubebuilder:rbac:groups=mesh.symcn.com,resources=*,verbs=get;list;watch;create;update;patch;delete

// Reconcile ...
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling MeshConfig")

	// Fetch the MeshConfig instance
	instance := &meshv1alpha1.MeshConfig{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		return ctrl.Result{}, err
	}

	saList := &meshv1alpha1.ServiceAccessorList{}
	err = r.List(context.TODO(), saList, &client.ListOptions{Namespace: corev1.NamespaceAll})
	if err != nil {
		klog.Infof("[meshconfig] Get ServiceConfig error: %s", err)
	}
	for i := range saList.Items {
		sa := saList.Items[i]
		if sa.Spec.MeshConfigGeneration != instance.Generation {
			sa.Spec.MeshConfigGeneration = instance.Generation
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				updateErr := r.Update(context.TODO(), &sa)
				if updateErr == nil {
					klog.Infof("[meshconfig] update ServiceAccessor[%s/%s] successfully", sa.Namespace, sa.Name)
					return nil
				}

				getErr := r.Get(context.TODO(), types.NamespacedName{
					Namespace: sa.Namespace,
					Name:      sa.Name,
				}, &sa)
				if getErr != nil {
					return getErr
				}

				return updateErr
			})
			if err != nil {
				klog.Errorf("[meshconfig] Update ServiceAccessor[%s/%s] in MeshConfig reconcile error: %+v",
					sa.Namespace, sa.Name, err)
			}
		}
	}

	scList := &meshv1alpha1.ServiceConfigList{}
	err = r.List(context.TODO(), scList, &client.ListOptions{Namespace: corev1.NamespaceAll})
	if err != nil {
		klog.Infof("[meshconfig] Get ServiceConfig error: %s", err)
	}
	for i := range scList.Items {
		sc := scList.Items[i]
		if sc.Spec.MeshConfigGeneration != instance.Generation {
			sc.Spec.MeshConfigGeneration = instance.Generation
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				updateErr := r.Update(context.TODO(), &sc)
				if updateErr == nil {
					klog.Infof("[meshconfig] update ServiceConfig[%s/%s] successfully", sc.Namespace, sc.Name)
					return nil
				}

				getErr := r.Get(context.TODO(), types.NamespacedName{
					Namespace: sc.Namespace,
					Name:      sc.Name,
				}, &sc)
				if getErr != nil {
					return getErr
				}

				return updateErr
			})
			if err != nil {
				klog.Errorf("[meshconfig] Update ServiceConfig[%s/%s] in MeshConfig reconcile error: %+v",
					sc.Namespace, sc.Name, err)
			}
		}
	}

	csList := &meshv1alpha1.ConfiguredServiceList{}
	err = r.List(context.TODO(), csList, &client.ListOptions{Namespace: corev1.NamespaceAll})
	if err != nil {
		klog.Infof("[meshconfig] Get ConfiguredService error: %s", err)
	}
	for i := range csList.Items {
		cs := csList.Items[i]
		if cs.Spec.MeshConfigGeneration != instance.Generation {
			cs.Spec.MeshConfigGeneration = instance.Generation
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				updateErr := r.Update(context.TODO(), &cs)
				if updateErr == nil {
					klog.Infof("[meshconfig] update ConfiguraredService[%s/%s] successfully", cs.Namespace, cs.Name)
					return nil
				}

				getErr := r.Get(context.TODO(), types.NamespacedName{
					Namespace: cs.Namespace,
					Name:      cs.Name,
				}, &cs)
				if getErr != nil {
					return getErr
				}

				return updateErr
			})
			if err != nil {
				klog.Errorf("[meshconfig] Update ConfiguredService[%s/%s] in MeshConfig reconcile error: %+v",
					cs.Namespace, cs.Name, err)
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager ...
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&meshv1alpha1.MeshConfig{}).
		WithEventFilter(predicate.Funcs{
			// Ignore updates to CR status in which case metadata.Generation does not change
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
			},
			// Ignore delete event
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
		}).
		Complete(r)
}
