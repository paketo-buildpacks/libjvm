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

	"github.com/buildpacks/libcnb"
	"github.com/heroku/color"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	md, err := libpak.NewBuildpackMetadata(context.Buildpack.Metadata)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to unmarshal buildpack metadata\n%w", err)
	}

	dr, err := libpak.NewDependencyResolver(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency resolver\n%w", err)
	}

	dc := libpak.NewDependencyCache(context.Buildpack)
	dc.Logger = b.Logger

	b.Logger.Title(context.Buildpack)
	result := libcnb.BuildResult{}

	if e, ok, err := pr.Resolve("jdk"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve jdk plan entry\n%w", err)
	} else if ok {
		dep, err := dr.Resolve("jdk", sherpa.ResolveVersion("BP_JAVA_VERSION", e, "jdk", md.DefaultVersions))
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		jdk := NewJDK(dep, dc, &result.Plan)
		jdk.Logger = b.Logger
		result.Layers = append(result.Layers, jdk)
	}

	if e, ok, err := pr.Resolve("jre"); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve jre plan entry\n%w", err)
	} else if ok {
		depJRE, err := dr.Resolve("jre", sherpa.ResolveVersion("BP_JAVA_VERSION", e, "jre", md.DefaultVersions))

		if libpak.IsNoValidDependencies(err) {
			warn := color.New(color.FgYellow, color.Bold)
			b.Logger.Header(warn.Sprint("No valid JRE available, providing matching JDK instead. Using a JDK at runtime has security implications."))
			depJRE, err = dr.Resolve("jdk", sherpa.ResolveVersion("BP_JAVA_VERSION", e, "jdk", md.DefaultVersions))
		}

		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		jre := NewJRE(depJRE, dc, e.Metadata, &result.Plan)
		jre.Logger = b.Logger
		result.Layers = append(result.Layers, jre)

		depJVMKill, err := dr.Resolve("jvmkill", "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		jk := NewJVMKill(depJVMKill, dc, &result.Plan)
		jk.Logger = b.Logger
		result.Layers = append(result.Layers, jk)

		lld := NewLinkLocalDNS(context.Buildpack, &result.Plan)
		lld.Logger = b.Logger
		result.Layers = append(result.Layers, lld)

		depMemCalc, err := dr.Resolve("memory-calculator", "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		mc := NewMemoryCalculator(context.Application.Path, depMemCalc, dc, depJRE.Version, &result.Plan)
		mc.Logger = b.Logger
		result.Layers = append(result.Layers, mc)

		cc := NewClassCounter(context.Buildpack, &result.Plan)
		cc.Logger = b.Logger
		result.Layers = append(result.Layers, cc)

		jsp := NewJavaSecurityProperties(context.Buildpack.Info)
		jsp.Logger = b.Logger
		result.Layers = append(result.Layers, jsp)

		spc := NewSecurityProvidersConfigurer(context.Buildpack, depJRE.Version, &result.Plan)
		spc.Logger = b.Logger
		result.Layers = append(result.Layers, spc)
	}

	return result, nil
}
