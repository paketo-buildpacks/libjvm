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

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type ClassCounter struct {
	LayerContributor libpak.HelperLayerContributor
	Logger           bard.Logger
}

func NewClassCounter(buildpack libcnb.Buildpack, plan *libcnb.BuildpackPlan) ClassCounter {
	return ClassCounter{
		LayerContributor: libpak.NewHelperLayerContributor(filepath.Join(buildpack.Path, "bin", "class-counter"),
			"Class Counter", buildpack.Info, plan),
		Logger: bard.NewLogger(os.Stdout),
	}
}

func (c ClassCounter) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	return c.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		c.Logger.Body("Copying to %s", layer.Path)
		if err := sherpa.CopyFile(artifact, filepath.Join(layer.Path, "bin", "class-counter")); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy: %w", err)
		}

		layer.Launch = true
		return layer, nil
	})
}

func (ClassCounter) Name() string {
	return "class-counter"
}
