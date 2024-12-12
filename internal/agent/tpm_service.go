package agent

import (
	"crypto/rsa"
	"fmt"

	"github.com/google/go-tpm/tpm2"
)

type TPMService struct {
	tpm *tpm2.TPMContext
}

func NewTPMService(tpm *tpm2.TPMContext) *TPMService {
	return &TPMService{tpm: tpm}
}

func (s *TPMService) GenerateAttestationKey() (*rsa.PublicKey, error) {
	// Template for attestation key
	template := tpm2.Public{
		Type:       tpm2.AlgRSA,
		NameAlg:    tpm2.AlgSHA256,
		Attributes: tpm2.FlagSignerDefault | tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin,
		RSAParameters: &tpm2.RSAParams{
			Sign: &tpm2.SigScheme{
				Alg:  tpm2.AlgRSASSA,
				Hash: tpm2.AlgSHA256,
			},
			KeyBits: 2048,
		},
	}

	// Generate the attestation key
	handle, pubKey, err := s.tpm.CreatePrimary(tpm2.HandleOwner, tpm2.PCRSelection{}, "", "", template)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary key: %v", err)
	}
	defer s.tpm.FlushContext(handle)

	// Convert TPM public key to RSA public key
	rsaPubKey, err := pubKey.RSAParameters.Key()
	if err != nil {
		return nil, fmt.Errorf("failed to convert public key: %v", err)
	}

	return rsaPubKey, nil
}
