// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package dnsservice

import (
	"embed"
	"fmt"
	"path"

	"github.com/gardener/gardener/pkg/utils"
	"sigs.k8s.io/yaml"

	"github.com/gardener/gardener-landscape-kit/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/components"
	"github.com/gardener/gardener-landscape-kit/pkg/ocm/components/helpers"
	utilscomponentvector "github.com/gardener/gardener-landscape-kit/pkg/utils/componentvector"
	"github.com/gardener/gardener-landscape-kit/pkg/utils/files"
)

const (
	// ComponentDirectory is the garden component directory within the base components directory.
	ComponentDirectory = "gardener-extensions/shoot-dns-service"
)

var (
	// baseTemplateDir is the directory where the base templates are stored.
	baseTemplateDir = "templates/base"
	//go:embed templates/base
	baseTemplates embed.FS

	// landscapeTemplateDir is the directory where the landscape templates are stored.
	landscapeTemplateDir = "templates/landscape"
	//go:embed templates/landscape
	landscapeTemplates embed.FS
)

type component struct{}

// NewComponent creates a new garden component.
func NewComponent() components.Interface {
	return &component{}
}

// Name returns the component name.
func (c *component) Name() string {
	return "shoot-dns-service"
}

// GenerateBase generates the component base directory.
func (c *component) GenerateBase(options components.Options) error {
	for _, op := range []func(components.Options) error{
		writeBaseTemplateFiles,
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

// GenerateLandscape generates the component landscape directory.
func (c *component) GenerateLandscape(options components.LandscapeOptions) error {
	for _, op := range []func(components.LandscapeOptions) error{
		writeLandscapeTemplateFiles,
	} {
		if err := op(options); err != nil {
			return err
		}
	}
	return nil
}

func getTemplateValues(opts components.Options) (map[string]any, error) {
	return components.GetComponentVectorTemplateValues(opts, componentvector.NameGardenerGardenerExtensionShootDnsService)
}

func writeBaseTemplateFiles(opts components.Options) error {
	objects, err := files.RenderTemplateFiles(baseTemplates, baseTemplateDir, nil)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem(), opts.GetMergeMode())
}

func writeLandscapeTemplateFiles(opts components.LandscapeOptions) error {
	relativeComponentPath := path.Join(components.DirName, ComponentDirectory)

	renderValue, err := getTemplateValues(opts)
	if err != nil {
		return err
	}
	dnsControllerManagerImageValue, err := addDNSControllerManagerImageValue(renderValue)
	if err != nil {
		return err
	}
	values := utils.MergeMaps(renderValue, map[string]any{
		"dnsControllerManagerImage":   dnsControllerManagerImageValue,
		"landscapeComponentPath":      path.Join(opts.GetRelativeLandscapePath(), relativeComponentPath),
		"relativePathToBaseComponent": opts.GetRelativeBaseComponentPath(ComponentDirectory),
	})
	objects, err := files.RenderTemplateFiles(landscapeTemplates, landscapeTemplateDir, values)
	if err != nil {
		return err
	}

	return files.WriteObjectsToFilesystem(objects, opts.GetTargetPath(), path.Join(components.DirName, ComponentDirectory), opts.GetFilesystem(), opts.GetMergeMode())
}

func addDNSControllerManagerImageValue(renderValue map[string]any) (map[string]any, error) {
	imageVectorOverwrite, ok := renderValue["imageVectorOverwrite"]
	if !ok {
		return nil, nil
	}
	data, ok := imageVectorOverwrite.(string)
	if !ok {
		return nil, fmt.Errorf("imageVectorOverwrite is not a string")
	}
	var obj utilscomponentvector.ImageVectorOverwrite
	if err := yaml.Unmarshal([]byte(data), &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal imageVectorOverwrite: %w", err)
	}
	for _, img := range obj.Images {
		if img.Name == "dns-controller-manager" {
			repository, tag, err := helpers.RepoTagFromRefOrParts(img.Repository, img.Tag, img.Ref)
			if err != nil {
				return nil, fmt.Errorf("invalid image %q: %w", img.Name, err)
			}
			return map[string]any{
				"repository": repository,
				"tag":        tag,
			}, nil
		}
	}
	return nil, nil
}
