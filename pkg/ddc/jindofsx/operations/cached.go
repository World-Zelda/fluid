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

package operations

import (
	"fmt"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
)

// clean cache with a preset timeout of 60s
func (a JindoFileUtils) CleanCache() (err error) {
	var (
		// jindo jfs -formatCache -force
		command = []string{"jindo", "fs", "-formatCache", "-force"}
		stdout  string
		stderr  string
	)

	stdout, stderr, err = a.exec(command, false)

	if err != nil {
		err = fmt.Errorf("execute command %v with expectedErr: %v stdout %s and stderr %s", command, err, stdout, stderr)
		if utils.IgnoreNotFound(err) == nil {
			fmt.Printf("Failed to clean cache due to %v", err)
			return nil
		}
		return
	} else {
		time.Sleep(30 * time.Second)
	}

	return
}
