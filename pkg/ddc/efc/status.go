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

package efc

import (
	"context"
	"reflect"
	"time"

	data "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

// CheckAndUpdateRuntimeStatus checks the related runtime status and updates it.
func (e *EFCEngine) CheckAndUpdateRuntimeStatus() (ready bool, err error) {
	var (
		masterReady, workerReady bool
		masterName               string = e.getMasterName()
		workerName               string = e.getWorkerName()
		namespace                string = e.namespace
	)

	// 1. Master should be ready
	master, err := kubeclient.GetStatefulSet(e.Client, masterName, namespace)
	if err != nil {
		return ready, err
	}

	// 2. Worker should be ready
	workers, err := ctrl.GetWorkersAsStatefulset(e.Client,
		types.NamespacedName{Namespace: e.namespace, Name: workerName})
	if err != nil {
		return ready, err
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		runtime, err := e.getRuntime()
		if err != nil {
			return err
		}

		runtimeToUpdate := runtime.DeepCopy()

		states, err := e.queryCacheStatus()
		if err != nil {
			return err
		}

		if len(runtime.Status.CacheStates) == 0 {
			runtimeToUpdate.Status.CacheStates = map[common.CacheStateName]string{}
		}

		runtimeToUpdate.Status.CacheStates[common.CacheCapacity] = states.cacheCapacity
		runtimeToUpdate.Status.CacheStates[common.CachedPercentage] = states.cachedPercentage
		runtimeToUpdate.Status.CacheStates[common.Cached] = states.cached
		// update cache hit ratio
		runtimeToUpdate.Status.CacheStates[common.CacheHitRatio] = states.cacheHitStates.cacheHitRatio
		runtimeToUpdate.Status.CacheStates[common.LocalHitRatio] = states.cacheHitStates.localHitRatio
		runtimeToUpdate.Status.CacheStates[common.RemoteHitRatio] = states.cacheHitStates.remoteHitRatio
		// update cache throughput ratio
		runtimeToUpdate.Status.CacheStates[common.LocalThroughputRatio] = states.cacheHitStates.localThroughputRatio
		runtimeToUpdate.Status.CacheStates[common.RemoteThroughputRatio] = states.cacheHitStates.remoteThroughputRatio
		runtimeToUpdate.Status.CacheStates[common.CacheThroughputRatio] = states.cacheHitStates.cacheThroughputRatio

		runtimeToUpdate.Status.CurrentMasterNumberScheduled = int32(master.Status.Replicas)
		runtimeToUpdate.Status.MasterNumberReady = int32(master.Status.ReadyReplicas)
		if *master.Spec.Replicas == master.Status.ReadyReplicas {
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseReady
			masterReady = true
		} else {
			runtimeToUpdate.Status.MasterPhase = data.RuntimePhaseNotReady
		}

		runtimeToUpdate.Status.CurrentWorkerNumberScheduled = int32(workers.Status.Replicas)
		runtimeToUpdate.Status.WorkerNumberReady = int32(workers.Status.ReadyReplicas)
		runtimeToUpdate.Status.WorkerNumberUnavailable = int32(*workers.Spec.Replicas - workers.Status.ReadyReplicas)
		runtimeToUpdate.Status.WorkerNumberAvailable = int32(workers.Status.CurrentReplicas)
		if runtime.Replicas() == workers.Status.ReadyReplicas {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseReady
			workerReady = true
		} else if workers.Status.ReadyReplicas >= 1 {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhasePartialReady
			workerReady = true
		} else {
			runtimeToUpdate.Status.WorkerPhase = data.RuntimePhaseNotReady
		}

		if masterReady && workerReady {
			ready = true
		}

		// Update the setup time of EFC runtime
		if ready && runtimeToUpdate.Status.SetupDuration == "" {
			runtimeToUpdate.Status.SetupDuration = utils.CalculateDuration(runtimeToUpdate.CreationTimestamp.Time, time.Now())
		}

		if !reflect.DeepEqual(runtime.Status, runtimeToUpdate.Status) {
			err = e.Client.Status().Update(context.TODO(), runtimeToUpdate)
		} else {
			e.Log.Info("Do nothing because the runtime status is not changed.")
		}

		return err
	})

	if err != nil {
		_ = utils.LoggingErrorExceptConflict(e.Log, err, "Failed to update runtime status", types.NamespacedName{Namespace: e.namespace, Name: e.name})
	}

	return
}
