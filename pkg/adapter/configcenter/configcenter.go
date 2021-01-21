package configcenter

import (
	"fmt"

	"github.com/symcn/meshach/pkg/adapter/component"
	"github.com/symcn/meshach/pkg/option"
	"k8s.io/klog"
)

type constructor func(regOpt option.Configuration) (component.ConfigurationCenter, error)

var (
	configInstance = make(map[string]constructor)
)

// Registry ...
func Registry(typ string, f constructor) {
	if _, ok := configInstance[typ]; ok {
		klog.Fatalf("repeat registry [config center instance]: %s", typ)
	}
	configInstance[typ] = f
}

// GetRegistry ...
func GetRegistry(opt option.Configuration) (component.ConfigurationCenter, error) {
	if f, ok := configInstance[opt.Type]; ok {
		return f(opt)
	}
	return nil, fmt.Errorf("config center {%s} was not implemented", opt.Type)
}
