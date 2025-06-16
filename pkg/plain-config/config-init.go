package plain_config

import (
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	globalConfig atomic.Value
)

type (
	//Option config init option
	Option interface {
		configOptionIs()
	}

	//WithSourceFile option
	WithSourceFile struct {
		Option
		FileName string
	}

	//WithSource option
	WithSource struct {
		Option
		Source io.Reader
		Type   string
	}

	defKeyValue struct {
		Option
		key string
		val any
	}

	//WithAcceptEnvironment option
	WithAcceptEnvironment struct {
		Option
		EnvPrefix string
	}

	// WithEnvKeyReplacer option
	WithEnvKeyReplacer struct {
		Option
		Old []string
		New []string
	}

	flag2bind struct {
		Option
		k string
		f *pflag.Flag
	}
)

// BindFlag option
func BindFlag[tKey ~string, tFlag *pflag.Flag | *flag.Flag](key tKey, f tFlag) Option {
	ret := flag2bind{
		k: string(key),
	}
	switch a := any(f).(type) {
	case *pflag.Flag:
		ret.f = a
	case *flag.Flag:
		ret.f = pflag.PFlagFromGoFlag(a)
	}
	return ret
}

// WithDefValue option
func WithDefValue[T ValueAccessorKeyType](key T, v any) Option {
	return defKeyValue{
		key: key.String(),
		val: v,
	}
}

func configStore() *viper.Viper {
	ret, _ := globalConfig.Load().(*viper.Viper)
	return ret
}

// InitGlobalConfig init global config
func InitGlobalConfig(opts ...Option) error {
	const api = "InitGlobalConfig"

	cfgHolder := viper.NewWithOptions(viper.KeyDelimiter("/"),
		viper.EnvKeyReplacer(strings.NewReplacer("/", "_")))

	for _, opt := range opts {
		switch t := opt.(type) {
		case defKeyValue:
			cfgHolder.SetDefault(t.key, t.val)
		case WithSourceFile:
			if len(t.FileName) == 0 {
				break
			}
			ext := filepath.Ext(t.FileName)
			if len(ext) == 0 {
				return errors.Wrapf(errors.New("no file type provided"),
					"%s: open file '%s'", api, t.FileName)
			}
			f, e := os.Open(t.FileName)
			if e != nil {
				return errors.Wrapf(e, "%s: open file '%s'", api, t.FileName)
			}
			cfgHolder.SetConfigType(ext[1:])
			e = cfgHolder.MergeConfig(f)
			_ = f.Close()
			if e != nil {
				return errors.Wrapf(e, "%s: consume config file '%s'", api, t.FileName)
			}
		case WithSource:
			cfgHolder.SetConfigType(t.Type)
			if e := cfgHolder.MergeConfig(t.Source); e != nil {
				return errors.Wrapf(e, "%s: consume source type '%s'", api, t.Type)
			}
		case WithAcceptEnvironment:
			cfgHolder.AutomaticEnv()
			cfgHolder.SetEnvPrefix(t.EnvPrefix)
		case WithEnvKeyReplacer:
			if len(t.Old) != len(t.New) {
				return errors.Errorf(
					"%s: in opt WithEnvKeyReplacer found len(Old) != len(New)", api,
				)
			}
			var pairs []string
			for i := range t.Old {
				pairs = append(pairs, t.Old[i], t.New[i])
			}
			cfgHolder.SetEnvKeyReplacer(strings.NewReplacer(pairs...))
		case flag2bind:
			if e := cfgHolder.BindPFlag(t.k, t.f); e != nil {
				return errors.Wrapf(e, "%s: bing flag '%s'", api, t.f.Name)
			}
		default:
			return errors.Wrapf(errors.New("unexpected option"),
				"%s: consume source type '%T'", api, opt)
		}
	}
	globalConfig.Store(cfgHolder)
	return nil
}

func init() {
	globalConfig.Store(viper.New())
}
