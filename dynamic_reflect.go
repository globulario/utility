// utility/dynamic_reflect.go
package Utility

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/gob"
	"errors"
	"log"
	"reflect"
	"strconv"
	"strings"
)

/*
   Dynamic typing, reflection, and function registry helpers
   ---------------------------------------------------------
   - Uses the concurrent-safe TypeManager (see typemanager.go).
   - Safe conversions via ToString/ToInt/ToBool/ToNumeric.
   - Gob name registration for round-trippable serialization.
*/

// -----------------------------
// Referenceable (optional trait)
// -----------------------------

// Referenceable can be implemented by types that expose a UUID.
type Referenceable interface {
	GetUUID() string
}

// -------------------------------------------
// Type registry, instantiation & serialization
// -------------------------------------------

// GetTypeOf returns the pointer type for a registered type name.
// Example: "mypkg.MyType" → *mypkg.MyType (reflect.Type)
func GetTypeOf(typeName string) reflect.Type {
	if t, ok := DefaultTypeManager().GetType(typeName); ok {
		return reflect.New(t).Type()
	}
	return nil
}

// GetInstanceOf creates a new *T instance of a registered type name.
// If the struct has an exported field "TYPENAME", it is set to typeName.
func GetInstanceOf(typeName string) interface{} {
	if t, ok := DefaultTypeManager().GetType(typeName); ok {
		instance := reflect.New(t).Interface()
		SetProperty(instance, "TYPENAME", typeName) // best-effort
		return instance
	}
	return nil
}

// RegisterType registers a type (by typed nil pointer) with the TypeManager
// and gob so values can be serialized/deserialized by name.
//
//   type Foo struct{}
//   RegisterType((*Foo)(nil))
func RegisterType(typedNil interface{}) {
	t := reflect.TypeOf(typedNil).Elem()
	idx := strings.LastIndex(t.PkgPath(), "/")
	typeName := t.Name()

	var fq string
	if idx > 0 {
		fq = t.PkgPath()[idx+1:] + "." + typeName
	} else {
		fq = t.PkgPath() + "." + typeName
	}

	if _, ok := DefaultTypeManager().GetType(fq); !ok {
		DefaultTypeManager().RegisterType(fq, t)
		gob.RegisterName(fq, typedNil)
	}
}

// ToBytes serializes any value via gob into a byte slice.
func ToBytes(val interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(val)
	return buf.Bytes(), err
}

// FromBytes deserializes data into a new instance of typeName if registered;
// otherwise into a map[string]interface{}.
func FromBytes(data []byte, typeName string) (interface{}, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	if t, ok := DefaultTypeManager().GetType(typeName); ok {
		v := reflect.New(t).Interface()
		err := dec.Decode(v)
		return v, err
	}

	v := make(map[string]interface{})
	err := dec.Decode(&v)
	return v, err
}

// ----------------------------------------
// Structure initialization from map values
// ----------------------------------------

// MakeInstance creates and initializes an instance of the given registered type
// with the provided map data. Optionally, setEntity is called for each created
// nested value (useful for building reference indexes).
func MakeInstance(typeName string, data map[string]interface{}, setEntity func(interface{})) reflect.Value {
	value := initializeStructureValue(typeName, data, setEntity)
	if setEntity != nil && value.IsValid() {
		setEntity(value.Interface())
	}
	return value
}

// InitializeStructure builds a single *T from a map containing "TYPENAME".
func InitializeStructure(data map[string]interface{}, setEntity func(interface{})) (reflect.Value, error) {
	var value reflect.Value
	tnAny, hasTN := data["TYPENAME"]
	if !hasTN {
		return value, errors.New("NotDynamicObject")
	}
	tn := ToString(tnAny)
	if _, ok := DefaultTypeManager().GetType(tn); ok {
		value = MakeInstance(tn, data, setEntity)
		if setEntity != nil && value.IsValid() {
			setEntity(value.Interface())
		}
		return value, nil
	}
	// If not registered, return the raw map.
	return reflect.ValueOf(data), nil
}

// InitializeStructures builds a slice of *T from []interface{} of maps.
// If typeName is empty and the first element has TYPENAME, that is used.
func InitializeStructures(data []interface{}, typeName string, setEntity func(interface{})) (reflect.Value, error) {
	var values reflect.Value

	if len(data) == 0 {
		if len(typeName) > 0 {
			if t, ok := DefaultTypeManager().GetType(typeName); ok {
				return reflect.MakeSlice(reflect.SliceOf(reflect.New(t).Type()), 0, 0), nil
			}
		}
		return reflect.ValueOf(make([]interface{}, 0)), nil
	}

	first := data[0]
	if m, ok := first.(map[string]interface{}); ok {
		tn := typeName
		if tn == "" {
			if v, ok := m["TYPENAME"]; ok {
				tn, _ = v.(string)
			}
		}
		if tn == "" {
			return reflect.ValueOf(data), nil
		}

		for i := 0; i < len(data); i++ {
			obj := MakeInstance(tn, data[i].(map[string]interface{}), setEntity)
			if i == 0 {
				if len(typeName) == 0 {
					values = reflect.MakeSlice(reflect.SliceOf(obj.Type()), 0, 0)
				} else if t, ok := DefaultTypeManager().GetType(typeName); ok {
					values = reflect.MakeSlice(reflect.SliceOf(reflect.New(t).Type()), 0, 0)
				} else {
					values = reflect.ValueOf(make([]interface{}, 0))
				}
			}
			values = reflect.Append(values, obj)
		}
		return values, nil
	}

	return values, errors.New("NotDynamicObject")
}

// InitializeArray converts []interface{} into a typed slice when the elements
// are uniform; otherwise returns []interface{}.
func InitializeArray(data []interface{}) (reflect.Value, error) {
	var values reflect.Value
	sameType := true

	if len(data) > 1 {
		for i := 1; i < len(data) && sameType; i++ {
			if data[i] != nil {
				sameType = reflect.TypeOf(data[i]).String() == reflect.TypeOf(data[i-1]).String()
			}
		}
	}

	for i := 0; i < len(data); i++ {
		if data[i] == nil {
			continue
		}
		if i == 0 {
			if sameType {
				values = reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(data[i])), 0, 0)
			} else {
				values = reflect.ValueOf(make([]interface{}, 0))
			}
		}
		values = reflect.Append(values, reflect.ValueOf(data[i]))
	}
	return values, nil
}

// initializeStructureValue creates a *T for the registered type and sets fields from data.
// If the type is not registered, it returns reflect.ValueOf(data).
func initializeStructureValue(typeName string, data map[string]interface{}, setEntity func(interface{})) reflect.Value {
	t, ok := DefaultTypeManager().GetType(typeName)
	if !ok {
		return reflect.ValueOf(data)
	}
	v := reflect.New(t)

	for name, raw := range data {
		if raw == nil {
			continue
		}
		if ft, exist := t.FieldByName(name); exist {
			initializeStructureFieldValue(v, name, ft.Type, raw, setEntity)
		}
	}
	return v
}

// InitializeStructureFieldArrayValue fills a slice with values converted from `values`.
func InitializeStructureFieldArrayValue(slice reflect.Value, fieldName string, fieldType reflect.Type, values reflect.Value, setEntity func(interface{})) {
	for i := 0; i < values.Len(); i++ {
		v_ := values.Index(i).Interface()
		if v_ == nil {
			continue
		}

		switch reflect.TypeOf(v_).String() {
		case "map[string]interface {}":
			m := v_.(map[string]interface{})
			if tn, hasTN := m["TYPENAME"]; hasTN {
				fv := initializeStructureValue(tn.(string), m, setEntity)
				if setEntity != nil && fv.IsValid() {
					setEntity(fv.Interface())
				}
				if strings.HasPrefix(fieldName, "M_") {
					if uuidAny, ok := m["UUID"]; ok {
						slice.Index(i).Set(reflect.ValueOf(ToString(uuidAny)))
					}
				} else {
					slice.Index(i).Set(fv)
				}
			} else {
				slice.Index(i).Set(reflect.ValueOf(m))
			}
		default:
			if reflect.TypeOf(v_).Kind() == reflect.Slice {
				slice_ := reflect.MakeSlice(fieldType, reflect.ValueOf(v_).Len(), reflect.ValueOf(v_).Len())
				InitializeStructureFieldArrayValue(slice_, fieldName, reflect.TypeOf(v_), reflect.ValueOf(v_), setEntity)
				if slice.Index(i).IsValid() {
					slice.Index(i).Set(slice_)
				}
			} else {
				fv := InitializeBaseTypeValue(slice.Type().Elem(), v_)
				if fv.IsValid() {
					if fv.Type() != slice.Index(i).Type() && fv.CanConvert(slice.Index(i).Type()) {
						fv = fv.Convert(slice.Index(i).Type())
					}
					slice.Index(i).Set(fv)
				}
			}
		}
	}
}

// initializeStructureFieldValue sets a struct field from an arbitrary value.
func initializeStructureFieldValue(v reflect.Value, fieldName string, fieldType reflect.Type, fieldValue interface{}, setEntity func(interface{})) {
	switch fieldType.Kind() {

	case reflect.Slice:
		// []byte special-case (often base64 in JSON payloads)
		rt := reflect.TypeOf(fieldValue)
		if rt != nil && (rt.String() == "[]uint8" || rt.String() == "[]byte") {
			fv := InitializeBaseTypeValue(rt, fieldValue)
			val := fv.Bytes()
			// try base64 decode if it looks like a base64-encoded string in a []byte shell
			if str := string(val); len(str) > 0 {
				if decoded, err := b64.StdEncoding.DecodeString(str); err == nil {
					val = decoded
				}
			}
			v.Elem().FieldByName(fieldName).Set(reflect.ValueOf(val))
			return
		}
		// Generic slice
		rvv := reflect.ValueOf(fieldValue)
		if rvv.IsValid() && rvv.Kind() == reflect.Slice {
			slice := reflect.MakeSlice(fieldType, rvv.Len(), rvv.Len())
			InitializeStructureFieldArrayValue(slice, fieldName, fieldType, rvv, setEntity)
			if slice.IsValid() {
				v.Elem().FieldByName(fieldName).Set(slice)
			}
		}

	case reflect.Struct:
		if m, ok := fieldValue.(map[string]interface{}); ok {
			if fv, _ := InitializeStructure(m, setEntity); fv.IsValid() {
				v.Elem().FieldByName(fieldName).Set(fv.Elem())
			}
		}

	case reflect.Ptr:
		if m, ok := fieldValue.(map[string]interface{}); ok {
			if fv, _ := InitializeStructure(m, setEntity); fv.IsValid() {
				v.Elem().FieldByName(fieldName).Set(fv)
			}
		}

	case reflect.Interface:
		initializeStructureFieldValue(v, fieldName, reflect.TypeOf(fieldValue), fieldValue, setEntity)

	case reflect.Map:
		if m, ok := fieldValue.(map[string]interface{}); ok {
			if fv, err := InitializeStructure(m, setEntity); err == nil && fv.IsValid() {
				v.Elem().FieldByName(fieldName).Set(fv)
			} else {
				v.Elem().FieldByName(fieldName).Set(reflect.ValueOf(fieldValue))
			}
		}

	case reflect.String:
		if m, ok := fieldValue.(map[string]interface{}); ok {
			if fv, err := InitializeStructure(m, setEntity); err == nil && fv.IsValid() {
				// write UUID field of nested value into string field
				u := fv.Elem().FieldByName("UUID")
				if u.IsValid() && u.Kind() == reflect.String {
					v.Elem().FieldByName(fieldName).Set(u)
					return
				}
			}
			v.Elem().FieldByName(fieldName).Set(reflect.ValueOf(fieldValue))
		} else {
			fv := InitializeBaseTypeValue(fieldType, fieldValue)
			if fv.IsValid() {
				v.Elem().FieldByName(fieldName).Set(fv.Convert(fieldType))
			}
		}

	default:
		fv := InitializeBaseTypeValue(fieldType, fieldValue)
		if fv.IsValid() {
			if fv.Type() != fieldType && fv.CanConvert(fieldType) {
				fv = fv.Convert(fieldType)
			}
			v.Elem().FieldByName(fieldName).Set(fv)
		}
	}
}

// InitializeBaseTypeValue converts an arbitrary value into a reflect.Value appropriate
// for the base type t. It prefers safe conversions via ToString/ToBool/ToInt/ToNumeric.
func InitializeBaseTypeValue(t reflect.Type, value interface{}) reflect.Value {
	if value == nil {
		return reflect.Value{}
	}
	if t.Kind() == reflect.Interface {
		return reflect.ValueOf(value)
	}

	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(ToString(value))
	case reflect.Bool:
		return reflect.ValueOf(ToBool(value))
	case reflect.Int:
		return reflect.ValueOf(ToInt(value))
	case reflect.Int8:
		return reflect.ValueOf(int8(ToInt(value)))
	case reflect.Int16:
		return reflect.ValueOf(int16(ToInt(value)))
	case reflect.Int32:
		return reflect.ValueOf(int32(ToInt(value)))
	case reflect.Int64:
		return reflect.ValueOf(int64(ToInt(value)))
	case reflect.Uint:
		return reflect.ValueOf(uint(ToInt(value)))
	case reflect.Uint8:
		return reflect.ValueOf(uint8(ToInt(value)))
	case reflect.Uint32:
		return reflect.ValueOf(uint32(ToInt(value)))
	case reflect.Uint64:
		return reflect.ValueOf(uint64(ToInt(value)))
	case reflect.Float32:
		return reflect.ValueOf(float32(ToNumeric(value)))
	case reflect.Float64:
		return reflect.ValueOf(float64(ToNumeric(value)))
	case reflect.Array:
		log.Println("InitializeBaseTypeValue: unexpected array kind")
		return reflect.Value{}
	default:
		log.Printf("InitializeBaseTypeValue: unexpected type %v\n", t)
		return reflect.Value{}
	}
}

// ---------------------------
// Dynamic function management
// ---------------------------

// RegisterFunction stores a function under a name for dynamic lookup/call.
func RegisterFunction(name string, fct interface{}) {
	DefaultTypeManager().RegisterFunc(name, fct)
}

// GetFunction retrieves a function by name (or nil if not found).
func GetFunction(name string) interface{} {
	if f, ok := DefaultTypeManager().GetFunc(name); ok {
		return f
	}
	return nil
}

// CallFunction calls a registered function by name with params.
// It validates arity for non-variadic functions and returns the raw reflect values.
func CallFunction(name string, params ...interface{}) (result []reflect.Value, err error) {
	fn := GetFunction(name)
	if fn == nil {
		return nil, errors.New("no function was register with name " + name)
	}

	fv := reflect.ValueOf(fn)
	ft := fv.Type()

	// Arity check for non-variadic functions
	if !ft.IsVariadic() && len(params) != ft.NumIn() {
		return nil, errors.New("Wrong number of parameter for " + name +
			" got " + strconv.Itoa(len(params)) +
			" but expect " + strconv.Itoa(ft.NumIn()))
	}

	// Build arguments with best-effort conversion
	in := make([]reflect.Value, len(params))
	for i, p := range params {
		if p == nil {
			// nil → zero of parameter type when not variadic; or a nil interface value for variadic tail
			if !ft.IsVariadic() || i < ft.NumIn()-1 {
				in[i] = reflect.Zero(ft.In(i))
			} else {
				var nilVal interface{}
				in[i] = reflect.ValueOf(&nilVal).Elem()
			}
			continue
		}

		v := reflect.ValueOf(p)
		// If types don't match but are convertible, convert.
		var target reflect.Type
		if !ft.IsVariadic() || i < ft.NumIn()-1 {
			target = ft.In(i)
		} else {
			// For variadic, the final param expects elements of the slice element type.
			target = ft.In(ft.NumIn() - 1).Elem()
		}
		if v.Type() != target && v.CanConvert(target) {
			v = v.Convert(target)
		}
		in[i] = v
	}

	result = fv.Call(in)
	return
}

// CallMethod calls a method by name on a given instance with params.
// Example:
//   result, err := CallMethod(myObj, "DoSomething", 42, "abc")
func CallMethod(instance interface{}, methodName string, params ...interface{}) ([]reflect.Value, error) {
	if instance == nil {
		return nil, errors.New("CallMethod: instance is nil")
	}

	v := reflect.ValueOf(instance)
	m := v.MethodByName(methodName)
	if !m.IsValid() {
		return nil, errors.New("CallMethod: method " + methodName + " not found")
	}

	mt := m.Type()
	// Check arity for non-variadic methods
	if !mt.IsVariadic() && len(params) != mt.NumIn() {
		return nil, errors.New("CallMethod: wrong number of parameters for " +
			methodName + " got " + strconv.Itoa(len(params)) +
			" but expect " + strconv.Itoa(mt.NumIn()))
	}

	in := make([]reflect.Value, len(params))
	for i, p := range params {
		if p == nil {
			if !mt.IsVariadic() || i < mt.NumIn()-1 {
				in[i] = reflect.Zero(mt.In(i))
			} else {
				var nilVal interface{}
				in[i] = reflect.ValueOf(&nilVal).Elem()
			}
			continue
		}
		vp := reflect.ValueOf(p)
		var target reflect.Type
		if !mt.IsVariadic() || i < mt.NumIn()-1 {
			target = mt.In(i)
		} else {
			target = mt.In(mt.NumIn() - 1).Elem()
		}
		if vp.Type() != target && vp.CanConvert(target) {
			vp = vp.Convert(target)
		}
		in[i] = vp
	}

	return m.Call(in), nil
}

// --------------------
// Small utility helpers
// --------------------

// GetProperty retrieves the value of a named exported field from a struct pointer.
// Returns (value, true) if the field exists and is accessible, or (nil, false) otherwise.
func GetProperty(ptr interface{}, field string) (interface{}, bool) {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}
	if !rv.IsValid() || rv.Kind() != reflect.Struct {
		return nil, false
	}

	f := rv.FieldByName(field)
	if !f.IsValid() {
		return nil, false
	}

	// Only exported fields are accessible.
	if !f.CanInterface() {
		return nil, false
	}

	return f.Interface(), true
}

// SetProperty sets an exported struct field if present (best-effort, no panic).
func SetProperty(ptr interface{}, field string, val interface{}) bool {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return false
	}
	rv = rv.Elem()
	if !rv.IsValid() {
		return false
	}
	f := rv.FieldByName(field)
	if !f.IsValid() || !f.CanSet() {
		return false
	}
	v := reflect.ValueOf(val)
	if v.IsValid() && v.Type() != f.Type() && v.CanConvert(f.Type()) {
		v = v.Convert(f.Type())
	}
	if v.IsValid() && v.Type() == f.Type() {
		f.Set(v)
		return true
	}
	return false
}

