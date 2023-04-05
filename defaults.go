package modulr

import (
	"github.com/Meduzz/modulr/lib/event"
	"github.com/Meduzz/modulr/lib/proxy"
	"github.com/Meduzz/modulr/lib/registry"
)

var (
	ServiceRegistry = registry.NewServiceRegistry()
	HttpProxy       = proxy.NewProxy(ServiceRegistry)
	EventSupport    = event.NewEventSupport(ServiceRegistry)
)
