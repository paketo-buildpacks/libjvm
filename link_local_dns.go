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

	_ "github.com/paketo-buildpacks/libjvm/statik"
)

type LinkLocalDNS struct {
	LayerContributor libpak.HelperLayerContributor
	Logger           bard.Logger
}

func NewLinkLocalDNS(buildpack libcnb.Buildpack, plan *libcnb.BuildpackPlan) LinkLocalDNS {
	return LinkLocalDNS{
		LayerContributor: libpak.NewHelperLayerContributor(filepath.Join(buildpack.Path, "bin", "link-local-dns"),
			"Link-Local DNS", buildpack.Info, plan),
	}
}

//go:generate statik -src . -include *.sh

func (l LinkLocalDNS) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	l.LayerContributor.Logger = l.Logger

	return l.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		l.Logger.Bodyf("Copying to %s", layer.Path)
		if err := sherpa.CopyFile(artifact, filepath.Join(layer.Path, "bin", "link-local-dns")); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy\n%w", err)
		}

		s, err := sherpa.StaticFile("/link-local-dns.sh")
		if err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to load link-local-dns.sh\n%w", err)
		}

		layer.Profile.Add("link-local-dns.sh", s)

		layer.Launch = true
		return layer, nil
	})
}

func (LinkLocalDNS) Name() string {
	return "link-local-dns"
}
