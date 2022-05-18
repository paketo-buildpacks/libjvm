/*
 * Copyright 2018-2022 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package libjvm

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// SDKInfo represents the information from each line in the `.sdkmanrc` file
type SDKInfo struct {
	Type    string
	Version string
	Vendor  string
}

// ReadSDKMANRC reads the `.sdkmanrc` format file from path and retuns the list of SDKS in it
func ReadSDKMANRC(path string) ([]SDKInfo, error) {
	sdkmanrcContents, err := ioutil.ReadFile(path)
	if err != nil {
		return []SDKInfo{}, fmt.Errorf("unable to read SDKMANRC file at %s\n%w", path, err)
	}

	sdks := []SDKInfo{}
	for _, line := range strings.Split(string(sdkmanrcContents), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "#", 2) // strip comments
		if len(parts) != 1 && len(parts) != 2 {
			return []SDKInfo{}, fmt.Errorf("unable to strip comments from %q resulted in %q", line, parts)
		}

		if strings.TrimSpace(parts[0]) != "" {
			kv := strings.SplitN(parts[0], "=", 2) // split key=value
			if len(kv) != 2 {
				return []SDKInfo{}, fmt.Errorf("unable to split key/value from %q resulted in %q", parts[0], kv)
			}

			versionAndVendor := []string{"", ""}
			if strings.TrimSpace(kv[1]) != "" {
				versionAndVendor = strings.SplitN(kv[1], "-", 2) // split optional vendor name
				if len(versionAndVendor) == 1 {
					versionAndVendor = append(versionAndVendor, "")
				}
				if len(versionAndVendor) != 2 {
					return []SDKInfo{}, fmt.Errorf("unable to split vendor from %q resulted in %q", kv[1], versionAndVendor)
				}
			}

			sdks = append(sdks, SDKInfo{
				Type:    strings.ToLower(strings.TrimSpace(kv[0])),
				Version: strings.TrimSpace(versionAndVendor[0]),
				Vendor:  strings.ToLower(strings.TrimSpace(versionAndVendor[1])),
			})
		}
	}

	return sdks, nil
}
