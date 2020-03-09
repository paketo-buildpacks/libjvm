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
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type SecurityProvidersConfigurer struct {
	JavaVersion      string
	LayerContributor libpak.HelperLayerContributor
	Logger           bard.Logger
}

func NewSecurityProvidersConfigurer(buildpack libcnb.Buildpack, javaVersion string, plan *libcnb.BuildpackPlan) SecurityProvidersConfigurer {
	return SecurityProvidersConfigurer{
		JavaVersion: javaVersion,
		LayerContributor: libpak.NewHelperLayerContributor(filepath.Join(buildpack.Path, "bin", "security-providers-configurer"),
		"Security Providers Configurer", buildpack.Info, plan),
		Logger:      bard.NewLogger(os.Stdout),
	}
}

func (s SecurityProvidersConfigurer) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	return s.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		s.Logger.Body("Copying to %s", layer.Path)
		if err := sherpa.CopyFile(artifact, filepath.Join(layer.Path, "bin", "security-providers-configurer")); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy: %w", err)
		}

		j9, _ := semver.NewVersion("9")
		v, err := semver.NewVersion(s.JavaVersion)
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to parse Java version %s: %w", s.JavaVersion, err)
		}

		var source string
		if v.LessThan(j9) {
			source = filepath.Join("lib", "security", "java.security")

			layer.Profile.Add("security-providers-classpath", `EXT_DIRS="$JAVA_HOME/lib/ext"

for I in ${SECURITY_PROVIDERS_CLASSPATH//:/$'\n'}; do
  EXT_DIRS="$EXT_DIRS:$(dirname $I)"
done

JAVA_OPTS="$JAVA_OPTS -Djava.ext.dirs=$EXT_DIRS"`)
		} else {
			source = filepath.Join("conf", "security", "java.security")

			layer.Profile.Add("security-providers-classpath", "export CLASSPATH=$CLASSPATH:$SECURITY_PROVIDERS_CLASSPATH")
		}

		layer.Profile.Add("security-providers-configurer",
			`SECURITY_PROVIDERS=$(echo $SECURITY_PROVIDERS | tr ' ' ,)

security-provider-configurer --source "$JAVA_HOME"/%s --additional-providers "$SECURITY_PROVIDERS"
`, source)

		layer.Launch = true
		return layer, nil
	})
}

func (SecurityProvidersConfigurer) Name() string {
	return "security-providers-configurer"
}
