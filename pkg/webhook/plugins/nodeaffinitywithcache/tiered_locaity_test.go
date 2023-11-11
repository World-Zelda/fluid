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

package nodeaffinitywithcache

import (
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func TestTieredLocality_hasRepeatedLocality(t1 *testing.T) {
	type args struct {
		pod *corev1.Pod
	}

	tieredLocality := &TieredLocality{
		Preferred: []Preferred{
			{
				Name:   "label.a",
				Weight: 1,
			},
			{
				Name:   "label.b",
				Weight: 2,
			},
		},
		Required: []string{"label.a"},
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty affinity and selector",
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{},
				},
			},
			want: false,
		},
		{
			name: "affinity and empty selector, has same label",
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Affinity: &corev1.Affinity{
							NodeAffinity: &corev1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
									NodeSelectorTerms: []corev1.NodeSelectorTerm{
										{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      "label.b",
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"b.value"},
												},
											},
										},
									},
								},
								PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
									{
										Weight: 10,
										Preference: corev1.NodeSelectorTerm{
											MatchExpressions: []corev1.NodeSelectorRequirement{
												{
													Key:      "label.b",
													Operator: corev1.NodeSelectorOpIn,
													Values:   []string{"b.value"},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "node selector with same label",
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						NodeSelector: map[string]string{
							"label.a": "a-value",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "node selector without same label",
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						NodeSelector: map[string]string{
							"label.c": "a-value",
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if got := tieredLocality.hasRepeatedLocality(tt.args.pod); got != tt.want {
				t1.Errorf("hasRepeatedLocality() = %v, want %v", got, tt.want)
			}
		})
	}
}
