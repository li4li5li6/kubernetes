/*
Copyright 2017 The Kubernetes Authors.

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

package federatedtypes

import (
	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	federationclientset "k8s.io/kubernetes/federation/client/clientset_generated/federation_clientset"
	fedutil "k8s.io/kubernetes/federation/pkg/federation-controller/util"
	kubeclientset "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

const (
	ReplicaSetKind                     = "replicaset"
	ReplicaSetControllerName           = "replicasets"
	FedReplicaSetPreferencesAnnotation = "federation.kubernetes.io/replica-set-preferences"
)

func init() {
	RegisterFederatedType(ReplicaSetKind, ReplicaSetControllerName, []schema.GroupVersionResource{extensionsv1.SchemeGroupVersion.WithResource(ReplicaSetControllerName)}, NewReplicaSetAdapter)
}

type ReplicaSetAdapter struct {
	*schedulingAdapter
	client federationclientset.Interface
}

func NewReplicaSetAdapter(client federationclientset.Interface) FederatedTypeAdapter {
	schedulingAdapter := schedulingAdapter{
		preferencesAnnotationName: FedReplicaSetPreferencesAnnotation,
		updateStatusFunc: func(obj pkgruntime.Object, status SchedulingStatus) error {
			rs := obj.(*extensionsv1.ReplicaSet)
			if status.Replicas != rs.Status.Replicas || status.FullyLabeledReplicas != rs.Status.FullyLabeledReplicas ||
				status.ReadyReplicas != rs.Status.ReadyReplicas || status.AvailableReplicas != rs.Status.AvailableReplicas {
				rs.Status = extensionsv1.ReplicaSetStatus{
					Replicas:             status.Replicas,
					FullyLabeledReplicas: status.Replicas,
					ReadyReplicas:        status.ReadyReplicas,
					AvailableReplicas:    status.AvailableReplicas,
				}
				_, err := client.Extensions().ReplicaSets(rs.Namespace).UpdateStatus(rs)
				return err
			}
			return nil
		},
	}
	return &ReplicaSetAdapter{&schedulingAdapter, client}
}

func (a *ReplicaSetAdapter) Kind() string {
	return ReplicaSetKind
}

func (a *ReplicaSetAdapter) ObjectType() pkgruntime.Object {
	return &extensionsv1.ReplicaSet{}
}

func (a *ReplicaSetAdapter) IsExpectedType(obj interface{}) bool {
	_, ok := obj.(*extensionsv1.ReplicaSet)
	return ok
}

func (a *ReplicaSetAdapter) Copy(obj pkgruntime.Object) pkgruntime.Object {
	rs := obj.(*extensionsv1.ReplicaSet)
	return &extensionsv1.ReplicaSet{
		ObjectMeta: fedutil.DeepCopyRelevantObjectMeta(rs.ObjectMeta),
		Spec:       *fedutil.DeepCopyApiTypeOrPanic(&rs.Spec).(*extensionsv1.ReplicaSetSpec),
	}
}

func (a *ReplicaSetAdapter) Equivalent(obj1, obj2 pkgruntime.Object) bool {
	return fedutil.ObjectMetaAndSpecEquivalent(obj1, obj2)
}

func (a *ReplicaSetAdapter) NamespacedName(obj pkgruntime.Object) types.NamespacedName {
	replicaset := obj.(*extensionsv1.ReplicaSet)
	return types.NamespacedName{Namespace: replicaset.Namespace, Name: replicaset.Name}
}

func (a *ReplicaSetAdapter) ObjectMeta(obj pkgruntime.Object) *metav1.ObjectMeta {
	return &obj.(*extensionsv1.ReplicaSet).ObjectMeta
}

func (a *ReplicaSetAdapter) FedCreate(obj pkgruntime.Object) (pkgruntime.Object, error) {
	replicaset := obj.(*extensionsv1.ReplicaSet)
	return a.client.Extensions().ReplicaSets(replicaset.Namespace).Create(replicaset)
}

func (a *ReplicaSetAdapter) FedDelete(namespacedName types.NamespacedName, options *metav1.DeleteOptions) error {
	return a.client.Extensions().ReplicaSets(namespacedName.Namespace).Delete(namespacedName.Name, options)
}

func (a *ReplicaSetAdapter) FedGet(namespacedName types.NamespacedName) (pkgruntime.Object, error) {
	return a.client.Extensions().ReplicaSets(namespacedName.Namespace).Get(namespacedName.Name, metav1.GetOptions{})
}

func (a *ReplicaSetAdapter) FedList(namespace string, options metav1.ListOptions) (pkgruntime.Object, error) {
	return a.client.Extensions().ReplicaSets(namespace).List(options)
}

func (a *ReplicaSetAdapter) FedUpdate(obj pkgruntime.Object) (pkgruntime.Object, error) {
	replicaset := obj.(*extensionsv1.ReplicaSet)
	return a.client.Extensions().ReplicaSets(replicaset.Namespace).Update(replicaset)
}

func (a *ReplicaSetAdapter) FedWatch(namespace string, options metav1.ListOptions) (watch.Interface, error) {
	return a.client.Extensions().ReplicaSets(namespace).Watch(options)
}

func (a *ReplicaSetAdapter) ClusterCreate(client kubeclientset.Interface, obj pkgruntime.Object) (pkgruntime.Object, error) {
	replicaset := obj.(*extensionsv1.ReplicaSet)
	return client.Extensions().ReplicaSets(replicaset.Namespace).Create(replicaset)
}

func (a *ReplicaSetAdapter) ClusterDelete(client kubeclientset.Interface, nsName types.NamespacedName, options *metav1.DeleteOptions) error {
	return client.Extensions().ReplicaSets(nsName.Namespace).Delete(nsName.Name, options)
}

func (a *ReplicaSetAdapter) ClusterGet(client kubeclientset.Interface, namespacedName types.NamespacedName) (pkgruntime.Object, error) {
	return client.Extensions().ReplicaSets(namespacedName.Namespace).Get(namespacedName.Name, metav1.GetOptions{})
}

func (a *ReplicaSetAdapter) ClusterList(client kubeclientset.Interface, namespace string, options metav1.ListOptions) (pkgruntime.Object, error) {
	return client.Extensions().ReplicaSets(namespace).List(options)
}

func (a *ReplicaSetAdapter) ClusterUpdate(client kubeclientset.Interface, obj pkgruntime.Object) (pkgruntime.Object, error) {
	replicaset := obj.(*extensionsv1.ReplicaSet)
	return client.Extensions().ReplicaSets(replicaset.Namespace).Update(replicaset)
}

func (a *ReplicaSetAdapter) ClusterWatch(client kubeclientset.Interface, namespace string, options metav1.ListOptions) (watch.Interface, error) {
	return client.Extensions().ReplicaSets(namespace).Watch(options)
}

func (a *ReplicaSetAdapter) EquivalentIgnoringSchedule(obj1, obj2 pkgruntime.Object) bool {
	replicaset1 := obj1.(*extensionsv1.ReplicaSet)
	replicaset2 := a.Copy(obj2).(*extensionsv1.ReplicaSet)
	replicaset2.Spec.Replicas = replicaset1.Spec.Replicas
	return fedutil.ObjectMetaAndSpecEquivalent(replicaset1, replicaset2)
}

func (a *ReplicaSetAdapter) NewTestObject(namespace string) pkgruntime.Object {
	replicas := int32(3)
	zero := int64(0)
	return &extensionsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-replicaset-",
			Namespace:    namespace,
		},
		Spec: extensionsv1.ReplicaSetSpec{
			Replicas: &replicas,
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"foo": "bar"},
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: &zero,
					Containers: []apiv1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			},
		},
	}
}
