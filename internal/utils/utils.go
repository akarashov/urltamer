// Package utils provides utility functions for the application.
package utils

// NA returns "N/A" if the input string is empty, otherwise it returns the input string.
func NA(s string) string {
	switch s {
	case "":
		return "N/A"
	default:
		return s
	}
}
