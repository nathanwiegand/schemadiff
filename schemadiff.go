package schemadiff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// UnmappedJSONFields finds all the fields mentioned in the json string 'data'
// that will not be unmarshalled to fields in the Go struct value (x). The return value is the
// list of the fields or an error.
//
// Nested fields get scoped: B.C means field name C in a struct value that was named B.
//
// Note, because we can't inspect what magic is going on in a custom UnmarshalJSON, we assume
// assume that any type that correctly implements that interface must correctly parse the value,
// and we always give it a pass.
func UnmappedJSONFields(typ reflect.Type, data []byte) (map[string]string, error) {
	var val map[string]interface{}
	if err := json.Unmarshal(data, &val); err != nil {
		return nil, err
	}
	unknown := make(map[string]interface{})
	getUnmappedFields(typ, val, "", unknown)

	out := make(map[string]string)
	for k, v := range unknown {
		b, _ := json.MarshalIndent(v, "", " ")
		out[k] = string(b)
	}

	return out, nil
}

var jsonUnmarshalerInterface = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()

func getUnmappedFields(typ reflect.Type, val interface{}, prefix string, unknown map[string]interface{}) {
	if reflect.PtrTo(typ).Implements(jsonUnmarshalerInterface) {
		// Let's be generous and assume anyone implementing the Unmarshaler interface knows what they're doing.
		return
	}

	switch typ.Kind() {
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		// We can't deserialize into either a chan or a function.
		unknown[prefix] = val

	case reflect.Interface:
		// Interface is a catch all. Anything can be deserialized into it.
		return

	case reflect.Map:
		processMap(typ.Key(), typ.Elem(), val, prefix, unknown)

	case reflect.Ptr:
		// Strip the pointer and get the type.
		getUnmappedFields(typ.Elem(), val, prefix, unknown)

	case reflect.Array:
		fallthrough
	case reflect.Slice:
		processSlice(typ, val, prefix, unknown)

	case reflect.Struct:
		processStruct(typ, val, prefix, unknown)

	default:
		// This is a primitive type.
		if !reflect.TypeOf(val).ConvertibleTo(typ) {
			unknown[prefix] = val
		}
	}
}

func processMap(keyType reflect.Type, valType reflect.Type, val interface{}, prefix string, unknown map[string]interface{}) {
	stringType := reflect.TypeOf(string(""))
	if !keyType.ConvertibleTo(stringType) {
		// JSON maps have to be keyed by stringlike things.
		unknown[prefix] = val
		return
	}
	valMap, ok := val.(map[string]interface{})
	if !ok {
		unknown[prefix] = val
		return
	}
	for k, v := range valMap {
		getUnmappedFields(valType, v, prefix+fmt.Sprintf(`["%s"]`, k), unknown)
	}
}

func processSlice(typ reflect.Type, val interface{}, prefix string, unknown map[string]interface{}) {
	valSlice, ok := val.([]interface{})
	if !ok {
		unknown[prefix] = val
		return
	}
	sliceType := typ.Elem()
	for i, v := range valSlice {
		getUnmappedFields(sliceType, v, prefix+fmt.Sprintf("[%d]", i), unknown)
	}
}

func processStruct(typ reflect.Type, val interface{}, prefix string, unknown map[string]interface{}) {
	valMap, ok := val.(map[string]interface{})
	if !ok {
		unknown[prefix] = val
		return
	}
	knownFields := make(map[string]reflect.StructField)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		parts := strings.Split(field.Tag.Get("json"), ",")

		// Default the jsonFieldName to the Go field name.
		jsonFieldName := field.Name

		if len(parts) > 0 && len(parts[0]) > 0 {
			// If the struct has a json tag name, override the field name with that.
			jsonFieldName = parts[0]
		}

		if jsonFieldName == "-" {
			continue
		}
		knownFields[jsonFieldName] = field
	}

	for k, v := range valMap {
		key := prefix
		if key != "" {
			key += "."
		}
		key += k
		structField, ok := knownFields[k]
		typ := structField.Type
		if !ok {
			unknown[key] = v
			continue
		}
		getUnmappedFields(typ, v, key, unknown)
	}
}
