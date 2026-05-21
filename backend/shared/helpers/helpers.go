package helpers

import "slices"

func PingStatus(err error) string {
	if err != nil {
		return "not connected"
	}
	return "connected"
}

func Contains[T comparable](arr []T, target T) bool {
	return slices.Contains(arr, target)
}

func ContainsAll[T comparable](main []T, sub []T) bool {
	set := make(map[T]struct{}, len(main))
	for _, v := range main {
		set[v] = struct{}{}
	}

	for _, v := range sub {
		if _, ok := set[v]; !ok {
			return false
		}
	}

	return true
}
