package utils

import (
	"fmt"
	"gomificator/internal/constnats"
	"os"
	"path/filepath"

	"golang.org/x/exp/constraints"
)

func ValidateRange[T constraints.Ordered](value, min, max T) error {
	if value > max || value < min {
		return fmt.Errorf("value %v not in range from %v to %v", value, min, max)
	}
	return nil
}

func GetAppDataLocation() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get user config dir: %w", err)
	}
	configPath := filepath.Join(configDir, constnats.Gomificator)

	return configPath, nil
}

func EnsureAppDataLocation() (string, error) {
	appDataLocation, err := GetAppDataLocation()
	if err != nil {
		return "", fmt.Errorf("get app data location: %w", err)
	}
	if err = os.MkdirAll(appDataLocation, os.ModePerm); err != nil {
		return "", fmt.Errorf("create app data location: %w", err)
	}
	return appDataLocation, nil
}

// FileExists checks if a file exists and is not a directory
func FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
