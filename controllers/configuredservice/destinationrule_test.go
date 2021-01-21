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

package configuredservice

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	meshv1alpha1 "github.com/symcn/meshach/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("DestinationRule", func() {
	var (
		cs *meshv1alpha1.ConfiguredService
	)

	BeforeEach(func() {
		cs = &meshv1alpha1.ConfiguredService{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test.normal.cs",
				Namespace: "mesh-test",
			},
			Spec: meshv1alpha1.ConfiguredServiceSpec{
				OriginalName:         "Test.Normal.Cs",
				Instances:            nil,
				MeshConfigGeneration: 0,
			},
		}
	})

	Context("test reconcile destinationrule", func() {
		It("create default destinationrule", func() {
			r := Reconciler{
				Client:     getFakeClient(),
				Log:        nil,
				Scheme:     getFakeScheme(),
				Opt:        testOpt,
				MeshConfig: getTestMeshConfig(),
			}
			err := r.reconcileDestinationRule(context.Background(), cs)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
