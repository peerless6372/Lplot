package env

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	SubConfDefault = ""
	SubConfMount   = "mount"
	SubConfApp     = "app"
)

func LoadConf(filename, subConf string, s interface{}) {
	var path string
	path = filepath.Join(GetConfDirPath(), subConf, filename)

	if yamlFile, err := ioutil.ReadFile(path); err != nil {
		panic(filename + " get error: %v " + err.Error())
	} else if err = yaml.Unmarshal(yamlFile, s); err != nil {
		panic(filename + " unmarshal error: %v" + err.Error())
	}
}
