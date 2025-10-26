package constnats

import (
	"fmt"
	"strings"
)

type Medal string

const (
	MedalWood   Medal = "wood"
	MedalSteel  Medal = "steel"
	MedalBronze Medal = "bronze"
	MedalSilver Medal = "silver"
	MedalGold   Medal = "gold"
)

var medalByString = map[string]Medal{
	"wood":   MedalWood,
	"steel":  MedalSteel,
	"bronze": MedalBronze,
	"silver": MedalSilver,
	"gold":   MedalGold,
}

// LoadMedal loads the medal value from its string representation.
func LoadMedal(s string) (Medal, error) {
	medal, ok := medalByString[strings.ToLower(s)]
	if !ok {
		return "", fmt.Errorf("constnats.Medal: unknown medal %q", s)
	}
	return medal, nil
}
