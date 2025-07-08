package utils

import (
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// BoolToString converts a boolean value to a string ("true"/"false").
func BoolToString(value bool) (strValue string) {
	if value {
		strValue = "true"
		return strValue
	}
	strValue = "false"
	return strValue
}

func ParseBoolOrDefault(logger *zap.Logger, input string, defaultVal bool) bool {
	if input == "" {
		return defaultVal
	}

	parsed, err := strconv.ParseBool(input)
	if err != nil {
		if logger != nil {
			logger.Warn("Invalid boolean parameter; defaulting to defaultVal",
				zap.String("paramValue", input),
				zap.Error(err),
			)
		}
		return defaultVal
	}
	return parsed
}

func MapTransactionStatus(input string) string {
	input = strings.ToLower(input)
	switch {
	case strings.Contains(input, "pending"),
		strings.Contains(input, "in_progress"),
		strings.Contains(input, "processing"):
		return "pending"
	case strings.Contains(input, "completed"),
		strings.Contains(input, "confirmed"):
		return "completed"
	case input == "paid":
		return "paid"
	case strings.Contains(input, "failed"),
		strings.Contains(input, "error"):
		return "failed"
	default:
		return "new" // Fallback to default
	}
}
