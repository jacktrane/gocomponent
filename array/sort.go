/**
 *  Author: jack
 *  Date: 2018/5/13
 *  Description: 数组
 */
package array

import "strconv"

// Description: 二元数组的排序(二分插入法)
// Param: arr:二元数组 condition:排序条件
// Return:  排序后的数组
func BinarySort(arr []interface{}, condition map[string]int) []interface{} {
	for k, v := range condition { // 条件(其中v:1为正序,0为倒序)

		for i := 1; i < len(arr); i++ {
			switch arr[i].(type) {
			case map[string]interface{}:

			default:
				return nil
			}

			current := arr[i].(map[string]interface{})
			left := 0
			right := i - 1
			for left <= right {
				mid := (left + right) / 2
				midMap := arr[mid].(map[string]interface{})
				if v == 1 { // 正序
					// 从小到大
					if Int(midMap[k]) > Int(current[k]) {
						right = mid - 1
					} else {
						left = mid + 1
					}
				} else { // 倒序
					// 从大到小
					if Int(midMap[k]) < Int(current[k]) {
						right = mid - 1
					} else {
						left = mid + 1
					}
				}

			}

			for j := i - 1; j >= left; j-- {
				arr[j+1] = arr[j]
			}

			arr[left] = current
		}
	}

	return arr
}

// 转int型
func Int(o interface{}) int {
	switch o.(type) {
	case string:
		v, err := strconv.Atoi(o.(string))
		if err != nil {
			return 0
		} else {
			return v
		}

	case int32:
		return int(o.(int32))
	case int64:
		return int(o.(int64))
	case int:
		return o.(int)
	case int8:
		return int(o.(int8))
	case float64:
		return int(o.(float64))
	case float32:
		return int(o.(float32))
	default:
		return 0
	}
}
