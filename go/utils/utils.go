package utils

func Contains(arr []interface{}, value interface{}) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}

	return false
}
