package signer

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"ipfs-monitor/command"
	"math/big"
)

var priv *ecdsa.PrivateKey

type Config struct {
	Identity
}

type Identity struct {
	PeerId  string
	PrivKey string
}

func Initialize() {
	var result Config
	configPath, err := command.GetRepoPath()
	if err != nil {
		panic("Can not get path of config file")
	}
	configPath = configPath + "/config"
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic("Can not read config file")
	}
	err = json.Unmarshal(content, &result)
	if err != nil {
		panic("Can not parse config file to json")
	}
	bytes, err := base64.StdEncoding.DecodeString(result.Identity.PrivKey)
	if err != nil {
		panic("Can not decode base64 private key")
	}
	k := new(big.Int)
	k.SetBytes(bytes)
	priv = new(ecdsa.PrivateKey)
	curve := elliptic.P256()
	priv.PublicKey.Curve = curve
	priv.D = k
	// priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(k.Bytes())
	// fmt.Printf("X: %d, Y: %d\n", priv.PublicKey.X, priv.PublicKey.Y)
}

func Sign(content string) (string, error) {
	h := md5.New()
	io.WriteString(h, content)
	signhash := h.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, priv, signhash)
	if err != nil {
		return "", err
	}
	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)
	return hex.EncodeToString(signature), nil
}
