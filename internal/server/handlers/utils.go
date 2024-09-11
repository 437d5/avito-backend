package handlers

import (
	"fmt"
	"net/http"
	"strconv"
)

type ErrorResponse struct {
	Reason string `json:"reason"`
}

// ParseQueryParam returns int value from get params if ok
// or it return defaultValue if it is not set
// or error if provided param is not a number
func ParseQueryParam(r *http.Request, param string, defaultValue int) (int, error) {
    valueStr := r.URL.Query().Get(param)
    
    if valueStr == "" {
        return defaultValue, nil
    }
    
    value, err := strconv.Atoi(valueStr)
    if err != nil {
        return 0, fmt.Errorf("invalid value for %s: %s", param, valueStr)
    }
    
    return value, nil
}