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

package prefernodeswithoutcache

import (
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetPreferredSchedulingTermForPodWithoutCacheWithGlobalMode(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	// Test case 1: Global fuse with selector enable
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{"test1": "test1"})
	term := getPreferredSchedulingTermForPodWithoutCache()

	expectTerm := corev1.PreferredSchedulingTerm{
		Weight: 100,
		Preference: corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      common.GetDatasetNumLabelName(),
					Operator: corev1.NodeSelectorOpDoesNotExist,
				},
			},
		},
	}

	if !reflect.DeepEqual(term, expectTerm) {
		t.Errorf("getPreferredSchedulingTermForPodWithoutCache failure, want:%v, got:%v", expectTerm, term)
	}

	// Test case 2: Global fuse with selector disable
	runtimeInfo.SetupFuseDeployMode(true, map[string]string{})
	term = getPreferredSchedulingTermForPodWithoutCache()

	if !reflect.DeepEqual(term, expectTerm) {
		t.Errorf("getPreferredSchedulingTermForPodWithoutCache failure, want:%v, got:%v", expectTerm, term)
	}
}

func TestGetPreferredSchedulingTermForPodWithoutCacheWithDefaultMode(t *testing.T) {
	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	runtimeInfo.SetupFuseDeployMode(false, map[string]string{})
	term := getPreferredSchedulingTermForPodWithoutCache()

	expectTerm := corev1.PreferredSchedulingTerm{
		Weight: 100,
		Preference: corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      common.GetDatasetNumLabelName(),
					Operator: corev1.NodeSelectorOpDoesNotExist,
				},
			},
		},
	}

	if !reflect.DeepEqual(term, expectTerm) {
		t.Errorf("getPreferredSchedulingTermForPodWithoutCache failure, want:%v, got:%v", expectTerm, term)
	}
}

func TestMutate(t *testing.T) {
	var (
		client client.Client
		pod    *corev1.Pod
	)

	plugin := NewPlugin(client)
	if plugin.GetName() != Name {
		t.Errorf("GetName expect %v, got %v", Name, plugin.GetName())
	}

	runtimeInfo, err := base.BuildRuntimeInfo("test", "fluid", "alluxio", datav1alpha1.TieredStore{})
	if err != nil {
		t.Errorf("fail to create the runtimeInfo with error %v", err)
	}

	pod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}

	shouldStop, err := plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"test": runtimeInfo})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	if !shouldStop {
		t.Errorf("expect shouldStop as true, but got %v", shouldStop)
	}

	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

	_, err = plugin.Mutate(pod, map[string]base.RuntimeInfoInterface{"test": nil})
	if err != nil {
		t.Errorf("fail to mutate pod with error %v", err)
	}

}
