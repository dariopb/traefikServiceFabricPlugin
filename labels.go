package traefikServiceFabricPlugin

import (
	"log"
	"strconv"
)

// GetStringValue get string value associated to a label.
func GetStringValue(labels map[string]string, labelName, defaultValue string) string {
	if value, ok := labels[labelName]; ok && len(value) > 0 {
		return value
	}
	return defaultValue
}

// GetBoolValue get bool value associated to a label.
func GetBoolValue(labels map[string]string, labelName string, defaultValue bool) bool {
	rawValue, ok := labels[labelName]
	if ok {
		v, err := strconv.ParseBool(rawValue)
		if err == nil {
			return v
		}
		log.Printf("Unable to parse %q: %q, falling back to %v. %v", labelName, rawValue, defaultValue, err)
	}
	return defaultValue
}

// GetIntValue get int value associated to a label.
func GetIntValue(labels map[string]string, labelName string, defaultValue int) int {
	if rawValue, ok := labels[labelName]; ok {
		value, err := strconv.Atoi(rawValue)
		if err == nil {
			return value
		}
		log.Printf("Unable to parse %q: %q, falling back to %v. %v", labelName, rawValue, defaultValue, err)
	}
	return defaultValue
}
