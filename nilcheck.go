package nilcheck

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
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
	Path             []string
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
		Path:             []string{},
		visitedAddresses: []uintptr{},
	}
	err := nc.check(context)
	nc.Log.Printf("Visited pointer addresses: %+v", nc.visitedAddresses)
	return err
}

func (nc *NilChecker) check(context *Context) error {

	typeOfIn := context.Current.Type()
	context.Path = append(context.Path, typeOfIn.String())
	nc.Log.Printf("Path: %#v", context.Path)
	in := context.Current
	if CanBeNil(in.Kind()) && in.IsNil() {
		msg := fmt.Sprintf("%v is NIL\n", typeOfIn)
		return fmt.Errorf(msg)
	}
	var err error
	switch in.Kind() {
	case reflect.Ptr:
		if !in.Elem().CanAddr() {
			msg := fmt.Sprintf("Cannot address pointer (probably nil)\n")
			return fmt.Errorf(msg)
		}
		ptr := in.Elem().Addr().Pointer()
		alreadyVisited := false
		for _, p := range nc.visitedAddresses {
			if ptr == p {
				//already visited. Return
				alreadyVisited = true
				break
			}
		}
		if !alreadyVisited {
			nc.visitedAddresses = append(nc.visitedAddresses, ptr)
			nc.Log.Printf("Visited pointer addresses: %+v", nc.visitedAddresses)
			if !in.Elem().IsValid() {
				return fmt.Errorf("Elem() not valid")
			}
			context.Current = in.Elem()
			err = nc.check(context)
		}
	case reflect.Interface:
		context.Current = in.Elem()
		err = nc.check(context)
	case reflect.Struct:
		for i := 0; i < in.NumField(); i++ {
			context.Path = append(context.Path, typeOfIn.Field(i).Name)
			context.Current = in.Field(i)
			err = nc.check(context)
			if err != nil {
				return err
			}
			context.Path = context.Path[:len(context.Path)-1]
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < in.Len(); i++ {
			context.Path = append(context.Path, strconv.Itoa(i))
			context.Current = in.Index(i)
			err = nc.check(context)
			if err != nil {
				return err
			}
			context.Path = context.Path[:len(context.Path)-1]
		}
	case reflect.Map:
		for _, key := range in.MapKeys() {
			if CanBeNil(key.Kind()) && key.IsNil() {
				return fmt.Errorf("Map key is nil")
			}
			context.Path = append(context.Path, key.String())
			context.Current = in.MapIndex(key)
			err = nc.check(context)
			if err != nil {
				return err
			}
			context.Path = context.Path[:len(context.Path)-1]
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

	context.Path = context.Path[:len(context.Path)-1]
	//OK
	return err
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
