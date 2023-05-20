package utils

import (
	"bytes"
	"strconv"
	"strings"

	json "github.com/json-iterator/go"
)

func StrToUint(strNumber string, value interface{}) error {
	number, err := strconv.ParseUint(strNumber, 10, 64)
	if err != nil {
		return err
	}
	switch d := value.(type) {
	case *uint64:
		*d = number
	case *uint:
		*d = uint(number)
	case *uint16:
		*d = uint16(number)
	case *uint32:
		*d = uint32(number)
	case *uint8:
		*d = uint8(number)
	}
	return nil
}

func IntStringContain(a []string, x string) bool {
	for _, v := range a {
		if v == x {
			return true
		}
	}
	return false
}

func MultiJoinString(str ...string) string {
	var b bytes.Buffer
	b.Grow(256)
	length := len(str)
	for i := 0; i < length; i++ {
		b.WriteString(str[i])
	}
	return b.String()
}

func IsJsonString(s string) bool {
	var js string
	return json.Unmarshal([]byte(s), &js) == nil
}

func IsJsonMap(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func IsJsonSlice(s string) bool {
	var js []interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// 比较已经排序好的字符串切片
func CompareSortedStrings(lhs, rhs []string) bool {
	if len(rhs) != len(lhs) {
		return false
	}
	for index, value := range lhs {
		if value != rhs[index] {
			return false
		}
	}
	return true
}

// return adds, dels
func DiffStrings(oldStrings, newStrings []string) ([]string, []string) {
	if len(oldStrings) <= 0 {
		return newStrings, nil
	}

	if len(newStrings) <= 0 {
		return nil, oldStrings
	}

	mapOlds := make(map[string]int)
	for index, value := range oldStrings {
		mapOlds[value] = index
	}

	mapNews := make(map[string]int)
	for index, value := range newStrings {
		mapNews[value] = index
	}

	slAdds := make([]string, 0, 8)
	slDels := make([]string, 0, 8)

	for k, _ := range mapNews {
		_, exists := mapOlds[k]
		if !exists {
			slAdds = append(slAdds, k)
		}
	}

	for k, _ := range mapOlds {
		_, exists := mapNews[k]
		if !exists {
			slDels = append(slDels, k)
		}
	}

	return slAdds, slDels
}

func UnmarshalJson(jsonVal []byte, objVal interface{}) error {
	decoder := json.NewDecoder(strings.NewReader(BytesToString(jsonVal)))
	decoder.UseNumber()
	err := decoder.Decode(objVal)
	return err
}

// 将传入字符转换成数字的hash算法（兼容odp）
func HashInt(s string) int64 {
	var num int64
	if len(s) == 0 {
		return num
	}

	r := []uint8(s)
	num = int64(r[0])

	r = r[1:]
	for _, v := range r {
		num = ((num << 5) - num + int64(v)) & 0xFFFFFFFF
	}

	return num
}
