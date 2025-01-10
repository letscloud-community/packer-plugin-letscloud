package letscloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/letscloud-community/letscloud-go"
)

const BuilderId = "packer.letscloud"

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, nil, errs
	}

	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	ui.Say("Running LetsCloud builder...")

	sdkClient, err := letscloud.New(b.config.APIKey)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize LetsCloud client: %v", err)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("sdkClient", sdkClient)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&StepCreateSSHKey{
			sdkClient: sdkClient,
			config:    &b.config,
		},
		&StepCreateInstance{
			sdkClient: sdkClient,
			config:    &b.config,
		},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      communicator.CommHost(b.config.Comm.Host(), "instance_ip"),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&StepShutdown{
			sdkClient: sdkClient,
			config:    &b.config,
		},
	}

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	artifact := &Artifact{
		// Add the builder generated data to the artifact StateData so that post-processors
		// can access them.
		StateData: map[string]interface{}{
			"instance_identifier": state.Get("instance_identifier"),
			"instance_ip":         state.Get("instance_ip"),
			"generated_password":  state.Get("generated_password"),
		},
	}
	return artifact, nil
}
