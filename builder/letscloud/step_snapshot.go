package letscloud

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/letscloud-community/letscloud-go"
)

type StepSnapshot struct {
	sdkClient *letscloud.LetsCloud
	config    *Config
}

// Run executes the StepSnapshot
func (s *StepSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	// Retrieve the instance ID from state
	identifier, ok := state.GetOk("instance_identifier")
	if !ok {
		ui.Say("No instance found to shut down.")
		return multistep.ActionContinue
	}
	instanceID := identifier.(string)

	label := s.config.SnapshotName
	if label == "" {
		label = fmt.Sprintf("packer-snapshot-%d", time.Now().Unix())
	}
	ui.Say(fmt.Sprintf("Requesting snapshot for instance '%s' with label '%s'...", instanceID, label))

	if err := s.sdkClient.SetTimeout(60 * time.Second); err != nil {
		ui.Error(fmt.Sprintf("Failed to set timeout: %v", err))
	}
	snapshot, err := s.sdkClient.NewSnapshot(label, instanceID)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to create snapshot %s", err))
		state.Put("error", err)
		return multistep.ActionHalt
	}
	slug := snapshot.Data.Slug
	ui.Say(fmt.Sprintf("Snapshot '%s' creation has been queued. Waiting for it to finish...", slug))

	err = waitForSnapshotCreation(ui, s.sdkClient, slug, 10*time.Minute)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("snapshot_name", s.config.SnapshotName)
	state.Put("snapshot_slug", slug)

	return multistep.ActionContinue
}

func (s *StepSnapshot) Cleanup(state multistep.StateBag) {
	// No cleanup needed
}
