// utility/env_windows.go
//go:build windows

package Utility

import (
	"errors"
	// Uncomment to enable registry access:
	"golang.org/x/sys/windows/registry"
)

// Windows-specific environment variable helpers
func SetWindowsEnvironmentVariable(key string, value string) error {

		k, err := registry.OpenKey(registry.LOCAL_MACHINE,
			`SYSTEM\ControlSet001\Control\Session Manager\Environment`, registry.ALL_ACCESS)
		if err != nil {
			return err
		}
		defer k.Close()

		err = k.SetStringValue(key, value)
		if err != nil {
			return err
		}
		return nil

	return errors.New("SetWindowsEnvironmentVariable requires registry access (unimplemented stub)")
}

func GetWindowsEnvironmentVariable(key string) (string, error) {

		k, err := registry.OpenKey(registry.LOCAL_MACHINE,
			`SYSTEM\ControlSet001\Control\Session Manager\Environment`, registry.ALL_ACCESS)
		if err != nil {
			return "", err
		}
		defer k.Close()

		value, _, err := k.GetStringValue(key)
		if err != nil {
			return "", err
		}
		return value, nil

	return "", errors.New("GetWindowsEnvironmentVariable requires registry access (unimplemented stub)")
}

func UnsetWindowsEnvironmentVariable(key string) error {

		k, err := registry.OpenKey(registry.LOCAL_MACHINE,
			`SYSTEM\ControlSet001\Control\Session Manager\Environment`, registry.ALL_ACCESS)
		if err != nil {
			return err
		}
		defer k.Close()

		err = k.DeleteValue(key)
		if err != nil {
			return err
		}
		return nil

}

