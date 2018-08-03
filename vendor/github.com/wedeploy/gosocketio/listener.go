package gosocketio

import (
	"errors"
	"fmt"
	"reflect"
)

// Handler for the message
type Handler struct {
	Func reflect.Value

	Out      bool
	Variadic bool

	args []reflect.Type
}

// ErrorArgument is used when trying to create a non-function listener with an invalid parameter
type ErrorArgument struct {
	kind reflect.Kind
}

func (e ErrorArgument) Error() string {
	return fmt.Sprintf(`invalid argument type "%v" for handler`, e.kind)
}

type ErrorInvalidInterface struct {
	method string
	reason string
}

func (e ErrorInvalidInterface) Error() string {
	return fmt.Sprintf(`invalid interface for handling request for "%s" call: %s`, e.method, e.reason)
}

func (e ErrorInvalidInterface) Method() string {
	return e.method
}

// ErrorNotFunction is used when trying to create a non-function listener
type ErrorNotFunction struct {
	kind reflect.Kind
}

func (e ErrorNotFunction) Error() string {
	return fmt.Sprintf("listener is a %v instead of a function", e.kind)
}

func isVariadicNonInterface(fType reflect.Type) bool {
	if !fType.IsVariadic() {
		return false
	}

	numIn := fType.NumIn()
	arg := fType.In(numIn - 1)

	if arg.Kind() != reflect.Slice {
		return false
	}

	return arg.Elem().Kind() != reflect.Interface
}

// NewHandler creates a new listener
func NewHandler(f interface{}) (*Handler, error) {
	fValue := reflect.ValueOf(f)

	if fValue.Kind() != reflect.Func {
		return nil, ErrorNotFunction{fValue.Kind()}
	}

	fType := fValue.Type()

	if fType.NumOut() > 1 {
		return nil, errors.New("f should return not more than one value")
	}

	if err := checkHandlerInputParams(fType); err != nil {
		return nil, err
	}

	numIn := fType.NumIn()

	if isVariadicNonInterface(fType) {
		return nil, errors.New("support for variadic is only partially implemented; see https://github.com/wedeploy/gosocket.io-client-go/issues/1")
	}

	h := &Handler{
		Func:     fValue,
		Out:      fType.NumOut() == 1,
		Variadic: fType.IsVariadic(),
	}

	for c := 0; c < numIn; c++ {
		h.args = append(h.args, fType.In(c))
	}

	return h, nil
}

// Call function
func (h *Handler) Call(args ...interface{}) []reflect.Value {
	// nil is untyped, so use the default empty value of correct type
	if args == nil {
		args = h.Args()
	}

	a := []reflect.Value{}

	if len(h.args) != 0 {
		a = h.matchArgs(args)
	}

	return h.Func.Call(a)
}

func (h *Handler) matchArgs(args []interface{}) (a []reflect.Value) {
	lengthFuncArgs := len(h.args)
	for pos := range h.args {
		if h.Variadic && pos == lengthFuncArgs-1 {
			break
		}

		rfv := reflect.ValueOf(args[pos])
		a = append(a, rfv.Elem())

	}

	if !h.Variadic {
		return a
	}

	for variadicPos := lengthFuncArgs - 1; variadicPos < len(args); variadicPos++ {
		a = append(a, reflect.ValueOf(args[variadicPos]))
	}

	return a
}

// Args returns the interfaces for the given function
func (h *Handler) Args() []interface{} {
	var interfaces = []interface{}{}

	for _, a := range h.args {
		interfaces = append(interfaces, reflect.New(a).Interface())
	}

	return interfaces
}

func checkParamKind(kind reflect.Kind) error {
	switch kind {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Interface,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice,
		reflect.String,
		reflect.Struct:
		return nil
	}

	return ErrorArgument{
		kind,
	}
}

func checkHandlerInputParams(fType reflect.Type) error {
	num := fType.NumIn()

	for c := 0; c < num; c++ {
		in := fType.In(c)
		if err := checkParamKind(in.Kind()); err != nil {
			return err
		}
	}

	return nil
}
