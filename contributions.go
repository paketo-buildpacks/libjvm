/*
 * Copyright 2018-2020, VMware, Inc. All Rights Reserved.
 * Proprietary and Confidential.
 * Unauthorized use, copying or distribution of this source code via any medium is
 * strictly prohibited without the express written consent of VMware, Inc.
 */

package libjvm

func IsBuildContribution(metadata map[string]interface{}) bool {
	v, ok := metadata["build"].(bool)
	return ok && v
}

func IsLaunchContribution(metadata map[string]interface{}) bool {
	v, ok := metadata["launch"].(bool)
	return ok && v
}
