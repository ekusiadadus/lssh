package conf

import (
	"fmt"
	"os"
	"os/user"
	r "reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Log     LogConfig
	Include map[string]IncludeConfig
	Common  ServerConfig
	Server  map[string]ServerConfig
}

type LogConfig struct {
	Enable bool   `toml:"enable"`
	Dir    string `toml:"dirpath"`
}

type IncludeConfig struct {
	Path string `toml:"path"`
}

type ServerConfig struct {
	Addr        string `toml:"addr"`
	Port        string `toml:"port"`
	User        string `toml:"user"`
	Pass        string `toml:"pass"`
	Key         string `toml:"key"`
	PreCmd      string `toml:"pre_cmd"`
	PostCmd     string `toml:"post_cmd"`
	ProxyServer string `toml:"proxy_server"`
	Note        string `toml:"note"`
}

type ServerConfigMaps map[string]ServerConfig

func ReadConf(confPath string) (checkConf Config) {
	if !isExist(confPath) {
		fmt.Printf("Config file(%s) Not Found.\nPlease create file.\n\n", confPath)
		fmt.Printf("sample: %s\n", "https://raw.githubusercontent.com/blacknon/lssh/master/example/config.tml")
		os.Exit(1)
	}

	// Read config file
	_, err := toml.DecodeFile(confPath, &checkConf)
	if err != nil {
		panic(err)
	}

	// reduce common setting (in .lssh.conf servers)
	for key, value := range checkConf.Server {
		setValue := serverConfigReduct(checkConf.Common, value)
		checkConf.Server[key] = setValue
	}

	// Read include files
	if checkConf.Include != nil {
		for _, v := range checkConf.Include {
			var includeConf Config

			// user path
			usr, _ := user.Current()
			path := strings.Replace(v.Path, "~", usr.HomeDir, 1)

			// Read include config file
			_, err := toml.DecodeFile(path, &includeConf)
			if err != nil {
				panic(err)
			}

			// reduce common setting
			setCommon := serverConfigReduct(checkConf.Common, includeConf.Common)

			// add include file serverconf
			for key, value := range includeConf.Server {
				// reduce common setting
				setValue := serverConfigReduct(setCommon, value)
				checkConf.Server[key] = setValue
			}
		}
	}

	// Check Config Parameter
	checkAlertFlag := checkServerConf(checkConf)
	if !checkAlertFlag {
		os.Exit(1)
	}

	return
}

func serverConfigReduct(perConfig, childConfig ServerConfig) ServerConfig {
	result := ServerConfig{}

	// struct to map
	perConfigMap, _ := structToMap(&perConfig)
	childConfigMap, _ := structToMap(&childConfig)

	resultMap := mapReduce(perConfigMap, childConfigMap)
	_ = mapToStruct(resultMap, &result)

	return result
}

func mapReduce(map1, map2 map[string]interface{}) map[string]interface{} {
	for ia, va := range map1 {
		if va != "" && map2[ia] == "" {
			map2[ia] = va
		}
	}
	return map2
}

func structToMap(val interface{}) (mapVal map[string]interface{}, ok bool) {
	structVal := r.Indirect(r.ValueOf(val))
	typ := structVal.Type()

	mapVal = make(map[string]interface{})

	for i := 0; i < typ.NumField(); i++ {
		field := structVal.Field(i)

		if field.CanSet() {
			mapVal[typ.Field(i).Name] = field.Interface()
		}
	}

	return
}

func mapToStruct(mapVal map[string]interface{}, val interface{}) (ok bool) {
	structVal := r.Indirect(r.ValueOf(val))
	for name, elem := range mapVal {
		structVal.FieldByName(name).Set(r.ValueOf(elem))
	}

	return
}

func GetNameList(listConf Config) (nameList []string) {
	for k := range listConf.Server {
		nameList = append(nameList, k)
	}
	return
}
