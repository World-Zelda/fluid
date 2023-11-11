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

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNode gets the latest node info
func GetNode(client client.Reader, name string) (node *corev1.Node, err error) {
	key := types.NamespacedName{
		Name: name,
	}

	node = &corev1.Node{}

	if err = client.Get(context.TODO(), key, node); err != nil {
		return nil, err
	}
	return node, err
}

// IsReady checks if the node is ready
// If the node is ready,it returns True.Otherwise,it returns False.
func IsReady(node corev1.Node) (ready bool) {
	ready = true
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status != corev1.ConditionTrue {
			ready = false
			break
		}
	}
	return ready
}

// GetIpAddressesOfNodes gets the ipAddresses of nodes
func GetIpAddressesOfNodes(nodes []corev1.Node) (ipAddresses []string) {
	// realIPs = make([]net.IP, 0, len(nodes))
	for _, node := range nodes {
		for _, address := range node.Status.Addresses {
			if address.Type == corev1.NodeInternalIP {
				if address.Address != "" {
					ipAddresses = append(ipAddresses, address.Address)
				} else {
					log.Info("Failed to get ipAddresses from the node", "node", node.Name)
				}
			}
		}
	}
	return utils.SortIpAddresses(ipAddresses)
}
