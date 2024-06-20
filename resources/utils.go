package resources

import (
	"encoding/json"
	"os"
)

func Parse(path string, wrapper func(*ProTableProps)) ProTableProps {
	bytes, err := os.ReadFile(path)
	var props ProTableProps
	if err == nil {
		_ = json.Unmarshal(bytes, &props)
	}
	wrapper(&props)
	return props
}
