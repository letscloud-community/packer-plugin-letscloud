package letscloud

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// Artifact implements the packer.Artifact interface.
type Artifact struct {
	StateData map[string]interface{}
}

// NewArtifact creates a new Artifact.
func NewArtifact(values map[string]interface{}) packer.Artifact {
	return &Artifact{
		StateData: values,
	}
}

// String returns a description of the artifact.
func (a *Artifact) String() string {
	return fmt.Sprintf("Artifact with ID: %s", a.Id())
}

// State returns the state data of the artifact.
func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

// Files returns the files associated with this artifact.
func (a *Artifact) Files() []string {
	return nil
}

// Destroy is a no-op for this artifact.
func (a *Artifact) Destroy() error {
	// Implement any cleanup logic here if necessary.
	return nil
}

func (a *Artifact) Id() string {
	if id, ok := a.StateData["instance_idetifier"].(string); ok {
		return id
	}
	return "unknown"
}

// BuilderId returns the builder ID.
func (a *Artifact) BuilderId() string {
	return BuilderId
}

// Config returns the artifact configuration.
func (a *Artifact) Config() map[string]interface{} {
	return a.StateData
}

// Serialize serializes the artifact to JSON.
func (a *Artifact) Serialize() ([]byte, error) {
	return json.Marshal(a.StateData)
}

// PrintOnUI displays the artifact details on the UI.
// Note: Generated passwords are not displayed to maintain security best practices.
func (a *Artifact) PrintOnUI(ui packer.Ui) error {
	// Retrieve the data from StateData.
	instanceIdentifier, idenfierOk := a.StateData["instance_identifier"].(string)
	instanceIP, ipOk := a.StateData["instance_ip"].(string)

	if !idenfierOk || !ipOk {
		ui.Error("Artifact data is incomplete.")
		return fmt.Errorf("artifact data is incomplete")
	}

	// Display the artifact details without the generated password.
	ui.Say(fmt.Sprintf("Build Artifact:\nInstance Identifier: %s\nInstance IP: %s",
		instanceIdentifier,
		instanceIP,
	))

	// Optionally, inform the user that a password was generated.
	if _, pwdOk := a.StateData["generated_password"].(string); pwdOk {
		ui.Say("A secure password has been generated for the instance.")
	}

	return nil

}
