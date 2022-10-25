package locksmith

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func PromptRekeyOptions() StartRekeyRequest {
	secretShares := promptInt("Number of secret shares")
	secretThreshold := promptInt("Secret threshold")

	var keybaseUsers []string
	for {
		keybaseUsers = strings.Split(promptString("Keybase users"), ",")
		if len(keybaseUsers) != secretShares {
			fmt.Println("Number of keybase users must match secret shares. Please try again.")
			continue
		}
		break
	}

	return StartRekeyRequest{
		SecretShares:    secretShares,
		SecretThreshold: secretThreshold,
		KeybaseUsers:    keybaseUsers,
	}
}

func Prompt(prompt string) string {
	return promptString(prompt)
}

func promptString(prompt string) string {
	var input string
	var err error

	reader := bufio.NewReader(os.Stdin)
	for {
		printPrompt(prompt)
		input, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input: ", err)
			continue
		}
		if input != "" {
			break
		}
	}
	return strings.TrimSpace(input)
}

func promptInt(prompt string) int {
	var input string
	var result int
	var err error

	reader := bufio.NewReader(os.Stdin)
	for {
		printPrompt(prompt)
		input, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input: ", err)
			continue
		}
		input = strings.TrimSpace(input)
		result, err = strconv.Atoi(input)
		if err != nil {
			fmt.Println("Input must be a valid integer. Please try again.")
			continue
		}
		break
	}
	return result
}

func printPrompt(prompt string) {
	fmt.Printf("➡️  " + prompt + ": ")
}
