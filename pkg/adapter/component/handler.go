package component

import "github.com/symcn/meshach/pkg/adapter/types"

// EventHandler described all component comes from adapter needs to be handle by various event handler.
type EventHandler interface {
	// AddService you should handle the event described that a service has been created
	AddService(event *types.ServiceEvent)
	// DeleteService you should handle the event describe that a service has been removed
	DeleteService(event *types.ServiceEvent)
	// AddInstance you should handle the event described that an instance has been registered
	AddInstance(event *types.ServiceEvent)
	// AddInstance you should handle the event described that an instance has been registered
	ReplaceInstances(event *types.ServiceEvent)
	// AddInstance you should handle the event describe that an instance has been unregistered
	DeleteInstance(event *types.ServiceEvent)

	// ReplaceAccessorInstances replacing the instances' host in an accessor CR should be done when
	// a subscribed mapping between any accessor and any service has been changed.
	ReplaceAccessorInstances(event *types.ServiceEvent, getScopedServices func(s string) map[string]struct{})

	// AddConfigEntry you should handle the event depicted that a dynamic configuration has been added
	AddConfigEntry(event *types.ConfigEvent)
	// ChangeConfigEntry you should handle the event depicted that a dynamic configuration has been changed
	ChangeConfigEntry(event *types.ConfigEvent)
	// DeleteConfigEntry you should handle the event depicted that a dynamic configuration has been deleted
	DeleteConfigEntry(event *types.ConfigEvent)
}
