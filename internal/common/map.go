package common

func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for key, value := range m {
			result[key] = value
		}
	}
	return result
}
