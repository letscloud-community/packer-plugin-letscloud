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

// generateRandomPassword creates a secure random password of the specified length.
func generateRandomPassword(length int) (string, error) {
	if length < 8 {
		return "", errors.New("length must be at least 8")
	}

	charsets := []string{
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"0123456789",
		"!@#$*={}",
	}

	password := make([]byte, length)

	for i, charset := range charsets {
		index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		password[i] = charset[index.Int64()]
	}

	allChars := charsets[0] + charsets[1] + charsets[2] + charsets[3]
	for i := 4; i < length; i++ {
		index, _ := rand.Int(rand.Reader, big.NewInt(int64(len(allChars))))
		password[i] = allChars[index.Int64()]
	}

	for i := range password {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(length)))
		password[i], password[j.Int64()] = password[j.Int64()], password[i]
	}

	return string(password), nil
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
					// Instance is found but not yet built or still locked/suspended; continue waiting.
					ui.Say("Instance is found but not yet built or still locked/suspended. Waiting...")
				}
			}
		case <-timeoutChan:
			return nil, fmt.Errorf("timed out waiting for instance to be built")
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
