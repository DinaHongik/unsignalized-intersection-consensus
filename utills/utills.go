package utills

// Function name: RemoveValue
// Removes the given value from the slice and returns a new slice.
func RemoveValue(slice []int32, valueToRemove int32) []int32 {
	if !Contains(slice, valueToRemove) {
		return slice
	} else {
		result := []int32{}
		for _, value := range slice {
			if value != valueToRemove {
				result = append(result, value)
			}
		}
		return result
	}
}

// Function name: Contains
// Returns true if the int32 value exists in the slice.
func Contains(slice []int32, item int32) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Function name: Difference
// Returns elements in a that are not present in b.
func Difference(a, b []int) []int {
	m := make(map[int]bool)
	for _, item := range b {
		m[item] = true
	}
	diff := []int{}
	for _, item := range a {
		if !m[item] {
			diff = append(diff, item)
		}
	}
	return diff
}

// Function name: ContainsInt
// Returns true if the int value exists in the slice.
func ContainsInt(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
