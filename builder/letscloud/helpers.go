// helpers.go
package letscloud

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"crypto/rand"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/letscloud-community/letscloud-go"
	"github.com/letscloud-community/letscloud-go/domains"
)

// generateRandomPassword generates a secure password with at least one lowercase letter, one uppercase letter, one number, and one special character.
func generateRandomPassword(length int) (string, error) {
	if length < 8 {
		return "", errors.New("length must be at least 8")
	}

	// Character sets
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers := "0123456789"
	specials := "!@#$={}"
	allChars := lowercase + uppercase + numbers + specials

	// Ensure at least one character from each set
	password := make([]byte, length)
	charsets := []string{lowercase, uppercase, numbers, specials}

	// Ensure at least one character from each required set
	for i, charset := range charsets {
		index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[index.Int64()]
	}

	// Fill the remaining characters randomly
	for i := 4; i < length; i++ {
		index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		password[i] = allChars[index.Int64()]
	}

	// Shuffle the password to avoid predictable placements
	shuffle(password)

	return string(password), nil
}

// shuffle randomly shuffles a byte slice in-place
func shuffle(password []byte) {
	for i := range password {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(len(password))))
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}
}

// waitForInstanceCreation polls the Instances API to find the created instance.
// It waits until the instance is built and not locked or suspended.
// Returns the instance if found within the timeout period.
func waitForInstanceCreation(ui packer.Ui, sdkClient *letscloud.LetsCloud, label string, hostname string, timeout time.Duration) (*domains.Instance, error) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			instances, err := sdkClient.Instances()
			if err != nil {
				ui.Error(fmt.Sprintf("Failed to list instances: %s", err))
				return nil, err
			}

			for _, inst := range instances {
				if inst.Label == label && inst.Hostname == hostname {
					// Instance matches label and hostname.
					// Check if it is built and not locked or suspended.
					if inst.Built && !inst.Locked && !inst.Suspended {
						return &inst, nil
					}
					ui.Message("Instance is not yet built or still locked. Waiting...")
				}
			}
		case <-timeoutChan:
			return nil, fmt.Errorf("timed out waiting for instance to be built")
		}
	}
}

func waitForSnapshotCreation(ui packer.Ui, sdkClient *letscloud.LetsCloud, slug string, timeout time.Duration) error {
	ui.Say(fmt.Sprintf("Waiting for snapshot '%s' to finish building...", slug))

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			snapshot, err := sdkClient.Snapshot(slug)
			if err != nil {
				ui.Message(fmt.Sprintf("Error checking snapshot status: %s", err))
				continue
			}

			if snapshot.Build {
				ui.Say(fmt.Sprintf("Snapshot '%s' is now ready.", slug))
				return nil
			}
			ui.Message("Snapshot still building...")

		case <-timeoutChan:
			return fmt.Errorf("snapshot '%s' did not finish building within %s", slug, timeout.String())
		}
	}
}

// savePrivateKeyToFile saves the private key to a temporary file and returns its path.
func savePrivateKeyToFile(privateKey string) string {
	timestamp := time.Now().Unix()
	keyPath := fmt.Sprintf("/tmp/packer_ssh_key_%d.pem", timestamp)

	err := os.WriteFile(keyPath, []byte(privateKey), 0600)
	if err != nil {
		log.Fatalf("Error saving private key: %s", err)
	}

	log.Printf("Private key saved at: %s", keyPath)
	return keyPath
}
