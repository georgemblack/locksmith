package locksmith

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/hashicorp/vault/helper/pgpkeys"
)

func GetRekeyStatus(baseURL string) (RekeyStatus, error) {
	client := &http.Client{}
	url := baseURL + "/v1/sys/rekey-recovery-key/init"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to create rekey status request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to execute rekey status request")
	}

	// Parse response
	var result RekeyStatus
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to decode rekey status response")
	}

	// Check response
	if resp.StatusCode != 200 {
		err = errors.New(result.ErrorMessage())
		return result, WrapError(err, "failed to get rekey status")
	}
	return result, nil
}

func StartRekey(baseURL string, input StartRekeyRequest) (RekeyStatus, error) {
	// Fetch public keys from Keybase
	var users []string
	for _, user := range input.KeybaseUsers {
		users = append(users, "keybase:"+user)
	}
	keyMap, err := pgpkeys.FetchKeybasePubkeys(users)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to fetch public keys from Keybase")
	}
	var keys []string
	for _, user := range users {
		keys = append(keys, keyMap[user])
	}

	// Build request body
	startRekeyRequest := startRekeyRequest{
		SecretShares:        input.SecretShares,
		SecretThreshold:     input.SecretThreshold,
		PGPKeys:             keys,
		RequireVerification: true,
	}

	// Execute request
	client := &http.Client{}
	url := baseURL + "/v1/sys/rekey-recovery-key/init"
	body, err := json.Marshal(startRekeyRequest)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to marshal start rekey request")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to create rekey start request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to execute rekey start request")
	}

	// Parse response
	var result RekeyStatus
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to decode rekey status response")
	}

	// Check response
	if resp.StatusCode != 200 {
		err := errors.New("failed to start rekey, unexpected status code: " + resp.Status)
		return result, WrapError(err, result.ErrorMessage())
	}

	return result, nil
}

func SubmitKey(baseURL string, key string) (RekeyStatus, error) {
	// Fetch nonce from API
	status, err := GetRekeyStatus(baseURL)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to get rekey status")
	}
	if status.HasError() {
		return RekeyStatus{}, errors.New("failed to get rekey status: " + status.ErrorMessage())
	}

	// Build request body
	submitKeyRequest := submitKeyRequest{
		Key:   key,
		Nonce: status.Nonce,
	}

	// Execute request
	client := &http.Client{}
	url := baseURL + "/v1/sys/rekey-recovery-key/update"
	body, err := json.Marshal(submitKeyRequest)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to marshal submit key request")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to create submit key request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to execute submit key request")
	}

	// Parse response
	var result RekeyStatus
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return RekeyStatus{}, WrapError(err, "failed to decode rekey status response")
	}

	// Check response
	if resp.StatusCode != 200 {
		err = errors.New(result.ErrorMessage())
		return result, WrapError(err, "failed to submit key")
	}

	return result, nil
}

func GetVerificationStatus(baseURL string) (VerificationStatus, error) {
	// Build & execute request
	client := &http.Client{}
	url := baseURL + "/v1/sys/rekey-recovery-key/verify"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return VerificationStatus{}, WrapError(err, "failed to create verification status request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return VerificationStatus{}, WrapError(err, "failed to execute verification status request")
	}

	// Parse response
	var result VerificationStatus
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return VerificationStatus{}, WrapError(err, "failed to decode verification status response")
	}
	return result, nil
}

func SubmitVerification(baseURL string, key string) (VerificationStatus, error) {
	// Fetch nonce from API
	status, err := GetRekeyStatus(baseURL)
	if err != nil {
		return VerificationStatus{}, WrapError(err, "failed to get rekey status")
	}
	if status.HasError() {
		return VerificationStatus{}, errors.New("failed to get rekey status: " + status.ErrorMessage())
	}

	// Build request body
	submitKeyRequest := submitKeyRequest{
		Key:   key,
		Nonce: status.VerificationNonce,
	}

	// Execute request
	client := &http.Client{}
	url := baseURL + "/v1/sys/rekey-recovery-key/verify"
	body, err := json.Marshal(submitKeyRequest)
	if err != nil {
		return VerificationStatus{}, WrapError(err, "failed to marshal submit key request")
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return VerificationStatus{}, WrapError(err, "failed to create submit key request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return VerificationStatus{}, WrapError(err, "failed to execute submit key request")
	}
	if resp.StatusCode != 200 {
		return VerificationStatus{}, errors.New("failed to submit key, unexpected status code: " + resp.Status)
	}

	// Parse response
	var result VerificationStatus
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return VerificationStatus{}, WrapError(err, "failed to decode verification status response")
	}

	// Check response
	if resp.StatusCode != 200 {
		err = errors.New(result.ErrorMessage())
		return result, WrapError(err, "failed to submit key")
	}

	return result, nil
}
