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

package goosefs

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/goosefs/operations"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
)

// Query the hcfs status
func (e *GooseFSEngine) GetHCFSStatus() (status *datav1alpha1.HCFSStatus, err error) {
	endpoint, err := e.queryHCFSEndpoint()
	if err != nil {
		e.Log.Error(err, "Failed to get HCFS Endpoint")
		return status, err
	}

	version, err := e.queryCompatibleUFSVersion()
	if err != nil {
		e.Log.Error(err, "Failed to get Compatiable Endpoint")
		return status, err
	}

	status = &datav1alpha1.HCFSStatus{
		Endpoint:                    endpoint,
		UnderlayerFileSystemVersion: version,
	}
	return
}

// query the hcfs endpoint
func (e *GooseFSEngine) queryHCFSEndpoint() (endpoint string, err error) {

	var (
		serviceName = fmt.Sprintf("%s-master-0", e.name)
		host        = fmt.Sprintf("%s.%s", serviceName, e.namespace)
	)

	svc, err := kubeclient.GetServiceByName(e.Client, serviceName, e.namespace)
	if err != nil {
		e.Log.Error(err, "Failed to get Endpoint")
		return endpoint, err
	}

	if svc == nil {
		e.Log.Error(fmt.Errorf("failed to find the svc %s in %s", e.name, e.namespace), "failed to find the svc, it's nil")
		return
	}

	for _, port := range svc.Spec.Ports {
		if port.Name == "rpc" {
			endpoint = fmt.Sprintf("goosefs://%s:%d", host, port.Port)
			return
		}
	}

	return
}

// query the compatible version of UFS
func (e *GooseFSEngine) queryCompatibleUFSVersion() (version string, err error) {
	podName, containerName := e.getMasterPodInfo()
	fileUtils := operations.NewGooseFSFileUtils(podName, containerName, e.namespace, e.Log)
	version, err = fileUtils.GetConf("goosefs.underfs.version")
	if err != nil {
		e.Log.Error(err, "Failed to getConf")
		return
	}
	return
}
