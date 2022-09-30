package utils

import (
	"os"
	"strings"

	json "github.com/json-iterator/go"
	"github.com/spf13/viper"
)

type Conf struct {
	*viper.Viper
	Name string
	Data interface{}
}

func (conf *Conf) Sub(key string) *Conf {
	return &Conf{Viper: conf.Viper.Sub(key), Name: conf.Name}
}

func Load(name string, data interface{}) (*Conf, error) {
	if data == nil {
		data = &map[string]interface{}{}
	}

	conf := &Conf{Viper: viper.New(), Name: name, Data: data}
	if name == "" {
		return conf, nil
	}

	conf.Viper.SetConfigFile(name)
	if e := conf.Viper.ReadInConfig(); e != nil {
		if strings.HasSuffix(name, ".json") {
			// 兼容JSON数组
			if f, e := os.Open(name); e != nil {
				return conf, e
			} else if e := json.NewDecoder(f).Decode(&conf.Data); e != nil {
				return conf, e
			}
			return conf, nil
		}
		return conf, e
	}

	return conf, conf.Viper.Unmarshal(conf.Data)
}
