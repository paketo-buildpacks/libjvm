/*
 * Copyright 2018-2020 the original author or authors.
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
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type JavaSecurityProperties struct {
	LayerContributor libpak.LayerContributor
	Logger           bard.Logger
}

func NewJavaSecurityProperties(info libcnb.BuildpackInfo) JavaSecurityProperties {
	return JavaSecurityProperties{
		LayerContributor: libpak.NewLayerContributor("Java Security Properties", info),
		Logger:           bard.NewLogger(os.Stdout),
	}
}

func (j JavaSecurityProperties) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	return j.LayerContributor.Contribute(layer, func() (libcnb.Layer, error) {
		file := filepath.Join(layer.Path, "java-security.properties")
		if err := ioutil.WriteFile(file, []byte{}, 0644); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to touch file %s: %w", file, err)
		}

		layer.LaunchEnvironment.Append("JAVA_OPTS", ` -Djava.security.properties=%s`, file)
		layer.LaunchEnvironment.Override("JAVA_SECURITY_PROPERTIES", file)

		layer.Launch = true
		return layer, nil
	})
}

func (JavaSecurityProperties) Name() string {
	return "java-security-properties"
}
