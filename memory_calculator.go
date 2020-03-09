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
)

type MemoryCalculator struct {
	ApplicationPath  string
	Crush            crush.Crush
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
		Logger:           bard.NewLogger(os.Stdout),
	}
}

func (m MemoryCalculator) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	return m.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		m.Logger.Body("%s", bard.LaunchConfigFormatter{Name: "BPL_HEAD_ROOM", Default: "0"})
		m.Logger.Body("%s", bard.LaunchConfigFormatter{Name: "BPL_LOADED_CLASS_COUNT", Default: "35% of classes"})
		m.Logger.Body("%s", bard.LaunchConfigFormatter{Name: "BPL_THREAD_COUNT", Default: "250"})

		m.Logger.Body("Expanding to %s", layer.Path)
		if err := m.Crush.ExtractTarGz(artifact, filepath.Join(layer.Path, "bin"), 0); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand Memory Calculator: %w", err)
		}

		jvmClassCount, err := m.JvmClassCount()
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to calculate JVM class count for %s: %w", m.JavaVersion, err)
		}

		layer.Profile.Add("memory-calculator", `HEAD_ROOM=${BPL_HEAD_ROOM:=0}

if [[ -z "${BPL_LOADED_CLASS_COUNT+x}" ]]; then
    LOADED_CLASS_COUNT=$(class-counter --source "%s" --jvm-class-count "%d")
else
	LOADED_CLASS_COUNT=${BPL_LOADED_CLASS_COUNT}
fi

THREAD_COUNT=${BPL_THREAD_COUNT:=250}

TOTAL_MEMORY=$(cat /sys/fs/cgroup/memory/memory.limit_in_bytes)

if [ ${TOTAL_MEMORY} -eq 9223372036854771712 ]; then
  printf "Container memory limit unset. Configuring JVM for 1G container.\n"
  TOTAL_MEMORY=1073741824
elif [ ${TOTAL_MEMORY} -gt 70368744177664 ]; then
  printf "Container memory limit too large. Configuring JVM for 64T container.\n"
  TOTAL_MEMORY=70368744177664
fi

MEMORY_CONFIGURATION=$(java-buildpack-memory-calculator \
    --head-room "${HEAD_ROOM}" \
    --jvm-options "${JAVA_OPTS}" \
    --loaded-class-count "${LOADED_CLASS_COUNT}" \
    --thread-count "${THREAD_COUNT}" \
    --total-memory "${TOTAL_MEMORY}")

printf "Calculated JVM Memory Configuration: ${MEMORY_CONFIGURATION} (Head Room: ${HEAD_ROOM}%%%%, Loaded Class Count: ${LOADED_CLASS_COUNT}, Thread Count: ${THREAD_COUNT}, Total Memory: ${TOTAL_MEMORY})\n"
export JAVA_OPTS="${JAVA_OPTS} ${MEMORY_CONFIGURATION}"
`, m.ApplicationPath, jvmClassCount)

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
		return 0, fmt.Errorf("unable to parse Java version %s: %w", m.JavaVersion, err)
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
