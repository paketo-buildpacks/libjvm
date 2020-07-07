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
	"github.com/paketo-buildpacks/libpak/crush"
	"github.com/paketo-buildpacks/libpak/sherpa"

	_ "github.com/paketo-buildpacks/libjvm/statik"
)

type MemoryCalculator struct {
	ApplicationPath  string
	JavaVersion      string
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewMemoryCalculator(applicationPath string, dependency libpak.BuildpackDependency, cache libpak.DependencyCache,
	javaVersion string, plan *libcnb.BuildpackPlan) MemoryCalculator {

	return MemoryCalculator{
		ApplicationPath:  applicationPath,
		LayerContributor: libpak.NewDependencyLayerContributor(dependency, cache, plan),
		JavaVersion:      javaVersion,
	}
}

//go:generate statik -src . -include *.sh

func (m MemoryCalculator) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	m.LayerContributor.Logger = m.Logger

	return m.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		m.Logger.Bodyf("Expanding to %s", layer.Path)
		if err := crush.ExtractTarGz(artifact, filepath.Join(layer.Path, "bin"), 0); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand Memory Calculator\n%w", err)
		}

		jvmClassCount, err := m.JvmClassCount()
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to calculate JVM class count for %s\n%w", m.JavaVersion, err)
		}

		s, err := sherpa.TemplateFile("/memory-calculator.sh", map[string]interface{}{
			"source":        m.ApplicationPath,
			"jvmClassCount": jvmClassCount,
		})
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load memory-calculator.sh\n%w", err)
		}

		layer.Profile.Add("memory-calculator.sh", s)

		layer.Launch = true
		return layer, nil
	})
}

func (MemoryCalculator) Name() string {
	return "memory-calculator"
}

func (m MemoryCalculator) JvmClassCount() (int, error) {
	v, err := semver.NewVersion(m.JavaVersion)
	if err != nil {
		return 0, fmt.Errorf("unable to parse Java version %s\n%w", m.JavaVersion, err)
	}

	if c, _ := semver.NewConstraint("^8"); c.Check(v) {
		return 27867, nil
	} else if c, _ := semver.NewConstraint("^9"); c.Check(v) {
		return 25565, nil
	} else if c, _ := semver.NewConstraint("^10"); c.Check(v) {
		return 28191, nil
	} else if c, _ := semver.NewConstraint("^11"); c.Check(v) {
		return 24219, nil
	} else if c, _ := semver.NewConstraint("^12"); c.Check(v) {
		return 24219, nil
	} else if c, _ := semver.NewConstraint("^13"); c.Check(v) {
		return 24219, nil
	} else if c, _ := semver.NewConstraint("^14"); c.Check(v) {
		return 24219, nil
	} else {
		return 24219, nil
	}
}
