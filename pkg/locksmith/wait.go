package locksmith

import (
	"fmt"
	"time"
)

func WaitForRekeyStart(vaultURL string) {
	count := 0
	for {
		status, err := GetRekeyStatus(vaultURL)
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
			continue
		}
		emoji := getEmoji(count)

		if !status.InProgress() {
			fmt.Printf("\r\033[K%s Waiting for rekey operation to start.", emoji)
		}
		if status.InProgress() {
			fmt.Printf("\r\033[K🙌 Rekey operation started. Please enter your key share.\n")
			break
		}

		count += 1
		time.Sleep(1 * time.Second)
	}
}

func WaitForRekeyCompletion(vaultURL string) {
	count := 0
	rekeyStarted := false
	verificationStarted := false
	for {
		status, err := GetRekeyStatus(vaultURL)
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
			continue
		}
		if status.Started {
			rekeyStarted = true
		}
		if status.VerificationNonce != "" {
			verificationStarted = true
		}
		emoji := getEmoji(count)

		if !rekeyStarted {
			fmt.Printf("\r\033[K%s Waiting for rekey to start...", emoji)
		}
		if rekeyStarted && !verificationStarted {
			fmt.Printf("\r\033[K%s %d/%d shares provided. Waiting for other participants to submit their keys.", emoji, status.Progress, status.Required)
		}
		if rekeyStarted && verificationStarted {
			fmt.Printf("\r\033[K🙌 %d/%d shares provided.\n", status.Required, status.Required)
			break
		}

		count += 1
		time.Sleep(1 * time.Second)
	}
}

func WaitForVerificationCompletion(vaultURL string) {
	count := 0
	for {
		status, err := GetVerificationStatus(vaultURL)
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
			continue
		}
		emoji := getEmoji(count)

		if status.InProgress() {
			fmt.Printf("\r\033[K%s %d/%d shares verified. Waiting for other participants to verify their keys.", emoji, status.Progress, status.Threshold)
		}
		if !status.InProgress() {
			fmt.Printf("\r\033[K🙌 All shares verified.\n")
			break
		}

		count += 1
		time.Sleep(1 * time.Second)
	}
}

func WaitForParticipantVerificationSubmissions(vaultURL string) {
	count := 0
	for {
		status, err := GetVerificationStatus(vaultURL)
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
			continue
		}
		emoji := getEmoji(count)

		if !status.InProgress() {
			fmt.Printf("\r\033[K%s Waiting for verification to start...", emoji)
		}
		if status.InProgress() && status.RemainingKeys() != 1 {
			fmt.Printf("\r\033[K%s %d/%d shares verified. You will be prompted for the final share.", emoji, status.Progress, status.Threshold)
		}
		if status.InProgress() && status.RemainingKeys() == 1 {
			fmt.Printf("\r\033[K🙌 %d/%d shares verified. Please provide the final share.\n", status.Progress, status.Threshold)
			break
		}

		count += 1
		time.Sleep(1 * time.Second)
	}
}

func WaitForParticipantRekeySubmissions(vaultURL string) {
	count := 0
	rekeyStarted := false
	for {
		status, err := GetRekeyStatus(vaultURL)
		if err != nil {
			fmt.Printf("\n%s\n", err.Error())
			continue
		}
		if status.InProgress() {
			rekeyStarted = true
		}
		emoji := getEmoji(count)

		if !rekeyStarted {
			fmt.Printf("\r\033[K%s Waiting for rekey to start...", emoji)
		}
		if rekeyStarted && status.RemainingKeys() != 1 {
			fmt.Printf("\r\033[K%s %d/%d shares provided. You will be prompted for the final share.", emoji, status.Progress, status.Required)
		}
		if rekeyStarted && status.RemainingKeys() == 1 {
			fmt.Printf("\r\033[K🙌 %d/%d shares provided. Please provide the final share.\n", status.Progress, status.Required)
			break
		}

		count += 1
		time.Sleep(1 * time.Second)
	}
}

func progressEmojis() []string {
	return []string{"🤔", "🤨", "🧐", "🤓", "🤩"}
}

func getEmoji(count int) string {
	return progressEmojis()[count%len(progressEmojis())]
}
