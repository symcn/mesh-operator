/*
Copyright 2020 The symcn authors.

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

package app

import (
	"github.com/spf13/cobra"
	"github.com/symcn/meshach/pkg/adapter"
	"github.com/symcn/meshach/pkg/option"
	"github.com/symcn/meshach/pkg/utils"
	"k8s.io/klog"
)

// NewAdapterCmd ...
func NewAdapterCmd(ropt *option.RootOption) *cobra.Command {
	opt := option.DefaultAdapterOption()
	cmd := &cobra.Command{
		Use:     "adapter",
		Aliases: []string{"adapter"},
		Short:   "The adapters is used for synchronizing services & instances from these different registry center and configuration center",
		Run: func(cmd *cobra.Command, args []string) {
			PrintFlags(cmd.Flags())
			opt.EventHandlers.Kubeconfig = ropt.Kubeconfig
			opt.EventHandlers.ConfigContext = ropt.ConfigContext
			opt.EventHandlers.QPS = ropt.QPS
			opt.EventHandlers.Burst = ropt.Burst
			a, err := adapter.NewAdapter(opt)
			if err != nil {
				klog.Fatalf("unable to NewAdapter err: %v", err)
				return
			}

			stopCh := utils.SetupSignalHandler()
			if err = a.Start(stopCh); err != nil {
				klog.Fatalf("unable to start the adapter, err: %v", err)
				return
			}
		},
	}

	cmd.PersistentFlags().StringArrayVar(
		&opt.Registry.Address,
		"reg-addr",
		opt.Registry.Address,
		"address for registry center, e.g. zk: 127.0.0.1:2181")

	cmd.PersistentFlags().Int64Var(
		&opt.Registry.Timeout,
		"reg-timeout",
		opt.Registry.Timeout,
		"the zookeeper session timeout second for registry")

	cmd.PersistentFlags().StringArrayVar(
		&opt.Configuration.Address,
		"conf-addr",
		opt.Configuration.Address,
		"address for configuration center, e.g. zk: 127.0.0.1:2181")

	cmd.PersistentFlags().Int64Var(
		&opt.Configuration.Timeout,
		"conf-timeout",
		opt.Configuration.Timeout,
		"the zookeeper session timeout second for configuration center")

	cmd.PersistentFlags().BoolVar(
		&opt.EventHandlers.IsMultiClusters,
		"is-multi-clusters",
		opt.EventHandlers.IsMultiClusters,
		"deciding if it will synchronizes to more than one kubernetes cluster")

	cmd.PersistentFlags().IntVar(
		&opt.EventHandlers.AcceleratorSize,
		"acc-size",
		opt.EventHandlers.AcceleratorSize,
		"the size of accelerator's channels, e.g. 500, 0 as default, using sequence executions if the size is less than 0 ")

	cmd.PersistentFlags().StringVar(
		&opt.EventHandlers.ClusterOwner,
		"cluster-owner",
		opt.EventHandlers.ClusterOwner,
		"the labels that multiple cluster manager used for select clusters")

	cmd.PersistentFlags().StringVar(
		&opt.EventHandlers.ClusterNamespace,
		"cluster-namespace",
		opt.EventHandlers.ClusterNamespace,
		"the namespace that multiple cluster manager uses when selecting the cluster config maps")
	return cmd
}
