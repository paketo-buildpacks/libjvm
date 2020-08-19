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

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
	"github.com/paketo-buildpacks/libpak/sherpa"

	_ "github.com/paketo-buildpacks/libjvm/statik"
)

type JRE struct {
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
	Metadata         map[string]interface{}
}

func NewJRE(dependency libpak.BuildpackDependency, cache libpak.DependencyCache, metadata map[string]interface{},
	plan *libcnb.BuildpackPlan) JRE {

	return JRE{
		LayerContributor: libpak.NewDependencyLayerContributor(dependency, cache, plan),
		Metadata:         metadata,
	}
}

//go:generate statik -src . -include *.sh

func (j JRE) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	j.LayerContributor.Logger = j.Logger

	return j.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		j.Logger.Bodyf("Expanding to %s", layer.Path)
		if err := crush.ExtractTarGz(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand JRE\n%w", err)
		}

		if IsBuildContribution(j.Metadata) {
			layer.BuildEnvironment.Default("JAVA_HOME", layer.Path)
			layer.Build = true
			layer.Cache = true
		}

		if IsLaunchContribution(j.Metadata) {
			layer.LaunchEnvironment.Default("JAVA_HOME", layer.Path)
			layer.LaunchEnvironment.Default("MALLOC_ARENA_MAX", "2")

			s, err := sherpa.StaticFile("/active-processor-count.sh")
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to load active-processor-count.sh\n%w", err)
			}
			layer.Profile.Add("active-processor-count.sh", s)

			s, err = sherpa.StaticFile("/java-tool-options.sh")
			if err != nil {
				return libcnb.Layer{}, fmt.Errorf("unable to load java-tool-options.sh\n%w", err)
			}
			layer.Profile.Add("java-tool-options.sh", s)

			layer.Launch = true
		}

		return layer, nil
	})
}

func (j JRE) Name() string {
	return j.LayerContributor.LayerName()
}
