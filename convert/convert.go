package convert

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/jacktrane/gocomponent/stringlib"
)

// JSON2PB 将json转成pb
func JSON2PB(i []byte, key string) (string, error) {
	res := ""
	iMap := make(map[string]interface{})
	err := json.Unmarshal(i, &iMap)
	if err != nil {
		return res, err
	}

	_, _, d := ast(iMap, key)
	return d[0], err
}

func ast(i interface{}, key string) (isArr bool, kType string, constructs []string) {
	kind := reflect.TypeOf(i).Kind().String()
	if _, ok := baseTypeMap[kind]; ok {
		kType = kind
		return
	}

	switch r := i.(type) {
	case map[string]interface{}:
		index := 0
		kType = stringlib.FirstUpper(key)
		syntaxs := make([]string, 0)
		message := make([]string, 0)
		for k, v := range r {
			index++
			subSyntaxs := make([]string, 0)
			subIsArr, subType, subConstruct := ast(v, k)
			if subIsArr {
				subSyntaxs = append(subSyntaxs, "repeated")
			}
			subSyntaxs = append(subSyntaxs, subType, k, "=", strconv.Itoa(index))
			field := strings.Join(subSyntaxs, " ") + ";"
			syntaxs = append(syntaxs, field)
			message = append(message, subConstruct...)
		}

		elements := strings.Join(append(message, syntaxs...), "\n")
		baseConstruct := strings.ReplaceAll("message "+kType+" {\n"+elements+"\n}", "\n", "\n    ")
		placeholder := strings.LastIndex(baseConstruct, "\n    ")
		structured := append([]byte(baseConstruct[:placeholder+1]), baseConstruct[placeholder+5:]...)
		constructs = append(constructs, string(structured))

	case []interface{}:
		for _, v := range r {
			_, ktype1, a := ast(v, key)
			kType = ktype1
			constructs = append(constructs, a...)
		}
		isArr = true
	}
	return isArr, kType, constructs
}

var (
	baseTypeMap = map[string]int{
		"int32":  INT32,
		"int64":  INT64,
		"float":  FLOAT,
		"string": STRING,
		// "map":    MAP,
		// "array":  ARRAY,
		// "slice":  SLICE,
	}
)

const (
	INT32 = iota
	INT64
	FLOAT
	STRING
	ARRAY
	MAP
	SLICE
)
