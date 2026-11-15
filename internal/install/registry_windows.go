package install

import (
	"errors"
	"log/slog"

	"golang.org/x/sys/windows/registry"
)

func setRegistryValues(path string, values map[string]string) error {
	slog.Info("writing registry values", "key", path)
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
	slog.Info("deleting registry value", "key", path, "name", name)
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
