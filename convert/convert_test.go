package convert

import (
	"fmt"
	"testing"
)

func TestJSON2PB(t *testing.T) {
	eg := `{
			"image": {
				"height": 1,
				"width": 1
			},
			"res": [
				{
					"test1": [
					-0.577081
					],
					"test2": {
						"test3": "111",
						"test4": 0.1,
						"test5": 1
					}
				}
			],
			"status": 0
	   }`

	fmt.Println(JSON2PB([]byte(eg), "TestPb"))
}
