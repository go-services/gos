package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"gos/fs"

	"github.com/spf13/afero"
)

const FileName = "gos.json"

type GosConfig struct {
	Module   string   `json:"module"`
	Services []string `json:"services"`
}

func Read() (*GosConfig, error) {
	rootFs := fs.AppFs()
	configData, err := fs.ReadFile(rootFs, FileName)
	if err != nil {
		return nil, errors.New("not in a GOS project, you need to be in a GOS project to run this command")
	}
	var gosConfig GosConfig
	err = json.NewDecoder(bytes.NewBufferString(configData)).Decode(&gosConfig)
	if err != nil {
		return nil, errors.New("GOS config malformed: " + err.Error())
	}
	return &gosConfig, nil
}

func Exists() error {
	b, err := afero.Exists(fs.AppFs(), FileName)
	if err != nil {
		return err
	} else if !b {
		return errors.New("not in a GOS project, you need to be in a project to run this command")
	}
	return nil
}
func Write(gosConfig *GosConfig) error {
	newData, err := json.MarshalIndent(gosConfig, "", "\t")
	if err != nil {
		return err
	}
	return fs.WriteFile(fs.AppFs(), FileName, string(newData))
}
