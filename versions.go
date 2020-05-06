/*
 * Copyright 2018-2020, VMware, Inc. All Rights Reserved.
 * Proprietary and Confidential.
 * Unauthorized use, copying or distribution of this source code via any medium is
 * strictly prohibited without the express written consent of VMware, Inc.
 */

package libjvm

import (
	"github.com/Masterminds/semver/v3"
)

var Java9, _ = semver.NewVersion("9")

func IsBeforeJava9(candidate string) bool {
	v, err := semver.NewVersion(candidate)
	if err != nil {
		return false
	}

	return v.LessThan(Java9)
}
