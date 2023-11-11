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

package kubeclient

import (
	"context"

	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetCronJobStatus gets CronJob's status given its namespace and name. It converts batchv1beta1.CronJobStatus
// to batchv1.CronJobStatus when batchv1.CronJob is not supported by the cluster.
func GetCronJobStatus(client client.Client, key types.NamespacedName) (*batchv1.CronJobStatus, error) {
	if compatibility.IsBatchV1CronJobSupported() {
		var cronjob batchv1.CronJob
		if err := client.Get(context.TODO(), key, &cronjob); err != nil {
			return nil, err
		}
		return &cronjob.Status, nil
	}

	var cronjob batchv1beta1.CronJob
	if err := client.Get(context.TODO(), key, &cronjob); err != nil {
		return nil, err
	}
	// Convert batchv1beta1.CronJobStatus to batchv1.CronJobStatus and return
	return &batchv1.CronJobStatus{
		Active:             cronjob.Status.Active,
		LastScheduleTime:   cronjob.Status.LastScheduleTime,
		LastSuccessfulTime: cronjob.Status.LastSuccessfulTime,
	}, nil
}
