package helper

import (
	"reflect"
)

// InArray 判断目标是否存在于数组当中
func InArray(target interface{}, array interface{}) bool {
	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)
		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(target, s.Index(i).Interface()) {
				return true
			}
		}
	}
	return false
}
