package utils

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
)

//将 map 转换为指定的结构体
func MapToStruct(mapInstance map[interface{}]interface{}, pojo interface{}) {
	if err := mapstructure.Decode(mapInstance, &pojo); err != nil {
		fmt.Println(err)
	}
	fmt.Printf("map2struct后得到的 struct 内容为:%v", pojo)
}
