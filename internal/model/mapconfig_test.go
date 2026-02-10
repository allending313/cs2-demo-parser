package models

import (
	"reflect"
	"testing"
)

func TestLoadMapConfigs(t *testing.T) {
	configsDir := "../../maps/configs"

	configs, err := LoadMapConfigs(configsDir)
	if err != nil {
		t.Fatalf("LoadMapConfigs failed: %v", err)
	}

	// Check for a specific map
	dust2, ok := configs["de_dust2"]
	if !ok {
		t.Fatal("Config for de_dust2 not found")
	}

	expected := &MapConfig{
		Name:        "de_dust2",
		DisplayName: "Dust II",
		PosX:        -2476,
		PosY:        3239,
		Scale:       4.4,
		RadarFile:   "de_dust2.png",
		RadarWidth:  1024, // default
		RadarHeight: 1024, // default
	}

	if !reflect.DeepEqual(expected, dust2) {
		t.Errorf("Config for 'de_dust2' does not match.\nExpected: %+v\nActual:   %+v", expected, dust2)
	}
}
