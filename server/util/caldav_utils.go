package util

import (
	"github.com/emersion/go-ical"
)

func GetPropertyValue(prop *ical.Prop) string {
	if prop == nil {
		return ""
	}
	return prop.Value
}
