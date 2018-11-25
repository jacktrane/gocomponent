package format

import (
	"fmt"
	"strconv"
)

func ToString(v interface{}) string {
	switch v.(type) {
	case string:
		return v.(string)
	default:
		r := fmt.Sprint(v)
		if r == "<nil>" {
			return ""
		} else if r == "nil" {
			return ""
		} else {
			return r
		}
	}
}

func ToInt(v interface{}) int {
	switch v.(type) {
	case string:
		v, err := strconv.Atoi(v.(string))
		if err != nil {
			return 0
		} else {
			return v
		}

	case int32:
		return int(v.(int32))
	case int64:
		return int(v.(int64))
	case int:
		return v.(int)
	case int8:
		return int(v.(int8))
	case float64:
		return int(v.(float64))
	case float32:
		return int(v.(float32))
	default:
		return 0
	}
}

func ToMap(v interface{}) map[string]interface{} {
	switch v.(type) {
	case map[string]interface{}:
		return v.(map[string]interface{})
	default:
		return nil
	}
}

func ToArray(v interface{}) []interface{} {
	switch v.(type) {
	case []interface{}:
		return v.([]interface{})
	default:
		return nil
	}
}
