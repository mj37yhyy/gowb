package mcp

import (
	"reflect"
	"strings"
)

// GenerateSchema 从Go结构体生成JSON Schema
func GenerateSchema(inputType interface{}) map[string]interface{} {
	if inputType == nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	t := reflect.TypeOf(inputType)
	// 如果是指针，获取实际类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return map[string]interface{}{
			"type": "object",
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// 获取json tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// 解析json tag（可能包含omitempty等选项）
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		// 生成字段schema
		propSchema := typeToSchema(field.Type)

		// 添加描述（从desc tag）
		if desc := field.Tag.Get("desc"); desc != "" {
			propSchema["description"] = desc
		}

		// 添加到properties
		schema["properties"].(map[string]interface{})[jsonName] = propSchema

		// 检查是否必填
		bindingTag := field.Tag.Get("binding")
		if strings.Contains(bindingTag, "required") {
			required = append(required, jsonName)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

// typeToSchema 将Go类型转换为JSON Schema类型
func typeToSchema(t reflect.Type) map[string]interface{} {
	// 如果是指针，获取实际类型
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := make(map[string]interface{})

	switch t.Kind() {
	case reflect.String:
		schema["type"] = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		schema["items"] = typeToSchema(t.Elem())
	case reflect.Map:
		schema["type"] = "object"
		if t.Elem().Kind() != reflect.Interface {
			schema["additionalProperties"] = typeToSchema(t.Elem())
		}
	case reflect.Struct:
		// 嵌套结构体，递归生成schema
		return GenerateSchema(reflect.New(t).Elem().Interface())
	default:
		schema["type"] = "object"
	}

	return schema
}
