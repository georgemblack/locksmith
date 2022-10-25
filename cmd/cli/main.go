package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/georgemblack/locksmith/pkg/locksmith"
)

const (
	leaderRole   = "leader"
	followerRole = "follower"
)

func main() {
	args := os.Args[1:]
	if !validArgs(args) {
		fmt.Println("Usage: locksmith <leader|follower> <vault url>")
		os.Exit(1)
	}
	role := args[0]
	vaultURL := args[1]

	fmt.Println("🔐 Welcome to Locksmith!")
	printProgressBar()

	var err error
	if role == leaderRole {
		err = executeLeaderTrack(vaultURL)
	} else {
		err = executeFollowerTrack(vaultURL)
	}
	if err != nil {
		printError(err)
		os.Exit(1)
	}
}

func executeLeaderTrack(vaultURL string) error {
	// Check for existing rekey operation
	status, err := locksmith.GetRekeyStatus(vaultURL)
	if err != nil {
		return locksmith.WrapError(err, "failed to get rekey status")
	}
	if status.InProgress() {
		return errors.New("a rekey operation is already in progress, please cancel operation before starting a new one")
	}

	fmt.Println("Starting a new rekey operation.")

	// Build & submit request to start new rekey
	rekeyRequest := locksmith.PromptRekeyOptions()
	status, err = locksmith.StartRekey(vaultURL, rekeyRequest)
	if err != nil {
		return locksmith.WrapError(err, "failed to start rekey operation")
	}

	fmt.Printf("Rekey operation started. %d key shares must be provided.\n", status.Required)

	// Wait for all other participants to submit their keys before asking for the leader's key
	locksmith.WaitForParticipantRekeySubmissions(vaultURL)

	// Submit leader's key
	// Poll for input until a valid key is submitted
	for {
		status, err = locksmith.SubmitKey(vaultURL, locksmith.PromptKeyShare())
		if err != nil {
			// If invalid keys were submitted, exit. Process will need to be restarted.
			if status.InvalidKeysError() {
				return errors.New("invalid keys submitted, please cancel rekey and try again")
			}
			printError(locksmith.WrapError(err, "failed to submit key"))
			continue
		}
		if len(status.Keys) == 0 {
			return errors.New("no keys returned from vault, please cancel rekey and try again")
		}
		break
	}

	// Save new recovery keys to disk
	err = locksmith.WriteKeysToFile(vaultURL, locksmith.WriteKeysToFileRequest{
		KeybaseUsers:    rekeyRequest.KeybaseUsers,
		PGPFingerprints: status.PGPFingerprints,
		Keys:            status.Keys,
		KeysBase64:      status.KeysBase64,
	})
	if err != nil {
		return locksmith.WrapError(err, "failed to generate key file")
	}

	fmt.Println("Verification has begun. Please wait for other participants to submit their keys.")

	locksmith.WaitForParticipantVerificationSubmissions(vaultURL)

	// Submit verification key
	// Validate a successful status in response
	finalStatus, err := locksmith.SubmitVerification(vaultURL, locksmith.PromptKeyShare())
	if err != nil {
		return locksmith.WrapError(err, "failed to submit final verification")
	}
	if !finalStatus.Completed() || status.HasError() {
		return locksmith.WrapError(err, "rekey verification failed")
	}

	fmt.Println("✅ Vault rekey operation complete. New keys have been verified. You're done!")

	return nil
}

func executeFollowerTrack(vaultURL string) error {
	// Check for existing rekey operation
	status, err := locksmith.GetRekeyStatus(vaultURL)
	if err != nil {
		return locksmith.WrapError(err, "failed to get rekey status")
	}

	if !status.InProgress() {
		locksmith.WaitForRekeyStart(vaultURL)
	} else {
		fmt.Println("A rekey operation is in-progress. Please enter your key share.")
	}

	// Submit key share
	// Poll for input until a valid key is submitted
	for {
		_, err = locksmith.SubmitKey(vaultURL, locksmith.PromptKeyShare())
		if err != nil {
			printError(locksmith.WrapError(err, "failed to submit key"))
			continue
		}
		break
	}

	fmt.Println("Key submitted successfully. Waiting for other participants to submit their keys.")

	locksmith.WaitForRekeyCompletion(vaultURL)

	fmt.Println("Verification has begun. Please enter your new key share to verify.")

	// Submit verification
	// Poll for input until a valid key is submitted
	for {
		_, err = locksmith.SubmitVerification(vaultURL, locksmith.PromptKeyShare())
		if err != nil {
			printError(locksmith.WrapError(err, "failed to submit verification"))
			continue
		}
		break
	}

	fmt.Println("Key verification submitted successfully. Waiting for other participants to submit their keys.")

	locksmith.WaitForVerificationCompletion(vaultURL)

	fmt.Println("☑️  Operation complete. Any potential errors will be returned to the leader.")

	return nil
}

func validArgs(args []string) bool {
	if len(args) != 2 {
		return false
	}
	if !(args[0] == leaderRole || args[0] == followerRole) {
		return false
	}
	return true
}

func printError(err error) {
	fmt.Printf("🚫 Error: %s\n", err.Error())
}

func printProgressBar() {
	for i := 0; i < 63; i++ {
		fmt.Printf("\r%s🔑", strings.Repeat("=", i))
		time.Sleep(10 * time.Millisecond)
	}
	fmt.Printf("\r%s\n", strings.Repeat("=", 64))
}
