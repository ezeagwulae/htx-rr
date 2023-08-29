package crossing

func isCrossingOpen(status string) bool {
	openStatuses := []string{"clear", "active", "operational"}
	if stringInSlice(status, openStatuses) {
		return true
	}
	return false
}

// hasCrossingStatusChanged reports where a crossing signal has changed
// since the previous check
func hasCrossingStatusChanged(id string, isOpen bool) bool {
	// get crossing by id and compare last status with current status
	return false
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
