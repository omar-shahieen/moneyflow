package billing

import (
	"errors"
	"strings"
)

var ErrPlanNotFound = errors.New("plan not found")

func GetPlanAmount(plan string, prices map[string]float64) (float64, error) {
	if prices == nil {
		return 0, ErrPlanNotFound
	}
	amount, ok := prices[strings.ToLower(plan)]
	if !ok {
		return 0, ErrPlanNotFound
	}
	return amount, nil
}
