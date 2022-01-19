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

// GetValueOrViper uses the key to get value from context, and if nil, tries
// to get viper instance from context and use the same key to get value from it.
// Note that zero values are considered non-present. Use with care.
func GetValueOrViper(ctx context.Context, key string) interface{} {
	val := ctx.Value(key)
	if val == nil {
		if vip := GetViper(ctx); vip != nil {
			val = vip.Get(key)
		}
	}
	return val
}
func GetBoolOrViper(ctx context.Context, key string) bool {
	b, ok := GetValueOrViper(ctx, key).(bool)
	return ok && b
}
func GetStringOrViper(ctx context.Context, key string) string {
	if s, ok := GetValueOrViper(ctx, key).(string); ok {
		return s
	}
	return ""
}

// ValueOrViper returns input v if it is non-nil, or the first non-nil value from
// the viper instance using the provided keys.
func ValueOrViper(v interface{}, vip *viper.Viper, keys ...string) interface{} {
	if v != nil {
		return v
	}
	return lookupViper(vip, keys)
}

func lookupViper(vip *viper.Viper, keys []string) interface{} {
	if vip != nil {
		for _, key := range keys {
			if val := vip.Get(key); val != nil {
				return val
			}
		}
	}
	return nil
}

func StringOrViper(v string, vip *viper.Viper, keys ...string) string {
	val := ValueOrViper(v, vip, keys...)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func FieldOrViper(val interface{}, name string, vip *viper.Viper, keys ...string) interface{} {
	v := reflect.Indirect(reflect.ValueOf(val))
	return fieldOrViper(v, name, vip, keys)
}

func fieldOrViper(v reflect.Value, name string, vip *viper.Viper, keys []string) interface{} {
	if v.Kind() == reflect.Struct {
		if f := v.FieldByName(name); f.IsValid() {
			if val := f.Interface(); val != nil {
				return val
			}
		}
	}

	return lookupViper(vip, keys)
}

// BuildMap creates a string-keyed map from a struct. Each expr slice consists
// of the field name (capitalized), followed by a list of keys to search in viper if the
// named field contains zero value. Furthermore, the key used for the returning map
// is the first (i.e. canonical) one of the keys slice, or the field name if keys is empty.
func BuildMap(val interface{}, vip *viper.Viper, exprs ...[]string) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(val))
	ret := make(map[string]interface{}, len(exprs))
	for _, expr := range exprs {
		if len(expr) > 0 {
			name := expr[0]
			keys := expr[1:]
			val := fieldOrViper(v, name, vip, keys)
			if val != nil {
				if len(keys) > 0 {
					ret[keys[0]] = val
				} else {
					ret[name] = val
				}
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
			val := fieldOrViper(v, name, vip, keys)
			if str, ok := val.(string); ok {
				if len(keys) > 0 {
					ret[keys[0]] = str
				} else {
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
