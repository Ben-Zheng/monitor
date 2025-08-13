package es

import (
	"fmt"
	"strings"
	"testing"
)

func TestEs(t *testing.T) {
	//appConfig := config.GetEsConfig()
	//ec := NewESService(appConfig)
	//result, err := ec.BatchCountFieldOccurrences()
	//if err != nil {
	//	t.Error(err)
	//	return
	//}
	str := "哈喽#, WORLD!"
	lowerStr := strings.ToLower(str)
	fmt.Println(lowerStr)
}
