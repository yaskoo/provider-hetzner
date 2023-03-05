package util

import "reflect"

func LabelsUpToDate(desired, actual map[string]string) bool {
	return (desired == nil && len(actual) == 0) || reflect.DeepEqual(desired, actual)
}
