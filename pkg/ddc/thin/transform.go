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

package thin

import (
	"fmt"
	"time"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/transfromer"
	corev1 "k8s.io/api/core/v1"
)

func (t *ThinEngine) transform(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinRuntimeProfile) (value *ThinValue, err error) {
	if runtime == nil {
		err = fmt.Errorf("the thinRuntime is null")
		return
	}
	defer utils.TimeTrack(time.Now(), "ThinRuntime.Transform", "name", runtime.Name)

	dataset, err := utils.GetDataset(t.Client, t.name, t.namespace)
	if err != nil {
		return value, err
	}

	value = &ThinValue{
		RuntimeIdentity: common.RuntimeIdentity{
			Namespace: runtime.Namespace,
			Name:      runtime.Name,
		},
	}

	value.FullnameOverride = t.name
	value.Owner = transfromer.GenerateOwnerReferenceFromObject(runtime)
	toRuntimeSetConfig, err := t.toRuntimeSetConfig(nil, nil)
	if err != nil {
		return
	}
	value.RuntimeValue = toRuntimeSetConfig

	// transform toleration
	t.transformTolerations(dataset, value)

	// transform the workers
	err = t.transformWorkers(runtime, profile, value)
	if err != nil {
		return
	}

	// transform the fuse
	err = t.transformFuse(runtime, profile, dataset, value)
	if err != nil {
		return
	}

	// set the placementMode
	t.transformPlacementMode(dataset, value)
	return
}

func (t *ThinEngine) transformWorkers(runtime *datav1alpha1.ThinRuntime, profile *datav1alpha1.ThinRuntimeProfile, value *ThinValue) (err error) {
	value.Worker = Worker{
		Envs:  []corev1.EnvVar{},
		Ports: []corev1.ContainerPort{},
	}

	// parse config from profile
	t.parseFromProfile(profile, value)

	// 1. image
	t.parseWorkerImage(runtime, value)

	// 2. env
	if len(runtime.Spec.Worker.Env) != 0 {
		value.Worker.Envs = append(value.Worker.Envs, runtime.Spec.Worker.Env...)
	}
	// 3. ports
	if len(runtime.Spec.Worker.Ports) != 0 {
		value.Worker.Ports = append(value.Worker.Ports, runtime.Spec.Worker.Ports...)
	}
	// 4. nodeSelector
	if len(runtime.Spec.Worker.NodeSelector) != 0 {
		value.Worker.NodeSelector = runtime.Spec.Worker.NodeSelector
	}

	// 5. cachedir
	if len(runtime.Spec.TieredStore.Levels) > 0 {
		value.Worker.CacheDir = runtime.Spec.TieredStore.Levels[0].Path
	}
	// 6. volume
	err = t.transformWorkerVolumes(runtime.Spec.Volumes, runtime.Spec.Worker.VolumeMounts, value)
	if err != nil {
		t.Log.Error(err, "failed to transform volumes for worker")
	}

	// 7. resources
	t.transformResourcesForWorker(runtime.Spec.Worker.Resources, value)

	// 8. probe
	if runtime.Spec.Worker.ReadinessProbe != nil {
		value.Worker.ReadinessProbe = runtime.Spec.Worker.ReadinessProbe
	}
	if runtime.Spec.Worker.LivenessProbe != nil {
		value.Worker.LivenessProbe = runtime.Spec.Worker.LivenessProbe
	}
	// 9. network
	value.Worker.HostNetwork = datav1alpha1.IsHostNetwork(runtime.Spec.Worker.NetworkMode)
	return
}

func (t *ThinEngine) transformPlacementMode(dataset *datav1alpha1.Dataset, value *ThinValue) {
	value.PlacementMode = string(dataset.Spec.PlacementMode)
	if len(value.PlacementMode) == 0 {
		value.PlacementMode = string(datav1alpha1.ExclusiveMode)
	}
}

func (t *ThinEngine) parseWorkerImage(runtime *datav1alpha1.ThinRuntime, value *ThinValue) {
	if len(runtime.Spec.Worker.Image) != 0 {
		value.Worker.Image = runtime.Spec.Worker.Image
	}
	if len(runtime.Spec.Worker.ImageTag) != 0 {
		value.Worker.ImageTag = runtime.Spec.Worker.ImageTag
	}
	if len(runtime.Spec.Worker.ImagePullPolicy) != 0 {
		value.Worker.ImagePullPolicy = runtime.Spec.Worker.ImagePullPolicy
	}
}

func (t *ThinEngine) parseFromProfile(profile *datav1alpha1.ThinRuntimeProfile, value *ThinValue) {
	if profile == nil {
		return
	}
	// 1. image
	value.Worker.Image = profile.Spec.Worker.Image
	value.Worker.ImageTag = profile.Spec.Worker.ImageTag
	value.Worker.ImagePullPolicy = profile.Spec.Worker.ImagePullPolicy
	// 2. volumes
	err := t.transformWorkerVolumes(profile.Spec.Volumes, profile.Spec.Worker.VolumeMounts, value)
	if err != nil {
		t.Log.Error(err, "failed to transform volumes from profile for worker")
	}
	// 3. resources
	t.transformResourcesForWorker(profile.Spec.Worker.Resources, value)

	// 4. env
	if len(profile.Spec.Worker.Env) != 0 {
		value.Worker.Envs = profile.Spec.Worker.Env
	}
	// 5. nodeSelector
	if len(profile.Spec.Worker.NodeSelector) != 0 {
		value.Worker.NodeSelector = profile.Spec.Worker.NodeSelector
	}
	// 6. ports
	if len(profile.Spec.Worker.Ports) != 0 {
		value.Worker.Ports = profile.Spec.Worker.Ports
	}
	// 7. probe
	if profile.Spec.Worker.ReadinessProbe != nil {
		value.Worker.ReadinessProbe = profile.Spec.Worker.ReadinessProbe
	}
	if profile.Spec.Worker.LivenessProbe != nil {
		value.Worker.LivenessProbe = profile.Spec.Worker.LivenessProbe
	}
	// 8. network
	value.Worker.HostNetwork = datav1alpha1.IsHostNetwork(profile.Spec.Worker.NetworkMode)
}

func (t *ThinEngine) transformTolerations(dataset *datav1alpha1.Dataset, value *ThinValue) {
	if len(dataset.Spec.Tolerations) > 0 {
		// value.Tolerations = dataset.Spec.Tolerations
		value.Tolerations = []corev1.Toleration{}
		for _, toleration := range dataset.Spec.Tolerations {
			toleration.TolerationSeconds = nil
			value.Tolerations = append(value.Tolerations, toleration)
		}
	}
}
