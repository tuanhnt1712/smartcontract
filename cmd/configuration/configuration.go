package configuration

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/smartcontract/signer"
)

type Configuration struct {
	Signer     *signer.Signer
	Contract   string `json:"contract"`
	SignerPath string `json:"signer_path"`
	AbiPath    string `json:"abi_path"`
}

func NewConfiguration() *Configuration {
	_, fileLocation, _, _ := runtime.Caller(1)
	file := filepath.Join(fileLocation, "../config.json")
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	configuration := &Configuration{}
	err = json.Unmarshal(raw, configuration)
	if err != nil {
		panic(err)
	}
	path := filepath.Join(fileLocation, configuration.SignerPath)
	signer := signer.NewSigner(path, fileLocation)

	configuration.Signer = signer

	return configuration
}
