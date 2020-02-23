package generator

import (
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

	if err := config.Exists(); err != nil {
		return err
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
	gosConfig, err := config.Read()
	if err != nil {
		return err
	}
	gosConfig.Services = append(gosConfig.Services, name)
	return config.Write(gosConfig)
}
