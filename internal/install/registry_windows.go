package install

import (
	"errors"

	"golang.org/x/sys/windows/registry"
)

func setRegistryValues(path string, values map[string]string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close() //nolint:errcheck
	for name, value := range values {
		err = key.SetStringValue(name, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteRegistryValue(path, name string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close() //nolint:errcheck
	err = key.DeleteValue(name)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}

	return err
}
