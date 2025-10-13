package main

import (
	"testing"
)

func BenchmarkLoadConfigFile(b *testing.B) {
	configDir, _, err := ScanConfigDir()
	if err != nil {
		b.Fatalf("ScanConfigDir failed: %v", err)
	}

	jsonPath := configDir + "/development.json"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := LoadConfigFile(jsonPath)
		if err != nil {
			b.Fatalf("LoadConfigFile failed: %v", err)
		}
	}
}

func BenchmarkIsConfigFile(b *testing.B) {
	testNames := []string{
		"config.json", "settings.yaml", "data.yml", "config.toml",
		"readme.txt", "script.sh", ".gitignore",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := testNames[i%len(testNames)]
		IsConfigFile(name)
	}
}

func BenchmarkNewCommandExecutor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewCommandExecutor()
	}
}
