package locksmith

import "errors"

type RekeyStatus struct {
	Nonce             string   `json:"nonce"`
	Started           bool     `json:"started"`
	Progress          int      `json:"progress"`
	Required          int      `json:"required"`
	PGPFingerprints   []string `json:"pgp_fingerprints"`
	Keys              []string `json:"keys"`
	KeysBase64        []string `json:"keys_base64"`
	VerificationNonce string   `json:"verification_nonce"`
	Errors            []string `json:"errors"`
}

func (r RekeyStatus) HasError() bool {
	return len(r.Errors) > 0
}

func (r RekeyStatus) Error() error {
	if r.HasError() {
		return errors.New(r.Errors[0])
	}
	return nil
}

func (r RekeyStatus) ErrorMessage() string {
	if r.HasError() {
		return r.Errors[0]
	}
	return ""
}

func (r RekeyStatus) InProgress() bool {
	return r.Started
}

func (r RekeyStatus) RemainingKeys() int {
	return r.Required - r.Progress
}

type VerificationStatus struct {
	Nonce     string   `json:"nonce"`
	Started   bool     `json:"started"`
	Threshold int      `json:"t"`
	NewShares int      `json:"n"`
	Progress  int      `json:"progress"`
	Complete  bool     `json:"complete"`
	Errors    []string `json:"errors"`
}

func (v VerificationStatus) HasError() bool {
	return len(v.Errors) > 0
}

func (v VerificationStatus) Error() error {
	if v.HasError() {
		return errors.New(v.Errors[0])
	}
	return nil
}

func (v VerificationStatus) ErrorMessage() string {
	if v.HasError() {
		return v.Errors[0]
	}
	return ""
}

func (v VerificationStatus) InProgress() bool {
	if v.HasError() {
		return v.ErrorMessage() != "no rekey configuration found"
	}
	return v.Started
}

func (v VerificationStatus) Completed() bool {
	return v.Complete
}

func (v VerificationStatus) RemainingKeys() int {
	return v.Threshold - v.Progress
}

type StartRekeyRequest struct {
	SecretShares    int
	SecretThreshold int
	KeybaseUsers    []string
}

type startRekeyRequest struct {
	SecretShares        int      `json:"secret_shares"`
	SecretThreshold     int      `json:"secret_threshold"`
	PGPKeys             []string `json:"pgp_keys"`
	RequireVerification bool     `json:"require_verification"`
}

type submitKeyRequest struct {
	Key   string `json:"key"`
	Nonce string `json:"nonce"`
}
