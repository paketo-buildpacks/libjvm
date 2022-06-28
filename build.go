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
	"github.com/mattn/go-shellwords"
	"github.com/paketo-buildpacks/libpak/effect"
	"strings"

	"github.com/buildpacks/libcnb"
	"github.com/heroku/color"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger          bard.Logger
	Result          libcnb.BuildResult
	CertLoader      CertificateLoader
	DependencyCache libpak.DependencyCache
}

func NewBuild(logger bard.Logger) Build {
	cl := NewCertificateLoader()
	cl.Logger = logger.BodyWriter()

	return Build{
		Logger:     logger,
		Result:     libcnb.NewBuildResult(),
		CertLoader: cl,
	}
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	var jdkRequired, jreRequired, jreMissing, jreSkipped, jLinkEnabled bool

	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	_, jdkRequired, err := pr.Resolve("jdk")
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve jdk plan entry\n%w", err)
	}

	jrePlanEntry, jreRequired, err := pr.Resolve("jre")
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve jre plan entry\n%w", err)
	}

	if !jdkRequired && !jreRequired {
		return b.Result, nil
	}
	b.Logger.Title(context.Buildpack)

	cr, err := libpak.NewConfigurationResolver(context.Buildpack, &b.Logger)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	jvmVersion := NewJVMVersion(b.Logger)
	v, err := jvmVersion.GetJVMVersion(context.Application.Path, cr)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to determine jvm version\n%w", err)
	}

	dr, err := libpak.NewDependencyResolver(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency resolver\n%w", err)
	}

	b.DependencyCache, err = libpak.NewDependencyCache(context)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
	}
	b.DependencyCache.Logger = b.Logger

	depJDK, err := dr.Resolve("jdk", v)
	if jdkRequired && err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
	}

	jreMissing = false
	depJRE, err := dr.Resolve("jre", v)
	if libpak.IsNoValidDependencies(err) {
		jreMissing = true
	}

	if t, _ := cr.Resolve("BP_JVM_TYPE"); strings.ToLower(t) == "jdk" {
		jreSkipped = true
	}

	if jl := cr.ResolveBool("BP_JVM_JLINK_ENABLED"); jl {
		jLinkEnabled = true
	}

	// jLink
	if jLinkEnabled {
		if IsBeforeJava9(v) {
			return libcnb.BuildResult{}, fmt.Errorf("unable to build, jlink is compatible with Java 9+ only\n")
		}
		if err = b.contributeJDK(depJDK); err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to contribute JDK for Jlink\n%w", err)
		}
		if err = b.contributeJLink(cr, jrePlanEntry.Metadata, context.Application.Path, depJDK); err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to contribute Jlink\n%w", err)
		}
		b.contributeHelpers(context, depJDK)
		return b.Result, nil
	}

	// use JDK as JRE
	if jreRequired && (jreSkipped || jreMissing) {
		b.warnIfJreNotUsed(jreMissing, jreSkipped)
		if err = b.contributeJDKAsJRE(depJDK, jrePlanEntry, context); err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to contribute JDK as JRE\n%w", err)
		}
		b.contributeHelpers(context, depJDK)
		return b.Result, nil
	}

	// contribute a JDK
	if jdkRequired {
		if err = b.contributeJDK(depJDK); err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to contribute JDK \n%w", err)
		}
	}

	// contribute a JRE
	if jreRequired {
		dt := JREType
		if err = b.contributeJRE(depJRE, context.Application.Path, dt, jrePlanEntry.Metadata); err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to contribute JDK \n%w", err)
		}
		if IsLaunchContribution(jrePlanEntry.Metadata) {
			b.contributeHelpers(context, depJRE)
		}
	}

	return b.Result, nil
}

func (b *Build) contributeJDK(jdkDep libpak.BuildpackDependency) error {
	jdk, be, err := NewJDK(jdkDep, b.DependencyCache, b.CertLoader)
	if err != nil {
		return fmt.Errorf("unable to create jdk\n%w", err)
	}

	jdk.Logger = b.Logger
	b.Result.Layers = append(b.Result.Layers, jdk)
	b.Result.BOM.Entries = append(b.Result.BOM.Entries, be)
	return nil
}

func (b *Build) contributeJDKAsJRE(jdkDep libpak.BuildpackDependency, jrePlanEntry libcnb.BuildpackPlanEntry, context libcnb.BuildContext) error {
	// This forces the contributed layer to be build + cache + launch so it's available everywhere
	jrePlanEntry.Metadata["build"] = true
	jrePlanEntry.Metadata["cache"] = true

	dt := JDKType
	if err := b.contributeJRE(jdkDep, context.Application.Path, dt, jrePlanEntry.Metadata); err != nil {
		return fmt.Errorf("unable to contribute JRE\n%w", err)
	}
	return nil
}

func (b *Build) contributeJRE(jreDep libpak.BuildpackDependency, appPath string, distributionType DistributionType, metadata map[string]interface{}) error {
	jre, be, err := NewJRE(appPath, jreDep, b.DependencyCache, distributionType, b.CertLoader, metadata)
	if err != nil {
		return fmt.Errorf("unable to create jdk\n%w", err)
	}

	jre.Logger = b.Logger
	b.Result.Layers = append(b.Result.Layers, jre)
	b.Result.BOM.Entries = append(b.Result.BOM.Entries, be)
	return nil
}

func (b *Build) contributeJLink(configurationResolver libpak.ConfigurationResolver, planEntryMetadata map[string]interface{}, appPath string, jdkDep libpak.BuildpackDependency) error {
	args, explicit := configurationResolver.Resolve("BP_JVM_JLINK_ARGS")
	argList, err := shellwords.Parse(args)
	if err != nil {
		return fmt.Errorf("unable to parse jlink arguments %s %w\n", args, err)
	}

	jlink, err := NewJLink(appPath, effect.NewExecutor(), argList, b.CertLoader, planEntryMetadata, explicit)
	if err != nil {
		return fmt.Errorf("unable to create jlink jre\n%w", err)
	}
	jlink.JavaVersion = jdkDep.Version
	jlink.Logger = b.Logger
	b.Result.Layers = append(b.Result.Layers, jlink)
	return nil
}

func (b *Build) contributeHelpers(context libcnb.BuildContext, depJRE libpak.BuildpackDependency) {
	helpers := []string{"active-processor-count", "java-opts", "jvm-heap", "link-local-dns", "memory-calculator",
		"security-providers-configurer", "jmx", "jfr"}

	if IsBeforeJava9(depJRE.Version) {
		helpers = append(helpers, "security-providers-classpath-8")
		helpers = append(helpers, "debug-8")
	} else {
		helpers = append(helpers, "security-providers-classpath-9")
		helpers = append(helpers, "debug-9")
		helpers = append(helpers, "nmt")
	}
	// Java 18 bug - cacerts keystore type not readable
	if IsBeforeJava18(depJRE.Version) {
		helpers = append(helpers, "openssl-certificate-loader")
	}

	h, be := libpak.NewHelperLayer(context.Buildpack, helpers...)
	h.Logger = b.Logger
	b.Result.Layers = append(b.Result.Layers, h)
	b.Result.BOM.Entries = append(b.Result.BOM.Entries, be)

	jsp := NewJavaSecurityProperties(context.Buildpack.Info)
	jsp.Logger = b.Logger
	b.Result.Layers = append(b.Result.Layers, jsp)
}

func (b Build) warnIfJreNotUsed(jreMissing, jreSkipped bool) {
	msg := "Using a JDK at runtime has security implications."

	if jreMissing && !jreSkipped {
		msg = fmt.Sprintf("No valid JRE available, providing matching JDK instead. %s", msg)
	}

	if jreSkipped {
		subMsg := "A JDK was specifically requested by the user"
		if !jreMissing {
			subMsg = fmt.Sprintf("%s, however a JRE is available", subMsg)
		} else {
			subMsg = fmt.Sprintf("%s and a JDK is the only option", subMsg)
		}
		msg = fmt.Sprintf("%s. %s", subMsg, msg)
	}

	b.Logger.Header(color.New(color.FgYellow, color.Bold).Sprint(msg))
}
