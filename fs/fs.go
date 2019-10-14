package fs

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

var fs afero.Fs

// AppFs returns the current location file system.
//  if we are in testing mode it returns a memory
func AppFs() afero.Fs {
	if viper.Get("testFs") != nil {
		return viper.Get("testFs").(afero.Fs)
	}
	if fs == nil {
		fs = afero.NewOsFs()
	}
	return fs
}

func DeleteFolder(fs afero.Fs, path string) error {
	return fs.RemoveAll(path)
}

func CreateFolder(fs afero.Fs, path string) error {
	b, err := afero.Exists(fs, path)
	if err != nil {
		return err
	} else if b {
		return fmt.Errorf("folder with the name `%s` already exists", path)
	}
	return fs.Mkdir(path, 0755)
}

func WriteFile(fs afero.Fs, path, data string) error {
	dir := filepath.Dir(path)
	b, err := afero.Exists(fs, dir)
	if err != nil {
		return err
	} else if !b {
		err := fs.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return afero.WriteFile(fs, path, []byte(data), 0644)
}

func ReadFile(fs afero.Fs, path string) (string, error) {
	b, err := afero.ReadFile(fs, path)
	return string(b), err
}
