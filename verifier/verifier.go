package verifier

import (
	"encoding/base64"
	"encoding/hex"

	ci "github.com/libp2p/go-libp2p-crypto"
)

func Verify(pubkey string, content string, signature string) bool {
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubkey)
	if err != nil {
		return false
	}
	publicKey, err := ci.UnmarshalPublicKey([]byte(pubKeyBytes))
	if err != nil {
		return false
	}
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	b, err := publicKey.Verify([]byte(content), sig)
	if err != nil {
		return false
	}
	return b
}
