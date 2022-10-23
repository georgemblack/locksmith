package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
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

	var err error
	if role == leaderRole {
		err = executeLeaderTrack(vaultURL)
	} else {
		err = executeFollowerTrack(vaultURL)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("✅ Vault rekey operation complete. New keys have been verified. You're done!")
}

func executeLeaderTrack(vaultURL string) error {
	// Check for existing rekey operation
	status, err := locksmith.GetRekeyStatus(vaultURL)
	if err != nil {
		return err
	}
	if status.InProgress() {
		return errors.New("a rekey operation is already in progress, please cancel operation before starting a new one")
	}

	fmt.Println("Starting a new rekey operation.")

	// Build & submit rekey request
	rekeyRequest := locksmith.PromptRekeyOptions()
	status, err = locksmith.StartRekey(vaultURL, rekeyRequest)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Rekey operation started. %d key shares must be provided.\n", status.Required)

	// Wait for all other participants to submit their keys before asking for the leader's key
	locksmith.WaitForParticipantRekeySubmissions(vaultURL)

	// Submit leader's key
	status, err = locksmith.SubmitKey(vaultURL, locksmith.PromptKeyShare())
	if err != nil {
		return err
	}
	if len(status.Keys) == 0 {
		return errors.New("no keys returned from Vault")
	}

	// Save new recovery keys to disk
	output := fmt.Sprintf("VAULT URL: %s\n\n", vaultURL)
	for i, key := range status.Keys {
		user := rekeyRequest.KeybaseUsers[i]
		fingerprint := status.PGPFingerprints[i]
		keyBase64 := status.KeysBase64[i]
		output += fmt.Sprintf("KEYBASE USER: %s\nFINGERPRINT: %s\nENCRYPTED_KEY: %s\nENCRYPTED_KEY_BASE64: %s\n\n", user, fingerprint, key, keyBase64)
	}
	fileName := fmt.Sprintf("recovery-keys-%s.txt", time.Now().Format("2006-01-02-15-04-05"))
	err = ioutil.WriteFile(fileName, []byte(output), 0644)
	if err != nil {
		return err
	}

	fmt.Printf("✍️  New recovery keys saved to: %s\n", fileName)
	fmt.Println("Verification has begun. Please enter your new key share to verify.")

	// Submit verification key
	_, err = locksmith.SubmitVerification(vaultURL, locksmith.PromptKeyShare())
	if err != nil {
		return err
	}

	fmt.Println("Key verification submitted successfully. Waiting for other participants to submit their keys.")

	locksmith.WaitForVerificationCompletion(vaultURL)

	return nil
}

func executeFollowerTrack(vaultURL string) error {
	// Check for existing rekey operation
	status, err := locksmith.GetRekeyStatus(vaultURL)
	if err != nil {
		return err
	}

	if !status.InProgress() {
		locksmith.WaitForRekeyStart(vaultURL)
	} else {
		fmt.Println("A rekey operation is in-progress. Please enter your key share.")
	}

	_, err = locksmith.SubmitKey(vaultURL, locksmith.PromptKeyShare())
	if err != nil {
		return err
	}

	fmt.Println("Key submitted successfully. Waiting for other participants to submit their keys.")

	locksmith.WaitForRekeyCompletion(vaultURL)

	fmt.Println("Verification has begun. Please enter your new key share to verify.")

	_, err = locksmith.SubmitVerification(vaultURL, locksmith.PromptKeyShare())
	if err != nil {
		return err
	}

	fmt.Println("Key verification submitted successfully. Waiting for other participants to submit their keys.")

	locksmith.WaitForVerificationCompletion(vaultURL)

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
