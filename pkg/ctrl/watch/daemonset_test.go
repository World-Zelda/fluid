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

package watch

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestDaemonSetEventHandler_OnCreateFunc(t *testing.T) {

	// 1. the Object is RuntimeInterface
	createEvent := event.CreateEvent{
		Object: &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       datav1alpha1.JindoRuntimeKind,
						APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
						Controller: utilpointer.BoolPtr(true),
					},
				},
			},
		},
	}
	daemonsetEventHandler := &daemonsetEventHandler{}

	f := daemonsetEventHandler.onCreateFunc(&FakeRuntimeReconciler{})
	predicate := f(createEvent)

	if !predicate {
		t.Errorf("The event %v should be reconciled, but skip.", createEvent)
	}

	// 2. the Object is not RuntimeInterface
	createEvent.Object = &corev1.Pod{}
	predicate = f(createEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", createEvent)
	}

	// 3. Skip the runtime which is deleting
	createEvent = event.CreateEvent{
		Object: &appsv1.DaemonSet{},
	}
	createEvent.Object.SetDeletionTimestamp(&metav1.Time{})
	predicate = f(createEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", createEvent)
	}

}

func TestDaemonSetEventHandler_OnUpdateFunc(t *testing.T) {

	updateRuntimeEvent := event.UpdateEvent{
		ObjectOld: &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       datav1alpha1.JindoRuntimeKind,
						APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
						Controller: utilpointer.BoolPtr(true),
					},
				},
				ResourceVersion: "123",
			},
		},
		ObjectNew: &appsv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Kind:       datav1alpha1.JindoRuntimeKind,
						APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
						Controller: utilpointer.BoolPtr(true),
					},
				},
				ResourceVersion: "456",
			},
		},
	}
	daemonsetEventHandler := &daemonsetEventHandler{}

	f := daemonsetEventHandler.onUpdateFunc(&FakeRuntimeReconciler{})
	predicate := f(updateRuntimeEvent)

	// 1. expect the updateEvent is validated
	if !predicate {
		t.Errorf("The event %v should be reconciled, but skip.", updateRuntimeEvent)
	}

	// 2. expect the updateEvent is not validated due to the resource version is equal
	updateRuntimeEvent.ObjectOld.SetResourceVersion("456")
	predicate = f(updateRuntimeEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", updateRuntimeEvent)
	}

	// 3. expect the updateEvent is not validated due to the object is not kind of runtimeInterface
	updateRuntimeEvent.ObjectOld = &corev1.Pod{}
	updateRuntimeEvent.ObjectNew = &corev1.Pod{}
	predicate = f(updateRuntimeEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", updateRuntimeEvent)
	}

	// 4. expect the updateEvent is not validate due the old Object  is not kind of the runtimeInterface
	updateRuntimeEvent.ObjectNew = &appsv1.DaemonSet{}
	predicate = f(updateRuntimeEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", updateRuntimeEvent)
	}
}

func TestDaemonSetEventHandler_OnDeleteFunc(t *testing.T) {

	// 1. the Object is RuntimeInterface
	delRuntimeEvent := event.DeleteEvent{
		Object: &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       datav1alpha1.JindoRuntimeKind,
					APIVersion: datav1alpha1.GroupVersion.Group + "/" + datav1alpha1.GroupVersion.Version,
					Controller: utilpointer.BoolPtr(true),
				},
			},
		}},
	}
	daemonsetEventHandler := &daemonsetEventHandler{}

	f := daemonsetEventHandler.onDeleteFunc(&FakeRuntimeReconciler{})
	predicate := f(delRuntimeEvent)

	if !predicate {
		t.Errorf("The event %v should be reconciled, but skip.", delRuntimeEvent)
	}

	// 2. the Object is not RuntimeInterface
	delRuntimeEvent.Object = &corev1.Pod{}
	predicate = f(delRuntimeEvent)
	if predicate {
		t.Errorf("The event %v should ben't reconciled, but pass.", delRuntimeEvent)
	}
}
