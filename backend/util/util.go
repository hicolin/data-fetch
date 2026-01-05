package util

import (
	"encoding/json"
	"fmt"
)

func Dump(v any) {
	fmt.Println("===== dump =====")
	indent, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(indent))
}
