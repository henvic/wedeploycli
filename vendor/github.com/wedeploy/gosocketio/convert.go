package gosocketio

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/wedeploy/gosocketio/internal/protocol"
)

func (h *Handler) getFunctionCallArgs(msg *protocol.Message) (is []interface{}, err error) {
	is = []interface{}{}

	if len(h.args) == 0 {
		is = append(is, &struct{}{})
		return is, nil
	}

	var parts []json.RawMessage

	if err := jsonUnmarshalUnpanic(msg.Data, &parts); err != nil {
		return nil, err
	}

	funcArgs := h.Args()

	if len(funcArgs) > len(parts) {
		return is, ErrorInvalidInterface{
			method: msg.Method,
			reason: fmt.Sprintf("message has %d arguments, but listener requires at least %d", len(funcArgs), len(parts)),
		}
	}

	handledArguments := len(parts)

	// cap the number of handled parameters to what the function can handle
	if len(funcArgs) < handledArguments && !h.Variadic {
		handledArguments = len(funcArgs)
	}

	funcsArgsNum := len(funcArgs)

	for c := 0; c < handledArguments; c++ {
		if h.Variadic && c == funcsArgsNum-1 {
			variadic := parts[funcsArgsNum-1:]
			vis, err := getVariadicFunctionCallArgs(variadic, funcArgs[funcsArgsNum-1])

			if err != nil {
				return nil, err
			}

			is = append(is, vis...)
			break
		}

		if err := jsonUnmarshalUnpanic(parts[c], &funcArgs[c]); err != nil {
			return nil, err
		}

		is = append(is, funcArgs[c])

	}

	return is, nil
}

func getVariadicFunctionCallArgs(variadic []json.RawMessage, funcArg interface{}) (vis []interface{}, err error) {
	var vp = []byte("[")

	for l, value := range variadic {
		vp = append(vp, value...)

		if l != len(variadic)-1 {
			vp = append(vp, ',')
		}
	}

	vp = append(vp, ']')

	if err := jsonUnmarshalUnpanic(vp, &funcArg); err != nil {
		return nil, err
	}

	elems := reflect.ValueOf(funcArg).Elem()
	lenFuncArg := elems.Len()

	for c := 0; c < lenFuncArg; c++ {
		addElem := elems.Index(c)
		vis = append(vis, addElem)
	}

	return vis, nil
}

func jsonUnmarshalUnpanic(data []byte, v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	return json.Unmarshal(data, v)
}
