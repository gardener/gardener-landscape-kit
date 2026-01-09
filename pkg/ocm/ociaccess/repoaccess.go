// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package ociaccess

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"

	"github.com/go-logr/logr"
	"k8s.io/component-base/version"
	descriptorruntime "ocm.software/open-component-model/bindings/go/descriptor/runtime"
	descriptorv2 "ocm.software/open-component-model/bindings/go/descriptor/v2"
	"ocm.software/open-component-model/bindings/go/oci"
	urlresolver "ocm.software/open-component-model/bindings/go/oci/resolver/url"
	ocmoci "ocm.software/open-component-model/bindings/go/oci/spec/access"
	ocmruntime "ocm.software/open-component-model/bindings/go/runtime"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const userAgentPrefix = "gardener-landscape-kit/"

// DefaultScheme is the scheme used by RepoAccess.
var DefaultScheme = ocmruntime.NewScheme()

func init() {
	ocmoci.MustAddToScheme(DefaultScheme)
	descriptorv2.MustAddToScheme(DefaultScheme)
}

// RepoAccess provides access to an OCI repository, allowing retrieval of component versions.
type RepoAccess struct {
	Name          string
	RepositoryURL string
	repo          *oci.Repository
	logOutput     *bytes.Buffer
}

// NewRepoAccess creates a new RepoAccess instance for accessing an OCI repository.
func NewRepoAccess(repositoryURL string) (*RepoAccess, error) {
	resolver, err := urlresolver.New(urlresolver.WithBaseURL(repositoryURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create URL resolver: %w", err)
	}

	// TODO (MartinWeindel): Support other authentication methods.
	user := "_json_key"
	password := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	resolver.SetClient(CreateAuthClient(repositoryURL, user, password))

	logOutput := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(logOutput, &slog.HandlerOptions{Level: slog.LevelInfo}))
	repo, err := oci.NewRepository(oci.WithResolver(resolver), oci.WithScheme(DefaultScheme), oci.WithLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("failed on NewRepository: %w", err)
	}

	return &RepoAccess{
		RepositoryURL: repositoryURL,
		repo:          repo,
		logOutput:     logOutput,
	}, nil
}

// GetComponentVersion retrieves the component descriptor for a specific component version from the repository.
func (r *RepoAccess) GetComponentVersion(ctx context.Context, component, version string) (*descriptorruntime.Descriptor, error) {
	r.logOutput.Reset()
	descriptor, err := r.repo.GetComponentVersion(ctx, component, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get component version %s:%s from repository %s: %w", component, version, r.RepositoryURL, err)
	}
	return descriptor, nil
}

// GetLocalResource retrieves a local resource for a specific component version and identity from the repository.
func (r *RepoAccess) GetLocalResource(ctx context.Context, component, version string, identity map[string]string) ([]byte, error) {
	r.logOutput.Reset()
	blob, _, err := r.repo.GetLocalResource(ctx, component, version, identity)
	if err != nil {
		return nil, fmt.Errorf("failed to get local resource for component version %s:%s from repository %s: %w", component, version, r.RepositoryURL, err)
	}
	reader, err := blob.ReadCloser()
	if err != nil {
		return nil, fmt.Errorf("failed to get read closer for blob: %w", err)
	}
	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob data: %w", err)
	}
	if err := reader.Close(); err != nil {
		return nil, fmt.Errorf("failed to close blob reader: %w", err)
	}
	return buf.Bytes(), nil
}

// FindComponentVersion searches for a specific component version across multiple repositories.
// It returns the descriptor and local blobs if any are found for the specified localBlobResourceTypes.
func FindComponentVersion(
	ctx context.Context,
	log logr.Logger,
	repos []*RepoAccess,
	component, version string,
	localBlobResourceTypes ...string,
) (*descriptorruntime.Descriptor, map[NameVersionType][]byte, error) {
	logOutputs := &bytes.Buffer{}
	var errs []error
	for _, repo := range repos {
		descriptor, err := repo.GetComponentVersion(ctx, component, version)
		if err == nil {
			// Collect local blobs if requested.
			repoLocalBlobs, err := loadLocalBlobs(ctx, repo, descriptor, localBlobResourceTypes...)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load local blobs for component version %s:%s from repository %s: %w", component, version, repo.RepositoryURL, err)
			}
			return descriptor, repoLocalBlobs, nil
		}
		errs = append(errs, fmt.Errorf("repository %s: %w", repo.RepositoryURL, err))
		logOutputs.Write(repo.logOutput.Bytes())
	}
	log.Info("Failed to find component version in any repository", "component", component, "version", version, "details", logOutputs.String())
	return nil, nil, fmt.Errorf("component version %s:%s not found in any repository: %s", component, version, errors.Join(errs...))
}

// NameVersionType represents the identity of a resource with its name, version, and type.
type NameVersionType struct {
	Name    string
	Version string
	Type    string
}

func loadLocalBlobs(ctx context.Context, repo *RepoAccess, descriptor *descriptorruntime.Descriptor, localBlobResourceTypes ...string) (map[NameVersionType][]byte, error) {
	if len(localBlobResourceTypes) == 0 {
		return nil, nil
	}
	localBlobs := make(map[NameVersionType][]byte)
	for _, res := range descriptor.Component.Resources {
		if !stringSliceContains(localBlobResourceTypes, res.Type) {
			continue
		}
		data, err := repo.GetLocalResource(ctx, descriptor.Component.Name, descriptor.Component.Version, res.ToIdentity())
		if err != nil {
			return nil, fmt.Errorf("failed to get local blob for resource %s of type %s: %w", res.Name, res.Type, err)
		}
		localBlobs[ResourceToBlobKey(res)] = data
	}
	return localBlobs, nil
}

// ResourceToBlobKey converts a resource to a NameVersionType key.
func ResourceToBlobKey(res descriptorruntime.Resource) NameVersionType {
	return NameVersionType{
		Name:    res.Name,
		Version: res.Version,
		Type:    res.Type,
	}
}

func stringSliceContains(slice []string, s string) bool {
	return slices.Contains(slice, s)
}

// CreateAuthClient creates an authenticated client for accessing OCI repositories.
func CreateAuthClient(address, username, password string) *auth.Client {
	url, err := ocmruntime.ParseURLAndAllowNoScheme(address)
	if err != nil {
		panic(fmt.Sprintf("invalid address %q: %v", address, err))
	}
	return &auth.Client{
		Client: retry.DefaultClient,
		Header: http.Header{
			"User-Agent": []string{userAgentPrefix + version.Get().GitVersion},
		},
		Credential: auth.StaticCredential(url.Host, auth.Credential{
			Username: username,
			Password: password,
		}),
	}
}
