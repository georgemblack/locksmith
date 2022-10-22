package locksmith

import (
	"encoding/json"
	"net/http"
)

type RekeyStatus struct {
	Nonce string `json:"nonce"`
}

// GetRekeyStatus fetches the current rekey status from Vault
func GetRekeyStatus(baseURL string) (RekeyStatus, error) {
	client := &http.Client{}
	url := baseURL + "v1/sys/rekey-recovery-key/init"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return RekeyStatus{}, wrapError(err, "failed to create rekey status request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return RekeyStatus{}, wrapError(err, "failed to execute rekey status request")
	}
	var rekeyStatus RekeyStatus
	err = json.NewDecoder(resp.Body).Decode(&rekeyStatus)
	if err != nil {
		return RekeyStatus{}, wrapError(err, "failed to decode rekey status response")
	}
	return rekeyStatus, nil
}
