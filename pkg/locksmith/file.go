package locksmith

import (
	"fmt"
	"io/ioutil"
	"time"
)

func WriteKeysToFile(vaultURL string, input WriteKeysToFileRequest) error {
	output := fmt.Sprintf("VAULT URL: %s\n\n", vaultURL)
	for i, key := range input.Keys {
		user := input.KeybaseUsers[i]
		fingerprint := input.PGPFingerprints[i]
		keyBase64 := input.KeysBase64[i]
		output += fmt.Sprintf("KEYBASE USER: %s\nFINGERPRINT: %s\nENCRYPTED_KEY: %s\nENCRYPTED_KEY_BASE64: %s\n\n", user, fingerprint, key, keyBase64)
	}
	fileName := fmt.Sprintf("recovery-keys-%s.txt", time.Now().Format("2006-01-02-15-04-05"))
	err := ioutil.WriteFile(fileName, []byte(output), 0644)
	if err != nil {
		return WrapError(err, "failed to write to file")
	}
	fmt.Printf("✍️  New recovery keys saved to: %s\n", fileName)
	return nil
}
