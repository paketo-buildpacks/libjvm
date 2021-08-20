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
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/heroku/color"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	result := libcnb.NewBuildResult()

	cr, err := libpak.NewConfigurationResolver(context.Buildpack, &b.Logger)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	dr, err := libpak.NewDependencyResolver(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency resolver\n%w", err)
	}

	dc, err := libpak.NewDependencyCache(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
	}
	dc.Logger = b.Logger

	cl := NewCertificateLoader()
	cl.Logger = b.Logger.BodyWriter()

	v, _ := cr.Resolve("BP_JVM_VERSION")

	jreSkipped := false
	if t, _ := cr.Resolve("BP_JVM_TYPE"); strings.ToLower(t) == "jdk" {
		jreSkipped = true
	}

	_, jdkRequired, err := pr.Resolve("jdk")
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve jdk plan entry\n%w", err)
	}

	jrePlanEntry, jreRequired, err := pr.Resolve("jre")
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve jre plan entry\n%w", err)
	}

	jreAvailable := jreRequired
	if jreRequired {
		_, err := dr.Resolve("jre", v)
		if libpak.IsNoValidDependencies(err) {
			jreAvailable = false
		}
	}

	// we need a JDK, we're not using the JDK as a JRE and the JRE has not been skipped
	if jdkRequired && !(jreRequired && !jreAvailable) && !jreSkipped {
		dep, err := dr.Resolve("jdk", v)
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		jdk, be, err := NewJDK(dep, dc, cl)
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to create jdk\n%w", err)
		}

		jdk.Logger = b.Logger
		result.Layers = append(result.Layers, jdk)
		result.BOM.Entries = append(result.BOM.Entries, be)
	}

	if jreRequired {
		dt := JREType
		depJRE, err := dr.Resolve("jre", v)

		if !jreAvailable || jreSkipped {
			b.warnIfJreNotUsed(jreAvailable, jreSkipped)

			// This forces the contributed layer to be build + cache + launch so it's available everywhere
			jrePlanEntry.Metadata["build"] = true
			jrePlanEntry.Metadata["cache"] = true

			dt = JDKType
			depJRE, err = dr.Resolve("jdk", v)
		}

		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		jre, be, err := NewJRE(context.Application.Path, depJRE, dc, dt, cl, jrePlanEntry.Metadata)
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to create jre\n%w", err)
		}

		jre.Logger = b.Logger
		result.Layers = append(result.Layers, jre)
		result.BOM.Entries = append(result.BOM.Entries, be)

		if IsLaunchContribution(jrePlanEntry.Metadata) {
			helpers := []string{"active-processor-count", "java-opts", "link-local-dns", "memory-calculator",
				"openssl-certificate-loader", "security-providers-configurer"}

			if IsBeforeJava9(depJRE.Version) {
				helpers = append(helpers, "security-providers-classpath-8")
			} else {
				helpers = append(helpers, "security-providers-classpath-9")
			}

			h, be := libpak.NewHelperLayer(context.Buildpack, helpers...)
			h.Logger = b.Logger
			result.Layers = append(result.Layers, h)
			result.BOM.Entries = append(result.BOM.Entries, be)

			depJVMKill, err := dr.Resolve("jvmkill", "")
			if err != nil {
				return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
			}

			jk, be := NewJVMKill(depJVMKill, dc)
			jk.Logger = b.Logger
			result.Layers = append(result.Layers, jk)
			result.BOM.Entries = append(result.BOM.Entries, be)

			jsp := NewJavaSecurityProperties(context.Buildpack.Info)
			jsp.Logger = b.Logger
			result.Layers = append(result.Layers, jsp)
		}
	}

	return result, nil
}

func (b Build) warnIfJreNotUsed(jreAvailable, jreSkipped bool) {
	msg := "Using a JDK at runtime has security implications."

	if !jreAvailable && !jreSkipped {
		msg = fmt.Sprintf("No valid JRE available, providing matching JDK instead. %s", msg)
	}

	if jreSkipped {
		subMsg := "A JDK was specifically requested by the user"
		if jreAvailable {
			subMsg = fmt.Sprintf("%s, however a JRE is available", subMsg)
		} else {
			subMsg = fmt.Sprintf("%s and a JDK is the only option", subMsg)
		}
		msg = fmt.Sprintf("%s. %s", subMsg, msg)
	}

	b.Logger.Header(color.New(color.FgYellow, color.Bold).Sprint(msg))
}
