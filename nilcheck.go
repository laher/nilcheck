package nilcheck

import (
	"fmt"
	"log"
	"reflect"
)

type NilChecker struct {
	visitedAddresses []uintptr
	MaxDepth         int
	Log              Logger
}

func NewNilChecker() *NilChecker {
	return &NilChecker{
		Log: SimpleLogger{},
	}
}

type Context struct {
	visitedAddresses []uintptr
	Current          reflect.Value
	Path             string
	Original         reflect.Value
}

func (nc *NilChecker) Check(obj interface{}) error {
	valueOfObj := reflect.ValueOf(obj)
	if !valueOfObj.IsValid() {
		msg := fmt.Sprintf("Supplied object is not valid")
		return fmt.Errorf(msg)
	}
	context := &Context{
		Original:         valueOfObj,
		Current:          valueOfObj,
		Path:             "",
		visitedAddresses: []uintptr{},
	}
	err := nc.check(context)
	nc.Log.Printf("Visited pointer addresses: %+v", nc.visitedAddresses)
	return err
}

func (nc *NilChecker) check(context *Context) error {

	typeOfIn := context.Current.Type()
	context.Path = fmt.Sprintf("%s/%+v", context.Path, typeOfIn)
	nc.Log.Printf("Path: %s", context.Path)
	in := context.Current
	if CanBeNil(in.Kind()) && in.IsNil() {
		msg := fmt.Sprintf("%v is NIL\n", typeOfIn)
		return fmt.Errorf(msg)
	}
	switch in.Kind() {
	case reflect.Ptr:
		el := in.Elem()
		if !el.CanAddr() {
			msg := fmt.Sprintf("Cannot address pointer (probably nil)\n")
			return fmt.Errorf(msg)
		}
		ptr := in.Elem().Addr().Pointer()
		nc.visitedAddresses = append(nc.visitedAddresses, ptr)
		nc.Log.Printf("Visited pointer addresses: %+v", nc.visitedAddresses)
		inValue := in.Elem()
		if !inValue.IsValid() {
			return fmt.Errorf("Elem() not valid")
		}
		context.Current = inValue
		return nc.check(context)

	case reflect.Interface:
		inValue := in.Elem()
		context.Current = inValue
		return nc.check(context)
	case reflect.Struct:
		for i := 0; i < in.NumField(); i++ {
			context.Path = fmt.Sprintf("%s/%s", context.Path, typeOfIn.Field(i).Name)
			context.Current = in.Field(i)
			err := nc.check(context)
			if err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < in.Len(); i++ {
			context.Path = fmt.Sprintf("%s/[%d]", context.Path, i)
			context.Current = in.Index(i)
			err := nc.check(context)
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range in.MapKeys() {
			if CanBeNil(key.Kind()) && key.IsNil() {
				return fmt.Errorf("Map key is nil")
			}
			context.Path = fmt.Sprintf("%s/[%v]", context.Path, key)
			context.Current = in.MapIndex(key)
			err := nc.check(context)
			if err != nil {
				return err
			}
		}

	case reflect.Chan:
		//nil checked above switch
	case reflect.Bool, reflect.String:
	case reflect.Float32, reflect.Float64:
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	case reflect.Complex64, reflect.Complex128:
		//ignore
	default:
		return fmt.Errorf("Unhandled type '%s'!", in.Kind())
	}

	//OK
	return nil
}

func CanBeNil(k reflect.Kind) bool {
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return true
	}
	return false
}

type Logger interface {
	Printf(msg string, args ...interface{})
}

type SimpleLogger struct {
}

func (s SimpleLogger) Printf(msg string, args ...interface{}) {
	log.Printf(msg, args...)
}
