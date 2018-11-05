package signer

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"ipfs-monitor/command"

	ci "github.com/libp2p/go-libp2p-crypto"
)

//var priv *ci.PrivKey

var privatekey []byte

type Config struct {
	Identity Identity
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
	privatekey, err = base64.StdEncoding.DecodeString(result.Identity.PrivKey)
	if err != nil {
		panic("Can not decode base64 private key")
	}

}

func Sign(content string) ([]byte, error) {
	priv, err := ci.UnmarshalPrivateKey(privatekey)
	if err != nil {
		return nil, err
	}
	return priv.Sign([]byte(content))
}
