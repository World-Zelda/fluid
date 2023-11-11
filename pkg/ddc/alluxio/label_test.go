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

package alluxio

import "testing"

func TestGetCommonLabelname(t *testing.T) {
	testCases := []struct {
		name      string
		namespace string
		out       string
	}{
		{
			name:      "hbase",
			namespace: "fluid",
			out:       "fluid.io/s-fluid-hbase",
		},
		{
			name:      "hadoop",
			namespace: "fluid",
			out:       "fluid.io/s-fluid-hadoop",
		},
		{
			name:      "common",
			namespace: "default",
			out:       "fluid.io/s-default-common",
		},
	}
	for _, testCase := range testCases {
		engine := &AlluxioEngine{
			name:      testCase.name,
			namespace: testCase.namespace,
		}
		out := engine.getCommonLabelname()
		if out != testCase.out {
			t.Errorf("in: %s-%s, expect: %s, got: %s", testCase.namespace, testCase.name, testCase.out, out)
		}
	}
}
