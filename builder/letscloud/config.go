//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package letscloud

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// Default state timeout duration
const (
	defaultStateTimeout = 10 * time.Minute
	defaultSSHUsername  = "root"
	defaultCommunicator = "ssh"
)

// Config represents the configuration for the LetsCloud builder.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Comm                communicator.Config `mapstructure:",squash"`
	ctx                 interpolate.Context

	APIKey       string `mapstructure:"api_key"`
	LocationSlug string `mapstructure:"location_slug"`
	PlanSlug     string `mapstructure:"plan_slug"`
	ImageSlug    string `mapstructure:"image_slug"`
	SSHSlug      string `mapstructure:"ssh_slug"`
	Hostname     string `mapstructure:"hostname"`
	Label        string `mapstructure:"label"`
	SnapshotName string `mapstructure:"snapshot_name"`
	StateTimeout string `mapstructure:"state_timeout,omitempty"` // Optional: Defaults to 10m
	KeepInstance bool   `mapstructure:"keep_instance"`           // Optional: Defaults to false
}

// Prepare decodes the configuration and validates required fields.
func (c *Config) Prepare(raws ...interface{}) error {
	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if c.Comm.Type == "" {
		c.Comm.Type = defaultCommunicator
	}

	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = defaultSSHUsername
	}

	if c.Comm.SSHPort == 0 {
		c.Comm.SSHPort = 22
	}

	if !c.KeepInstance {
		c.KeepInstance = false
	}

	// Initialize a MultiError to collect all validation errors
	var errs *packer.MultiError

	// Validate required fields
	requiredFields := map[string]string{
		"api_key":       c.APIKey,
		"location_slug": c.LocationSlug,
		"plan_slug":     c.PlanSlug,
		"image_slug":    c.ImageSlug,
	}

	for field, value := range requiredFields {
		if value == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("`%s` is required", field))
		}
	}

	// Validate StateTimeout format or set default
	if c.StateTimeout == "" {
		c.StateTimeout = defaultStateTimeout.String()
	} else {
		if _, err := time.ParseDuration(c.StateTimeout); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("invalid format for `state_timeout`: %s", err))
		}
	}

	// Prepare the communicator
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		log.Println("*** Prepare Comm ***")
		errs = packer.MultiErrorAppend(errs, es...)
	}

	// Return all accumulated errors, if any
	if errs != nil && len(errs.Errors) > 0 {
		log.Println("*** ERRORS errs ***")
		return errs
	}

	// Protect the APIKey from being logged
	packer.LogSecretFilter.Set(c.APIKey)

	return nil
}

// ConfigSpec returns the HCL object spec for the configuration.
func (c *Config) ConfigSpec() hcldec.ObjectSpec {
	return c.FlatMapstructure().HCL2Spec()
}
