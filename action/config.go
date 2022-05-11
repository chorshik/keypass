package action

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/ebladrocher/keypass/fsutil"
	"github.com/ebladrocher/keypass/storepass"
	"github.com/ghodss/yaml"
	"github.com/urfave/cli/v2"
)

const (
	path = ".config/.keypass.yml"
)

// Config ...
func (s *Action) Config(c *cli.Context) error {
	if len(c.Args().Slice()) < 1 {
		return s.printConfigValues()
	}

	if len(c.Args().Slice()) == 1 {
		return s.printConfigValues(c.Args().Slice()[0])
	}

	if len(c.Args().Slice()) > 2 {
		return fmt.Errorf("Использование: keypass config key value")
	}

	return s.setConfigValue(c.Args().Slice()[0], c.Args().Slice()[1])
}

func (s *Action) printConfigValues(filter ...string) error {
	out := make([]string, 0, 10)
	o := reflect.ValueOf(s.Store).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("json")
		if jsonArg == "" || jsonArg == "-" {
			continue
		}
		if !contains(filter, jsonArg) {
			continue
		}
		f := o.Field(i)
		strVal := ""
		switch f.Kind() {
		case reflect.String:
			strVal = f.String()
		case reflect.Bool:
			strVal = fmt.Sprintf("%t", f.Bool())
		case reflect.Int:
			strVal = fmt.Sprintf("%d", f.Int())
		default:
			continue
		}
		out = append(out, fmt.Sprintf("%s: %s", jsonArg, strVal))
	}
	sort.Strings(out)
	for _, line := range out {
		fmt.Println(line)
	}
	return nil
}

func contains(haystack []string, needle string) bool {
	if len(haystack) < 1 {
		return true
	}
	for _, blade := range haystack {
		if blade == needle {
			return true
		}
	}
	return false
}

func (s *Action) setConfigValue(key, value string) error {
	if key != "path" {
		value = strings.ToLower(value)
	}
	o := reflect.ValueOf(s.Store).Elem()
	for i := 0; i < o.NumField(); i++ {
		jsonArg := o.Type().Field(i).Tag.Get("json")
		if jsonArg == "" || jsonArg == "-" {
			continue
		}
		if jsonArg != key {
			continue
		}
		f := o.Field(i)
		switch f.Kind() {
		case reflect.String:
			f.SetString(value)
		case reflect.Bool:
			if value == "true" {
				f.SetBool(true)
			} else if value == "false" {
				f.SetBool(false)
			} else {
				return fmt.Errorf("Не bool: %s", value)
			}
		case reflect.Int:
			iv, err := strconv.Atoi(value)
			if err != nil {
				return err
			}
			f.SetInt(int64(iv))
		default:
			continue
		}
	}
	return writeConfig(s.Store)
}

func writeConfig(s *storepass.RootStore) error {
	buf, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(configFile(), buf, 0600); err != nil {
		return err
	}
	return nil
}

func configFile() string {
	if cf := os.Getenv("KEYPASS_CONFIG"); cf != "" {
		return cf
	}

	return filepath.Join(os.Getenv("HOME"), path)
}

func hasConfig() bool {
	return fsutil.IsFile(configFile())
}
