/**
 *  Author: jack
 *  Date: 2018/5/13
 *  Description:
 */
package array

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBinarySort(t *testing.T) {
	result := make([]interface{}, 0)
	// 对包内容转流(json\xml\Excel\txt)
	testJson := filepath.Join("test.json")
	file, _ := os.Open(testJson)
	json.NewDecoder(file).Decode(&result)
	conditon := make(map[string]int)
	conditon["num"] = 0
	if res := BinarySort(result, conditon); res == nil {
		t.Error("fail")
	} else {
		resJson, _ := json.Marshal(res)
		resString := string(resJson)
		t.Log(resString)
	}
}
