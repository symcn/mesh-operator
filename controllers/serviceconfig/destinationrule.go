/*
Copyright 2020 The Symcn Authors.

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

package serviceconfig

import (
	"context"

	meshv1alpha1 "github.com/symcn/meshach/api/v1alpha1"
	"github.com/symcn/meshach/pkg/utils"
	v1beta1 "istio.io/api/networking/v1beta1"
	networkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	loadBalanceSimple = "simple"
)

var lbMap = map[string]v1beta1.LoadBalancerSettings_SimpleLB{
	"ROUND_ROBIN": v1beta1.LoadBalancerSettings_ROUND_ROBIN,
	"LEAST_CONN":  v1beta1.LoadBalancerSettings_LEAST_CONN,
	"RANDOM":      v1beta1.LoadBalancerSettings_RANDOM,
	"PASSTHROUGH": v1beta1.LoadBalancerSettings_PASSTHROUGH,
}

func (r *Reconciler) reconcileDestinationRule(ctx context.Context, sc *meshv1alpha1.ServiceConfig) error {
	foundMap, err := r.getDestinationRuleMap(ctx, sc)
	if err != nil {
		klog.Errorf("%s/%s get DestinationRules error: %+v", sc.Namespace, sc.Name, err)
		return err
	}

	subsets, err := r.getSubset(ctx, sc)
	if err != nil {
		klog.Errorf("Get subsets[%s/%s] error: %+v", sc.Namespace, sc.Name, err)
		return err
	}

	// Skip if the service's subset is none
	klog.V(6).Infof("destinationrule subsets length: %d", len(subsets))
	if len(subsets) != 0 {
		dr := r.buildDestinationRule(sc, subsets)
		// Set ServiceConfig instance as the owner and controller
		if err := controllerutil.SetControllerReference(sc, dr, r.Scheme); err != nil {
			klog.Errorf("SetControllerReference error: %v", err)
			return err
		}

		// Check if this DestinationRule already exists
		found, ok := foundMap[dr.Name]
		if !ok {
			klog.Infof("Creating a new DestinationRule, Namespace: %s, Name: %s", dr.Namespace, dr.Name)
			err = r.Create(ctx, dr)
			if err != nil {
				if apierrors.IsAlreadyExists(err) {
					return nil
				}
				klog.Errorf("Create DestinationRule error: %+v", err)
				return err
			}
		} else {
			// Update DestinationRule
			if compareDestinationRule(dr, found) {
				klog.Infof("Update DestinationRule, Namespace: %s, Name: %s",
					found.Namespace, found.Name)
				err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
					// TODO get found
					dr.Spec.DeepCopyInto(&found.Spec)
					found.Finalizers = dr.Finalizers
					found.Labels = dr.ObjectMeta.Labels

					updateErr := r.Update(ctx, found)
					if updateErr == nil {
						klog.Infof("%s/%s update DestinationRule successfully",
							dr.Namespace, dr.Name)
						return nil
					}
					return updateErr
				})

				if err != nil {
					klog.Warningf("Update DestinationRule [%s] spec failed, err: %+v", dr.Name, err)
					return err
				}
			}
			delete(foundMap, dr.Name)
		}

		// Delete old DestinationRules
		for name, dr := range foundMap {
			klog.Infof("Delete unused DestinationRule: %s", name)
			err := r.Delete(ctx, dr)
			if err != nil {
				klog.Errorf("Delete unused DestinationRule error: %+v", err)
				return err
			}
		}
	}

	return nil
}

func (r *Reconciler) buildDestinationRule(svc *meshv1alpha1.ServiceConfig, actualSubsets []*meshv1alpha1.Subset) *networkingv1beta1.DestinationRule {
	var subsets []*v1beta1.Subset
	for _, sub := range actualSubsets {
		subset := &v1beta1.Subset{Name: sub.Name, Labels: sub.Labels}
		if sub.Policy != nil {
			subset.TrafficPolicy = &v1beta1.TrafficPolicy{
				LoadBalancer: &v1beta1.LoadBalancerSettings{
					LbPolicy: &v1beta1.LoadBalancerSettings_Simple{
						Simple: getlb(sub.Policy.LoadBalancer[loadBalanceSimple]),
					},
				},
			}
		}
		subsets = append(subsets, subset)
	}
	return &networkingv1beta1.DestinationRule{
		ObjectMeta: v1.ObjectMeta{
			Name:      utils.FormatToDNS1123(svc.Name),
			Namespace: svc.Namespace,
			Labels:    map[string]string{r.Opt.SelectLabel: truncated(svc.Spec.OriginalName)},
		},
		Spec: v1beta1.DestinationRule{
			Host: svc.Name,
			TrafficPolicy: &v1beta1.TrafficPolicy{
				LoadBalancer: &v1beta1.LoadBalancerSettings{
					LbPolicy: &v1beta1.LoadBalancerSettings_Simple{
						Simple: getlb(svc.Spec.Policy.LoadBalancer[loadBalanceSimple]),
					},
				},
			},
			Subsets: subsets,
		},
	}
}

func (r *Reconciler) getSubset(ctx context.Context, sc *meshv1alpha1.ServiceConfig) ([]*meshv1alpha1.Subset, error) {
	var subsets []*meshv1alpha1.Subset
	cs := &meshv1alpha1.ConfiguredService{}
	err := r.Get(ctx, types.NamespacedName{Namespace: sc.Namespace, Name: sc.Name}, cs)
	if err != nil {
		if errors.IsNotFound(err) {
			klog.V(6).Infof("The configuredservice[%s/%s] not found, skip", sc.Namespace, sc.Name)
			return subsets, nil
		}
		return subsets, err
	}

	for _, global := range r.MeshConfig.Spec.GlobalSubsets {
		for _, ins := range cs.Spec.Instances {
			if mapContains(ins.Labels, global.Labels, r.MeshConfig.Spec.MeshLabelsRemap) && !global.In(subsets) {
				subsets = append(subsets, global)
			}
		}
	}
	return subsets, nil
}

func compareDestinationRule(new, old *networkingv1beta1.DestinationRule) bool {
	if !equality.Semantic.DeepEqual(new.ObjectMeta.Finalizers, old.ObjectMeta.Finalizers) {
		return true
	}

	if !equality.Semantic.DeepEqual(new.ObjectMeta.Labels, old.ObjectMeta.Labels) {
		return true
	}

	if !equality.Semantic.DeepEqual(new.Spec, old.Spec) {
		return true
	}
	return false
}

func getlb(s string) v1beta1.LoadBalancerSettings_SimpleLB {
	lb, ok := lbMap[s]
	if !ok {
		lb = v1beta1.LoadBalancerSettings_RANDOM
	}
	return lb
}

func (r *Reconciler) getDestinationRuleMap(ctx context.Context, sc *meshv1alpha1.ServiceConfig) (map[string]*networkingv1beta1.DestinationRule, error) {
	list := &networkingv1beta1.DestinationRuleList{}
	labels := &client.MatchingLabels{r.Opt.SelectLabel: truncated(sc.Spec.OriginalName)}
	opts := &client.ListOptions{Namespace: sc.Namespace}
	labels.ApplyToList(opts)

	err := r.List(ctx, list, opts)
	if err != nil {
		return nil, err
	}
	m := make(map[string]*networkingv1beta1.DestinationRule)
	for i := range list.Items {
		item := list.Items[i]
		m[item.Name] = &item
	}
	return m, nil
}
