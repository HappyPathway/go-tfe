package tfe

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ RegistryModules = (*registryModules)(nil)

// OAuthClients describes all the OAuth client related methods that the
// Terraform Enterprise API supports.
//
// TFE API docs:
// https://www.terraform.io/docs/enterprise/api/oauth-clients.html
type RegistryModules interface {
	// Create a VCS connection between an organization and a VCS provider.
	Create(ctx context.Context, organization string, options RegistryModuleCreateOptions) (*RegistryModule, error)
	Delete(ctx context.Context, organization string, module string, provider string) (*RegistryModule, error)
}

// oAuthClients implements OAuthClients.
type registryModules struct {
	client *Client
}

// OAuthClient represents a connection between an organization and a VCS
// provider.
type RegistryModulePermissions struct {
	can_delete string `json:"can-delete"`
	can_resync string `json:"can-resync"`
	can_retry  string `json:"can-retry"`
}

type RegistryModule struct {
	ID       string `jsonapi:"primary,id"`
	Name     string `jsonapi:"attr,name"`
	Token    string `jsonapi:"attr,vcs-repo,oauth-token-id"`
	Repo     string `jsonapi:"attr,vcs-repo,identifier"`
	Provider string `jsonapi:"attr,provider"`

	// Relations
	Organization              *Organization              `jsonapi:"relation,organization"`
	RegistryModulePermissions *RegistryModulePermissions `jsonapi:"attr,permissions"`
}

// OAuthClientCreateOptions represents the options for creating an OAuth client.
type RegistryModuleCreateOptions struct {
	// For internal use only!
	Token *string `jsonapi:"attr,vcs-repo,oauth-token-id"`
	Repo  *string `jsonapi:"attr,vcs-repo,identifier"`
	ID    string  `jsonapi:"primary,id"`
}

func (o RegistryModuleCreateOptions) valid() error {
	if !validString(o.Token) {
		return errors.New("Token is required")
	}
	if !validString(o.Repo) {
		return errors.New("Repo is required")
	}
	return nil
}

// Create a VCS connection between an organization and a VCS provider.
func (s *registryModules) Create(ctx context.Context, organization string, options RegistryModuleCreateOptions) (*RegistryModule, error) {
	if !validStringID(&organization) {
		return nil, errors.New("Invalid value for organization")
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	// Make sure we don't send a user provided ID.
	options.ID = ""

	u := fmt.Sprintf("registry-modules")
	req, err := s.client.newRequest("POST", u, &options)
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
