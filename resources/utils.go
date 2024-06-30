package resources

import (
	"encoding/json"
	"fmt"
	"os"
)

func RemoveDuplicates(arr []any) []any {
	uniqueMap := make(map[string]bool)
	var uniqueArr []any

	for _, value := range arr {
		// 将元素转换为字符串作为 map 的键
		key := fmt.Sprintf("%v", value)
		if _, exists := uniqueMap[key]; !exists {
			uniqueMap[key] = true
			uniqueArr = append(uniqueArr, value)
		}
	}

	return uniqueArr
}

func Parse(path string, wrapper func(*ProTableProps)) ProTableProps {
	bytes, err := os.ReadFile(path)
	var props ProTableProps
	if err == nil {
		_ = json.Unmarshal(bytes, &props)
	}
	wrapper(&props)
	return props
}
