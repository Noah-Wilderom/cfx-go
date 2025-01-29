package cfx

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"syscall/js"
)

const functionSuffix = "()"

var mapReturns = map[reflect.Kind]func(js.Value) []reflect.Value{
	reflect.Float64: returnFloat,
	reflect.Int:     returnInt,
	reflect.Bool:    returnBool,
	reflect.String:  returnString,
	reflect.Struct:  returnStruct, // This handles any struct dynamically
}

type function struct {
	fn        func() func(...interface{}) js.Value
	mapReturn func(js.Value) []reflect.Value
}

func isFunction(t string) bool {
	return strings.HasSuffix(t, functionSuffix)
}

func bindFunction(tag string, t reflect.Type, parent func() js.Value) reflect.Value {
	return reflect.MakeFunc(t, newFunction(tag, t, parent).call)
}

func newFunction(tag string, t reflect.Type, parent func() js.Value) function {
	name := strings.TrimSuffix(tag, functionSuffix)
	fn := func() func(...interface{}) js.Value { return parent().Get(name).Invoke }

	//FIXME check func return type

	var mapReturn func(js.Value) []reflect.Value

	if t.NumOut() == 0 {
		mapReturn = returnVoid
	} else {
		var ok bool
		mapReturn, ok = mapReturns[t.Out(0).Kind()]
		if !ok {
			fmt.Println("fmt println werkt")
			Print("type: ", t.Out(0).Kind().String())
			panic(fmt.Sprintf("FIXME | %s", t.Out(0).Kind().String())) //FIXME
		}
	}

	return function{fn, mapReturn}
}

func (f function) call(argValues []reflect.Value) []reflect.Value {
	args := make([]interface{}, len(argValues))
	for i, argValue := range argValues {
		//TODO if argument is of kind struct, map it to a new JS object...

		args[i] = argValue.Interface()
	}

	return f.mapReturn(f.fn()(args...))
}

func returnVoid(_ js.Value) []reflect.Value {
	return []reflect.Value{}
}

func returnFloat(v js.Value) []reflect.Value {
	return []reflect.Value{reflect.ValueOf(v.Float())}
}

func returnInt(v js.Value) []reflect.Value {
	return []reflect.Value{reflect.ValueOf(v.Int())}
}

func returnBool(v js.Value) []reflect.Value {
	return []reflect.Value{reflect.ValueOf(v.Bool())}
}

func returnString(v js.Value) []reflect.Value {
	return []reflect.Value{reflect.ValueOf(v.String())}
}

func returnStruct(v js.Value) []reflect.Value {
	// Dynamically create an empty struct based on the JS value
	structType := reflect.TypeOf(v) // Type of the Go struct
	structValue := reflect.New(structType).Elem()

	// Iterate over the struct's fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)
		jsFieldValue := v.Get(field.Name) // Get the JS object field by the name of the Go field

		// Handle the field dynamically based on its type
		if jsFieldValue.IsUndefined() {
			continue // Skip undefined fields
		}

		// Set the field's value based on its Go type
		switch fieldValue.Kind() {
		case reflect.Int:
			if jsFieldValue.Type() == js.TypeNumber {
				fieldValue.SetInt(int64(jsFieldValue.Int()))
			}
		case reflect.Float64:
			if jsFieldValue.Type() == js.TypeNumber {
				fieldValue.SetFloat(jsFieldValue.Float())
			}
		case reflect.String:
			if jsFieldValue.Type() == js.TypeString {
				fieldValue.SetString(jsFieldValue.String())
			}
		case reflect.Bool:
			if jsFieldValue.Type() == js.TypeBoolean {
				fieldValue.SetBool(jsFieldValue.Bool())
			}
		case reflect.Ptr:
			// Handle pointers (e.g., pointer to another struct)
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			// Recursively handle nested structs or pointers
			fieldValue.Elem().Set(reflect.ValueOf(returnStruct(jsFieldValue)[0].Interface()))
		case reflect.Array, reflect.Slice:
			// Handle slices and arrays (assuming JS array matches Go slice)
			sliceValue := reflect.MakeSlice(fieldValue.Type(), jsFieldValue.Length(), jsFieldValue.Length())
			for j := 0; j < sliceValue.Len(); j++ {
				// Get the individual element from the JS array
				jsElement := jsFieldValue.Index(j)

				// Convert the JavaScript value to a Go value based on the expected type
				switch sliceValue.Index(j).Kind() {
				case reflect.Int:
					sliceValue.Index(j).SetInt(int64(jsElement.Int()))
				case reflect.Float64:
					sliceValue.Index(j).SetFloat(jsElement.Float())
				case reflect.String:
					sliceValue.Index(j).SetString(jsElement.String())
				case reflect.Bool:
					sliceValue.Index(j).SetBool(jsElement.Bool())
				default:
					log.Printf("Warning: Unsupported slice element type %s\n", sliceValue.Index(j).Kind())
				}
			}
			fieldValue.Set(sliceValue)
		default:
			log.Printf("Warning: Unsupported field type %s for field %s\n", fieldValue.Kind(), field.Name)
		}
	}

	return []reflect.Value{structValue}
}
