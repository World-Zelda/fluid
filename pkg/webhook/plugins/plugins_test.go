/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package plugins

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"math/rand"
	"testing"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/nodeaffinitywithcache"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins/prefernodeswithoutcache"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestPods(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	tieredLocalityConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.TieredLocalityConfigMapName,
			Namespace: common.NamespaceFluidSystem,
		},
		Data: map[string]string{
			"tieredLocality": "" +
				"preferred:\n" +
				"  # fluid existed node affinity, the name can not be modified.\n" +
				"  - name: fluid.io/node\n" +
				"    weight: 100\n",
		},
	}
	jindoRuntime := &datav1alpha1.JindoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hbase",
			Namespace: "default",
		},
	}

	schema := runtime.NewScheme()
	_ = corev1.AddToScheme(schema)
	_ = datav1alpha1.AddToScheme(schema)
	var (
		pod               corev1.Pod
		c                 = fake.NewFakeClientWithScheme(schema, tieredLocalityConfigMap, jindoRuntime)
		plugin            MutatingHandler
		pluginName        string
		lenNodePrefer     int
		lenNodeRequire    int
		lenPodPrefer      int
		lenPodAntiPrefer  int
		lenPodRequire     int
		lenPodAntiRequire int
	)

	// build slice of RuntimeInfos
	var nilRuntimeInfos map[string]base.RuntimeInfoInterface = map[string]base.RuntimeInfoInterface{}
	runtimeInfo, err := base.BuildRuntimeInfo("hbase", "default", "jindo", datav1alpha1.TieredStore{})
	if err != nil {
		t.Error("fail to build runtimeInfo because of err", err)
	}
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})
	runtimeInfo.SetDeprecatedNodeLabel(false)
	// runtimeInfos := append(nilRuntimeInfos, runtimeInfo)
	runtimeInfos := map[string]base.RuntimeInfoInterface{"hbase": runtimeInfo}

	// test all plugins for 3 turns
	for i := 0; i < 3; i++ {
		lenNodePrefer = rand.Intn(3) + 1
		lenNodeRequire = rand.Intn(3) + 1
		lenPodPrefer = rand.Intn(3) + 1
		lenPodAntiPrefer = rand.Intn(3) + 1
		lenPodRequire = rand.Intn(3) + 1
		lenPodAntiRequire = rand.Intn(3) + 1

		// build affinity of pod
		pod.Spec.Affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: make([]corev1.NodeSelectorTerm, lenNodeRequire),
				},
				PreferredDuringSchedulingIgnoredDuringExecution: make([]corev1.PreferredSchedulingTerm, lenNodePrefer),
			},
			PodAffinity: &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  make([]corev1.PodAffinityTerm, lenPodRequire),
				PreferredDuringSchedulingIgnoredDuringExecution: make([]corev1.WeightedPodAffinityTerm, lenPodPrefer),
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  make([]corev1.PodAffinityTerm, lenPodAntiRequire),
				PreferredDuringSchedulingIgnoredDuringExecution: make([]corev1.WeightedPodAffinityTerm, lenPodAntiPrefer),
			},
		}

		// test of plugin preferNodesWithoutCache
		plugin = prefernodeswithoutcache.NewPlugin(c)
		pluginName = plugin.GetName()
		_, err = plugin.Mutate(&pod, runtimeInfos)
		if err != nil {
			t.Error("failed to mutate because of err", err)
		}

		if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodPrefer ||
			len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodAntiRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodAntiPrefer {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		} else {
			if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != lenNodeRequire {
				t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
			}
		}
		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenNodePrefer {
			t.Errorf("the plugin %v should exit and call other plugins if the pod has mounted datasets", pluginName)
		}

		_, err = plugin.Mutate(&pod, nilRuntimeInfos)
		if err != nil {
			t.Error("failed to mutate because of err", err)
		}

		if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodPrefer ||
			len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodAntiRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodAntiPrefer {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		} else {
			if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != lenNodeRequire {
				t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
			}
		}
		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenNodePrefer+1 {
			t.Errorf("the plugin %v inject wrong terms when the pod has no mounted datasets", pluginName)
		}

		// test of plugin preferNodesWithCache
		plugin = nodeaffinitywithcache.NewPlugin(c)
		pluginName = plugin.GetName()
		_, err = plugin.Mutate(&pod, nilRuntimeInfos)
		if err != nil {
			t.Error("failed to mutate because of err", err)
		}

		if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodPrefer ||
			len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodAntiRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodAntiPrefer {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		} else {
			if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != lenNodeRequire {
				t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
			}
		}
		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenNodePrefer+1 {
			t.Errorf("the plugin %v should exit and call other plugins if the pod has no mounted datasets", pluginName)
		}

		_, err = plugin.Mutate(&pod, runtimeInfos)
		if err != nil {
			t.Error("failed to mutate because of err", err)
		}
		if len(pod.Spec.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodPrefer ||
			len(pod.Spec.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != lenPodAntiRequire ||
			len(pod.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenPodAntiPrefer {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		}
		if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
		} else {
			if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) != lenNodeRequire {
				t.Errorf("the plugin %v should only inject into NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution", pluginName)
			}
		}
		if len(pod.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != lenNodePrefer+2 {
			t.Errorf("the plugin %v inject wrong terms when the pod has mounted datasets", pluginName)
		}

	}

}

func TestRegistry(t *testing.T) {
	var (
		client client.Client
	)

	plugins := Registry(client)
	if len(plugins.GetPodWithDatasetHandler()) != 3 {
		t.Errorf("expect GetPodWithDatasetHandler len=3, got %v", plugins.GetPodWithDatasetHandler())
	}

	if len(plugins.GetPodWithoutDatasetHandler()) != 1 {
		t.Errorf("expect GetPodWithoutDatasetHandler len=1 got %v", plugins.GetPodWithoutDatasetHandler())
	}

	if len(plugins.GetServerlessPodHandler()) != 1 {
		t.Errorf("expect GetServerlessPodHandler len=1 got %v", plugins.GetServerlessPodHandler())
	}

}
