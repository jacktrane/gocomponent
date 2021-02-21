package basic

func IfElseStr(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func IfElseInt(cond bool, a, b int) int {
	if cond {
		return a
	}
	return b
}

func IfElseInt64(cond bool, a, b int64) int64 {
	if cond {
		return a
	}
	return b
}

func IfElseInt32(cond bool, a, b int32) int32 {
	if cond {
		return a
	}
	return b
}
