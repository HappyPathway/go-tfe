package tfe

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ RegistryModules = (*registryModules)(nil)

// RegistryModules describes all the Registry Module related methods that the
// Terraform Enterprise API supports.
//
// TFE API docs:
// https://www.terraform.io/docs/enterprise/api/modules.html
//
type RegistryModules interface {
	// Create a VCS connection between an organization and a VCS provider.
	Create(ctx context.Context, options RegistryModuleCreateOptions) (*RegistryModule, error)
	Delete(ctx context.Context, organization string, module string, provider string) (*RegistryModule, error)
}

// registryModules implements RegistryModule.
type registryModules struct {
	client *Client
}

type RegistryModule struct {
	ID       string `jsonapi:"primary,id"`
	Name     string `jsonapi:"attr,name"`
	Provider string `jsonapi:"attr,provider"`
}

type RegistryModuleVCSRepo struct {
	OAUTH_TOKEN_ID string `json:"oauth-token-id"`
	IDENTIFIER     string `json:"identifier"`
}

type RegistryModuleCreateOptions struct {
	// For internal use only!
	VCSRepo *RegistryModuleVCSRepo `jsonapi:"attr,vcs-repo"`
}

func (o RegistryModuleCreateOptions) valid() error {
	if o.VCSRepo == nil {
		return errors.New("Token is required")
	}
	return nil
}

// Create a VCS connection between an organization and a VCS provider.
//
func (s *registryModules) Create(ctx context.Context, options RegistryModuleCreateOptions) (*RegistryModule, error) {
	if err := options.valid(); err != nil {
		return nil, err
	}

	req, err := s.client.newRequest("POST", "registry-modules", &options)
	if err != nil {
		return nil, err
	}

	rm := &RegistryModule{}
	err = s.client.do(ctx, req, rm)
	if err != nil {
		return nil, err
	}

	return rm, nil
}

func (s *registryModules) Delete(ctx context.Context, organization string, module string, provider string) (*RegistryModule, error) {

	u := fmt.Sprintf("registry-modules/actions/delete/%s/%s/%s", url.QueryEscape(organization), url.QueryEscape(module), url.QueryEscape(provider))

	req, err := s.client.newRequest("DELETE", u, s)
	if err != nil {
		return nil, err
	}

	rm := &RegistryModule{}
	err = s.client.do(ctx, req, nil)
	if err != nil && err.Error() != "resource not found" {
		return nil, err
	}

	if err != nil && err.Error() == "resource not found" {
		return rm, nil
	}

	return rm, nil
}
