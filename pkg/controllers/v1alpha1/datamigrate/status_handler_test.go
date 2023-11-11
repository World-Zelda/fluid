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
package datamigrate

import (
	"reflect"
	"testing"
	"time"

	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestOnceGetOperationStatus(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	mockDataMigrate := v1alpha1.DataMigrate{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataMigrateSpec{},
	}

	mockJob := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-migrate-migrate",
			Namespace: "default",
		},
		Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{
				{
					Type:               batchv1.JobComplete,
					LastProbeTime:      v1.NewTime(time.Now()),
					LastTransitionTime: v1.NewTime(time.Now()),
				},
			},
		},
	}

	mockFailedJob := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-migrate-migrate",
			Namespace: "default",
		},
		Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{
				{
					Type:               batchv1.JobFailed,
					LastProbeTime:      v1.NewTime(time.Now()),
					LastTransitionTime: v1.NewTime(time.Now()),
				},
			},
		},
	}

	testcases := []struct {
		name          string
		job           batchv1.Job
		expectedPhase common.Phase
	}{
		{
			name:          "job success",
			job:           mockJob,
			expectedPhase: common.PhaseComplete,
		},
		{
			name:          "job failed",
			job:           mockFailedJob,
			expectedPhase: common.PhaseFailed,
		},
	}

	for _, testcase := range testcases {
		client := fake.NewFakeClientWithScheme(testScheme, &mockDataMigrate, &testcase.job)
		onceStatusHandler := &OnceStatusHandler{Client: client, dataMigrate: &mockDataMigrate}
		ctx := cruntime.ReconcileRequestContext{
			NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "",
			},
			Log: fake.NullLogger(),
		}
		opStatus, err := onceStatusHandler.GetOperationStatus(ctx, &mockDataMigrate.Status)
		if err != nil {
			t.Errorf("fail to GetOperationStatus with error %v", err)
		}
		if opStatus.Phase != testcase.expectedPhase {
			t.Error("Failed to GetOperationStatus", "expected phase", testcase.expectedPhase, "get", opStatus.Phase)
		}
	}
}

func TestCronGetOperationStatus(t *testing.T) {
	testScheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(testScheme)
	_ = batchv1.AddToScheme(testScheme)

	startTime := time.Date(2023, 8, 1, 12, 0, 0, 0, time.Local)
	lastScheduleTime := v1.NewTime(startTime)
	lastSuccessfulTime := v1.NewTime(startTime.Add(time.Second * 10))

	mockCronDataMigrate := v1alpha1.DataMigrate{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1alpha1.DataMigrateSpec{
			Policy:   "Cron",
			Schedule: "* * * * *",
		},
		Status: v1alpha1.OperationStatus{
			Phase: common.PhaseComplete,
		},
	}

	mockCronJob := batchv1.CronJob{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-migrate-migrate",
			Namespace: "default",
		},
		Spec: batchv1.CronJobSpec{
			Schedule: "* * * * *",
		},
		Status: batchv1.CronJobStatus{
			LastScheduleTime:   &lastScheduleTime,
			LastSuccessfulTime: &lastSuccessfulTime,
		},
	}

	mockJob := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-migrate-migrate-1",
			Namespace: "default",
			Labels: map[string]string{
				"cronjob": "test-migrate-migrate",
			},
			CreationTimestamp: lastScheduleTime,
		},
		Status: batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{
				{
					Type:               batchv1.JobComplete,
					LastProbeTime:      lastSuccessfulTime,
					LastTransitionTime: lastSuccessfulTime,
				},
			},
		},
	}

	mockRunningJob := batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-migrate-migrate-1",
			Namespace: "default",
			Labels: map[string]string{
				"cronjob": "test-migrate-migrate",
			},
			CreationTimestamp: lastScheduleTime,
		},
	}

	testcases := []struct {
		name          string
		job           batchv1.Job
		expectedPhase common.Phase
	}{
		{
			name:          "job complete",
			job:           mockJob,
			expectedPhase: common.PhaseComplete,
		},
		{
			name:          "job running",
			job:           mockRunningJob,
			expectedPhase: common.PhasePending,
		},
	}

	for _, testcase := range testcases {
		client := fake.NewFakeClientWithScheme(testScheme, &mockCronDataMigrate, &mockCronJob, &testcase.job)
		cronStatusHandler := &CronStatusHandler{Client: client, dataMigrate: &mockCronDataMigrate}
		ctx := cruntime.ReconcileRequestContext{Log: fake.NullLogger()}
		opStatus, err := cronStatusHandler.GetOperationStatus(ctx, &mockCronDataMigrate.Status)
		if err != nil {
			t.Errorf("fail to GetOperationStatus with error %v", err)
		}
		if !reflect.DeepEqual(opStatus.LastScheduleTime, &lastScheduleTime) || !reflect.DeepEqual(opStatus.LastSuccessfulTime, &lastSuccessfulTime) {
			t.Error("fail to get correct Operation Status", "expected LastScheduleTime", lastScheduleTime, "expected LastSuccessfulTime", lastSuccessfulTime, "get", opStatus)
		}
		if opStatus.Phase != testcase.expectedPhase {
			t.Error("Failed to GetOperationStatus", "expected phase", testcase.expectedPhase, "get", opStatus.Phase)
		}
	}
}
