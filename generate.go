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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libcnb/v2"
	"github.com/paketo-buildpacks/libpak/v2"
	"github.com/paketo-buildpacks/libpak/v2/bard"
)

type GenerateContentBuilder func(GenerateContentContext) (GenerateContentResult, error)

type Generate struct {
	Logger                 bard.Logger
	GenerateContentBuilder GenerateContentBuilder
}

type GenerateContentContext struct {
	ConfigurationResolver libpak.ConfigurationResolver
	DependencyResolver    libpak.DependencyResolver
	DependencyCache       libpak.DependencyCache
	Context               libcnb.GenerateContext
	PlanEntryResolver     libpak.PlanEntryResolver
	Logger                bard.Logger
	Result                libcnb.GenerateResult
}

type GenerateContentResult struct {
	ExtendConfig    ExtendConfig
	BuildDockerfile io.Reader
	RunDockerfile   io.Reader
	GenerateResult  libcnb.GenerateResult
}

type ExtendConfig struct {
	Build ExtendImageConfig `toml:"build"`
}

type ExtendImageConfig struct {
	Args []ExtendImageConfigArg `toml:"args"`
}

type ExtendImageConfigArg struct {
	Name  string `toml:"name"`
	Value string `toml:"value"`
}

func NewGenerate(logger bard.Logger, contentBuilder GenerateContentBuilder) Generate {
	return Generate{
		Logger:                 logger,
		GenerateContentBuilder: contentBuilder,
	}
}

func (b Generate) Generate(context libcnb.GenerateContext) (libcnb.GenerateResult, error) {
	var jdkRequired, jreRequired, nativeImage bool

	//as per libjvm buildpack.Build, look for jdk/jre/native-image-builder and exit if none present.
	pr := libpak.PlanEntryResolver{Plan: context.Plan}

	_, jdkRequired, err := pr.Resolve("jdk")
	if err != nil {
		return libcnb.GenerateResult{}, fmt.Errorf("unable to resolve jdk plan entry\n%w", err)
	}

	_, jreRequired, err = pr.Resolve("jre")
	if err != nil {
		return libcnb.GenerateResult{}, fmt.Errorf("unable to resolve jre plan entry\n%w", err)
	}

	_, nativeImage, err = pr.Resolve("native-image-builder")
	if err != nil {
		return libcnb.GenerateResult{}, fmt.Errorf("unable to resolve native-image-builder plan entry\n%w", err)
	}

	if !jdkRequired && !jreRequired && !nativeImage {
		return libcnb.NewGenerateResult(), nil
	}

	//still here? log out we're going to do something.
	b.Logger.Title(context.Extension.Info.Name, context.Extension.Info.Version, context.Extension.Info.Homepage)

	//build out the GenerateContentContext
	bpm, err := libpak.NewBuildModuleMetadata(context.Extension.Metadata)
	if err != nil {
		return libcnb.GenerateResult{}, fmt.Errorf("unable to create build module metadata\n%w", err)
	}

	cr, err := libpak.NewConfigurationResolver(bpm, &b.Logger)
	if err != nil {
		return libcnb.GenerateResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	dr, err := libpak.NewDependencyResolver(bpm, context.StackID)
	if err != nil {
		return libcnb.GenerateResult{}, fmt.Errorf("unable to create dependency resolver\n%w", err)
	}

	dc, err := libpak.NewDependencyCache(context.Extension.Info.ID, context.Extension.Info.Version, context.Extension.Path, context.Platform.Bindings)
	if err != nil {
		return libcnb.GenerateResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
	}

	//pass control to the callback to allow it to generate Dockerfiles as required.
	result, err := b.GenerateContentBuilder(GenerateContentContext{
		PlanEntryResolver:     pr,
		ConfigurationResolver: cr,
		DependencyResolver:    dr,
		DependencyCache:       dc,
		Context:               context,
		Logger:                b.Logger,
	})
	if err != nil {
		return libcnb.GenerateResult{}, fmt.Errorf("error invoking dockerfile callback\n%w", err)
	}

	//write any returned content to appropriate files, where required.
	if result.BuildDockerfile != nil {
		err = writeFile(result.BuildDockerfile, context.OutputDirectory, "build.Dockerfile")
		if err != nil {
			return libcnb.GenerateResult{}, err
		}
	}
	if result.RunDockerfile != nil {
		err = writeFile(result.RunDockerfile, context.OutputDirectory, "run.Dockerfile")
		if err != nil {
			return libcnb.GenerateResult{}, err
		}
	}
	err = writeTOML(result.ExtendConfig, context.OutputDirectory, "extend-config.toml")
	if err != nil {
		if err != nil {
			return libcnb.GenerateResult{}, err
		}
	}

	//return generateResult
	return result.GenerateResult, nil
}

func writeFile(content io.Reader, dirname string, filename string) error {

	file, err := os.Create(filepath.Join(dirname, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	if err != nil {
		return err
	}

	return nil
}

func writeTOML(content interface{}, dirname string, filename string) error {

	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(content)
	if err != nil {
		return err
	}
	return writeFile(bytes.NewReader(buf.Bytes()), dirname, filename)
}

//TODO: add util method here to allow an extension to have a companion buildpack to perform layer based config of a jvm installed via dockerfiles.
//      needs thought, re how companion will know what to do etc.
