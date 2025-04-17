package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func modifyRegistry(keyPath, name string, modifyFunc func(key registry.Key) error) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	if err := modifyFunc(key); err != nil {
		return err
	}

	notifyEnvironmentChange()
	return nil
}

func SetEnvironmentVariable(keyPath, name, value string) error {
	return modifyRegistry(keyPath, name, func(key registry.Key) error {
		return key.SetStringValue(name, value)
	})
}

func DeleteEnvironmentVariable(keyPath, name string) error {
	return modifyRegistry(keyPath, name, func(key registry.Key) error {
		return key.DeleteValue(name)
	})
}

func AddToPath(keyPath, newPath string) error {
	return modifyRegistry(keyPath, "Path", func(key registry.Key) error {
		path, _, err := key.GetStringValue("Path")
		if err != nil {
			return fmt.Errorf("failed to read Path value: %w", err)
		}

		paths := strings.Split(path, ";")
		for _, p := range paths {
			if strings.EqualFold(strings.TrimSpace(p), newPath) {
				return nil // 已经存在，无需添加
			}
		}

		path = path + ";" + newPath
		return key.SetStringValue("Path", path)
	})
}

func notifyEnvironmentChange() {
	cmd := exec.Command("powershell", "-Command", "[System.Environment]::SetEnvironmentVariable('Environment', $null, [System.EnvironmentVariableTarget]::User)")
	_ = cmd.Run() // 忽略错误，因为刷新失败通常不会影响程序运行
}
