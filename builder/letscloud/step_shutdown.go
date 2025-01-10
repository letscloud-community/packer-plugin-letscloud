package letscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/letscloud-community/letscloud-go"
)

// StepShutdown shuts down the instance after provisioning.
type StepShutdown struct {
	sdkClient *letscloud.LetsCloud
	config    *Config
}

// Run executes the StepShutdown.
func (s *StepShutdown) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	// Retrieve the instance ID from state
	identifier, ok := state.GetOk("instance_identifier")
	if !ok {
		ui.Say("No instance found to shut down.")
		return multistep.ActionContinue
	}
	instanceID := identifier.(string)

	ui.Say(fmt.Sprintf("Shutting down instance: %s", instanceID))

	// Call the LetsCloud API to power off the instance
	// Increase timeout to 60 seconds
	if err := s.sdkClient.SetTimeout(60 * time.Second); err != nil {
		ui.Error(fmt.Sprintf("Failed to set timeout: %v", err))
	}
	err := s.sdkClient.PowerOffInstance(instanceID)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to shut down instance %s: %s", instanceID, err))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Instance %s has been powered off successfully.", instanceID))
	return multistep.ActionContinue
}

func (s *StepShutdown) Cleanup(state multistep.StateBag) {
	// No cleanup needed after shutdown
}
