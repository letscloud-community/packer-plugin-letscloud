package letscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/letscloud-community/letscloud-go"
)

// StepCreateSSHKey defines a step to create a new SSH key.
type StepCreateSSHKey struct {
	sdkClient *letscloud.LetsCloud
	config    *Config
}

// Run executes the StepCreateSSHKey.
func (s *StepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if s.config.SSHSlug != "" {
		ui.Say("Using provided SSH key slug: " + s.config.SSHSlug)
		state.Put("ssh_key_slug", s.config.SSHSlug)

		return multistep.ActionContinue
	}

	ui.Say("Creating a new SSH key...")

	timestamp := time.Now().Unix()
	sshKeyTitle := fmt.Sprintf("packer-ssh-key-%d", timestamp)

	sshKey, err := s.sdkClient.NewSSHKey(sshKeyTitle, "")
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to create SSH key: %s", err))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	s.config.Comm.SSHPrivateKeyFile = savePrivateKeyToFile(sshKey.PrivateKey)
	ui.Say("SSH key created successfully.")

	// Store SSH key details in the state bag for later use.
	state.Put("slugKey", sshKey.Slug)
	state.Put("publicKey", sshKey.PublicKey)
	state.Put("privateKey", sshKey.PrivateKey)

	return multistep.ActionContinue
}

// Cleanup is called after Run completes, whether it succeeded or failed.
func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	if s.config.SSHSlug != "" {
		ui.Say("Used provided SSH key; nothing to clean up.")
		return
	}

	sshKeySlug, ok := state.GetOk("slugKey")
	if !ok {
		// SSH key was not created; nothing to clean up.
		return
	}

	ui.Say("Deleting temporary SSH key...")

	err := s.sdkClient.DeleteSSHKey(sshKeySlug.(string))

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete SSH key (slug: %s): %s", sshKeySlug, err))
	} else {
		ui.Say(fmt.Sprintf("SSH key (slug: %s) deleted successfully.", sshKeySlug))
	}
}
