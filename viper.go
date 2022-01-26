// Copyright (c) 2021 Jing-Ying Chen. Subject to the MIT License.

package jcli

import (
	"context"
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

const (
	ViperKey = "__viper__"
)

func WithViper(ctx context.Context, vip *viper.Viper) context.Context {
	return context.WithValue(ctx, ViperKey, vip)
}

func GetViper(ctx context.Context) *viper.Viper {
	if vip, ok := ctx.Value(ViperKey).(*viper.Viper); ok {
		return vip
	}
	return nil
}

// GetStringOrViper gets the value from the context using the key; if fails, tries
// to get the viper instance from the context then uses viperKey to get the value.
func GetStringOrViper(ctx context.Context, key, viperKey string) string {
	if val, ok := ctx.Value(key).(string); ok {
		return val
	}
	if vip := GetViper(ctx); vip != nil {
		return vip.GetString(viperKey)
	}
	return ""
}

func GetIntOrViper(ctx context.Context, key string, viperKey string) int {
	if val, ok := ctx.Value(key).(int); ok {
		return val
	}
	if vip := GetViper(ctx); vip != nil {
		return vip.GetInt(viperKey)
	}
	return 0
}

func GetBoolOrViper(ctx context.Context, key string, viperKey string) bool {
	if val, ok := ctx.Value(key).(bool); ok {
		return val
	}
	if vip := GetViper(ctx); vip != nil {
		return vip.GetBool(viperKey)
	}
	return false
}

func GetFloatOrViper(ctx context.Context, key string, viperKey string) float64 {
	if val, ok := ctx.Value(key).(float64); ok {
		return val
	}
	if vip := GetViper(ctx); vip != nil {
		return vip.GetFloat64(viperKey)
	}
	return 0
}

func StringFlagOrViper(ctx context.Context, key string, viperKey string) string {
	val := StringFlag(ctx, key, "")
	return StringOrViper(val, GetViper(ctx), viperKey)
}

func GetStringMap(ctx context.Context, key string) map[string]interface{} {
	if vip := GetViper(ctx); vip != nil {
		return vip.GetStringMap(key)
	}
	return nil
}

func GetStringMapString(ctx context.Context, key string) map[string]string {
	if vip := GetViper(ctx); vip != nil {
		return vip.GetStringMapString(key)
	}
	return nil
}

func StringOrViper(v string, vip *viper.Viper, key string) string {
	if v == "" && vip != nil {
		v = vip.GetString(key)
	}
	return v
}

func FieldOrViper(val interface{}, name string, vip *viper.Viper, keys ...string) interface{} {
	v := reflect.Indirect(reflect.ValueOf(val))
	if val, ok := fieldOrViper(v, name, vip, keys); ok {
		return val
	}
	return nil
}

func fieldOrViper(v reflect.Value, name string, vip *viper.Viper, keys []string) (interface{}, bool) {
	if v.Kind() == reflect.Struct {
		if f := v.FieldByName(name); f.IsValid() {
			return f.Interface(), true
		}
	}

	if vip != nil && len(keys) > 0 {
		return vip.Get(keys[0]), true
	}

	return nil, false
}

func BuildMap(val interface{}, vip *viper.Viper, exprs ...[]string) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(val))
	ret := make(map[string]interface{}, len(exprs))
	for _, expr := range exprs {
		if len(expr) > 0 {
			name := expr[0]
			keys := expr[1:]
			if val, ok := fieldOrViper(v, name, vip, keys); ok {
				ret[name] = val
			}
		}
	}
	return ret
}

func BuildStrMap(val interface{}, vip *viper.Viper, exprs ...[]string) map[string]string {
	v := reflect.Indirect(reflect.ValueOf(val))
	ret := map[string]string{}
	for _, expr := range exprs {
		if len(expr) > 0 {
			name := expr[0]
			keys := expr[1:]
			if val, ok := fieldOrViper(v, name, vip, keys); ok {
				if str, ok := val.(string); ok {
					ret[name] = str
				}
			}
		}
	}
	return ret
}

type ViperConfig struct {
	ConfigFile  string
	ConfigName  string
	ConfigType  string
	ConfigPaths []string
}

func NewViper(cfg ViperConfig) (*viper.Viper, error) {
	var vip *viper.Viper
	if cfg.ConfigFile != "" { // in cfg or from command flag
		vip = viper.New()
		vip.SetConfigFile(cfg.ConfigFile)
	} else if cfg.ConfigName != "" && len(cfg.ConfigPaths) > 0 {
		vip = viper.New()
		vip.SetConfigName(cfg.ConfigName)

		ct := cfg.ConfigType
		if ct == "" {
			ct = "yaml"
		}
		vip.SetConfigType(ct)

		for _, cp := range cfg.ConfigPaths {
			vip.AddConfigPath(cp)
		}
	} else {
		return nil, fmt.Errorf("Insufficicient config file information")
	}

	vip.AutomaticEnv()

	if err := vip.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	return vip, nil
}
