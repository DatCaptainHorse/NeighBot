package adapters

import (
	"errors"
	"fmt"
	"reflect"
)

type adapterEntry struct {
	adapterType reflect.Type
	configType  reflect.Type
}

var adapterRegistry = make(map[string]adapterEntry)

func RegisterAdapter(name string, adapterPrototype ChatAdapter, configPrototype interface{}) error {
	adapterVal := reflect.ValueOf(adapterPrototype)
	adapterType := adapterVal.Type()
	if adapterType.Kind() != reflect.Ptr || adapterType.Elem().Kind() != reflect.Struct {
		return errors.New("adapterPrototype must be a pointer to a struct")
	}

	configVal := reflect.ValueOf(configPrototype)
	configType := configVal.Type()
	if configType.Kind() != reflect.Struct {
		return errors.New("configPrototype must be a struct")
	}

	adapterRegistry[name] = adapterEntry{
		adapterType: adapterType,
		configType:  configType,
	}

	return nil
}

func CreateAdapter(name string) (ChatAdapter, reflect.Type, error) {
	entry, ok := adapterRegistry[name]
	if !ok {
		return nil, nil, fmt.Errorf("adapter not found: %s", name)
	}

	adapterVal := reflect.New(entry.adapterType.Elem())
	adapter, ok := adapterVal.Interface().(ChatAdapter)
	if !ok {
		return nil, nil, fmt.Errorf("failed to cast to ChatAdapter: %s", name)
	}

	return adapter, entry.configType, nil
}

func RegisteredAdapters() []string {
	keys := make([]string, 0, len(adapterRegistry))
	for k := range adapterRegistry {
		keys = append(keys, k)
	}
	return keys
}

func ConfigTypeForAdapter(name string) reflect.Type {
	entry, ok := adapterRegistry[name]
	if !ok {
		return nil
	}
	return entry.configType
}

func NewAdapterDefaultConfig(adapterName string) interface{} {
	configType := ConfigTypeForAdapter(adapterName)
	if configType == nil {
		return nil
	}

	defaultConfig := reflect.New(configType).Interface()
	return defaultConfig
}
