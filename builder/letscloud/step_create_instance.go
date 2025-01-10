package letscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/letscloud-community/letscloud-go"
	"github.com/letscloud-community/letscloud-go/domains"
)

type StepCreateInstance struct {
	sdkClient *letscloud.LetsCloud
	config    *Config
}

// Run executes the StepCreateInstance.
func (s *StepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Retrieve the Packer UI interface for user interactions.
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating a new instance...")

	// Retrieve SSHSlug and Password from the configuration.
	sshSlug := state.Get("slugKey").(string)
	password, err := generateRandomPassword(16) // Generates a 16-character password.
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to generate password: %s", err))
		state.Put("error", err)
		return multistep.ActionHalt
	} else {
		state.Put("generated_password", password)
	}
	ui.Say("Generated password: " + password)
	timestamp := time.Now().Unix()
	if s.config.Label == "" {
		s.config.Label = fmt.Sprintf("packer-%d", timestamp)
	} else {
		s.config.Label = fmt.Sprintf("%s-%d", s.config.Label, timestamp)
	}
	if s.config.Hostname == "" {
		s.config.Hostname = fmt.Sprintf("packer-%d", timestamp)
	}

	// Define the parameters for creating the instance based on the configuration.
	createReq := &domains.CreateInstanceRequest{
		LocationSlug: s.config.LocationSlug,
		PlanSlug:     s.config.PlanSlug,
		Hostname:     s.config.Hostname,
		Label:        s.config.Label,
		ImageSlug:    s.config.ImageSlug,
		SSHSlug:      sshSlug,
		Password:     password,
	}

	// Create the instance using the SDK client.
	err = s.sdkClient.CreateInstance(createReq)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to create instance: %s", err))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Instance created successfully.")

	// Wait for the instance to be built and retrieve its details.
	createdInstance, err := waitForInstanceCreation(ui, s.sdkClient, s.config.Label, s.config.Hostname, 300*time.Second)
	if err != nil {
		ui.Error(fmt.Sprintf("Error retrieving created instance: %s", err))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Instance Details:\nIdentifier: %s\nIP: %s", createdInstance.Identifier, createdInstance.IPAddresses[0].Address))

	// Store the instance details in the state bag for later use.
	state.Put("instance_identifier", createdInstance.Identifier)
	state.Put("instance_ip", createdInstance.IPAddresses[0].Address)

	return multistep.ActionContinue

}

// Cleanup is called after Run completes, whether it succeeded or failed.
func (s *StepCreateInstance) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	if s.config.KeepInstance {
		ui.Say("Skipping instance cleanup as per configuration.")
		return
	}

	identifier, ok := state.GetOk("instance_identifier")
	if !ok {
		ui.Say("No instance found to clean up.")
		return
	}
	instanceID := identifier.(string)

	ui.Say(fmt.Sprintf("Destroying instance: %s", instanceID))

	err := s.sdkClient.DeleteInstance(instanceID)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete instance %s: %s", instanceID, err))
	} else {
		ui.Say(fmt.Sprintf("Instance %s deleted successfully.", instanceID))
	}
}
