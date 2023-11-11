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
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fluid-cloudnative/fluid/pkg/utils/compatibility"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetCronJobStatus(t *testing.T) {
	nowTime := time.Now()
	testDate := metav1.NewTime(time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), nowTime.Hour(), 0, 0, 0, nowTime.Location()))

	namespace := "default"
	testCronJobInputs := []*batchv1.CronJob{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: namespace,
			},
			Status: batchv1.CronJobStatus{
				LastScheduleTime: &testDate,
			},
		},
	}

	testCronJobs := []runtime.Object{}

	for _, cj := range testCronJobInputs {
		testCronJobs = append(testCronJobs, cj.DeepCopy())
	}

	client := fake.NewFakeClientWithScheme(testScheme, testCronJobs...)

	type args struct {
		key types.NamespacedName
	}
	tests := []struct {
		name    string
		args    args
		want    *batchv1.CronJobStatus
		wantErr bool
	}{
		{
			name: "CronJob exists",
			args: args{
				key: types.NamespacedName{
					Namespace: namespace,
					Name:      "test1",
				},
			},
			want: &batchv1.CronJobStatus{
				LastScheduleTime: &testDate,
			},
			wantErr: false,
		},
		{
			name: "CronJob exists",
			args: args{
				key: types.NamespacedName{
					Namespace: namespace,
					Name:      "test-notexist",
				},
			},
			want:    nil,
			wantErr: true,
		},
	}

	patch := gomonkey.ApplyFunc(compatibility.IsBatchV1CronJobSupported, func() bool {
		return true
	})
	defer patch.Reset()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetCronJobStatus(client, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCronJobStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCronJobStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
