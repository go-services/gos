package generator

import (
	"github.com/spf13/afero"
	"gos/fs"
	"gos/template"
	"strings"

	"github.com/ozgio/strutil"
)

func NewProject(name string) error {
	appFs := fs.AppFs()

	// we should remove the '_' because of this guide https://blog.golang.org/package-names
	moduleName := strings.ReplaceAll(strutil.ToSnakeCase(name), "_", "")

	if err := fs.CreateFolder(appFs, moduleName); err != nil {
		return err
	}

	gomod, err := template.CompileFromPath("templates/project/go.mod.gotmpl", map[string]string{
		"ProjectModule": moduleName,
	})
	if err != nil {
		return err
	}
	projectFs := afero.NewBasePathFs(appFs, moduleName)

	gitignore, err := template.FromPath("project/gitignore")
	if err != nil {
		return err
	}
	kitJson, err := template.CompileFromPath("templates/project/kit.json.gotmpl", map[string]string{
		"ProjectModule": moduleName,
	})
	if err != nil {
		return err
	}
	if err := fs.WriteFile(projectFs, ".gitignore", gitignore); err != nil {
		return err
	}
	if err := fs.WriteFile(projectFs, "go.mod", gomod); err != nil {
		return err
	}
	if err := fs.WriteFile(projectFs, "kit.json", kitJson); err != nil {
		return err
	}

	return nil
}
