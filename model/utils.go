package model

import (
	"github.com/jhump/protoreflect/dynamic"
)

func GetEnumValue(val *dynamic.Message, name string) string {
	return val.FindFieldDescriptorByName(name).GetEnumType().
		FindValueByNumber(val.GetFieldByName(name).(int32)).GetName()
}
