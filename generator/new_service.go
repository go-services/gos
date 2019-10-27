package generator

import (
	"bytes"
	"encoding/json"
	"errors"
	"gos/config"
	"gos/fs"
	"gos/template"
	"strings"

	"github.com/ozgio/strutil"
	"github.com/spf13/afero"
)

// NewService generates a new service with the given name
func NewService(name string) error {
	appFs := fs.AppFs()

	b, err := afero.Exists(appFs, "kit.json")

	if err != nil {
		return err
	} else if !b {
		return errors.New("not in a kit project, you need to be in a project to run this command")
	}

	// we should remove the '_' because of this guide https://blog.golang.org/package-names
	folderName := strings.ReplaceAll(strutil.ToSnakeCase(name), "_", "")

	if err := fs.CreateFolder(appFs, folderName); err != nil {
		return err
	}

	data := map[string]string{
		"ProjectModule": folderName,
	}

	serviceFile, err := template.CompileGoFromPath("templates/service/service.go.gotmpl", data)
	if err != nil {
		return err
	}
	svcFs := afero.NewBasePathFs(appFs, folderName)
	err = fs.WriteFile(svcFs, "service.go", serviceFile)
	if err != nil {
		return err
	}
	configData, err := fs.ReadFile(appFs, "kit.json")
	if err != nil {
		return errors.New("could not read kit.json")
	}
	var kitConfig config.KitConfig
	err = json.NewDecoder(bytes.NewBufferString(configData)).Decode(&kitConfig)
	if err != nil {
		return err
	}
	kitConfig.Services = append(kitConfig.Services, name)
	newData, err := json.MarshalIndent(kitConfig, "", "\t")
	if err != nil {
		return err
	}
	return fs.WriteFile(appFs, "kit.json", string(newData))
}
