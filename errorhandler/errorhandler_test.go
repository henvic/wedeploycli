package errorhandler

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/wedeploycli/apihelper"
)

var (
	originalErrorReasonMessage                 = errorReasonMessage
	originalErrorReasonCommandMessageOverrides = errorReasonCommandMessageOverrides
)

func TestHandleNil(t *testing.T) {
	CommandName = ""
	if got := Handle(nil); got != nil {
		t.Errorf("Expected nil for handling nil error, got %v instead", got)
	}
}

func TestHandleNilOnCommand(t *testing.T) {
	CommandName = "create"
	if got := Handle(nil); got != nil {
		t.Errorf("Expected nil for handling nil error, got %v instead", got)
	}
}

func TestHandleGenericErrorNotHumanized(t *testing.T) {
	CommandName = ""
	var err = errors.New("my error")
	var handle = Handle(err)
	var want = "my error"

	if handle.Error() != want {
		t.Errorf("Error message %v differ from expected value %v", handle.Error(), want)
	}
}

func TestHandleWrappedGenericErrorNotHumanized(t *testing.T) {
	CommandName = ""
	var err = errwrap.Wrapf("wrapped: {{err}}", errors.New("my error"))
	var handle = Handle(err)
	var want = "wrapped: my error"

	if handle.Error() != want {
		t.Errorf("Error message %v differ from expected value %v", handle.Error(), want)
	}
}

func TestHandleGenericErrorNotHumanizedOnCommandUnknown(t *testing.T) {
	CommandName = "foo"
	var err = errors.New("my error")
	var handle = Handle(err)
	var want = "my error"

	if handle.Error() != want {
		t.Errorf("Error message %v differ from expected value %v", handle.Error(), want)
	}
}

func TestHandleWrappedGenericErrorNotHumanizedOnCommandUnknown(t *testing.T) {
	CommandName = "foo"
	var err = errwrap.Wrapf("wrapped: {{err}}", errors.New("my error"))
	var handle = Handle(err)
	var want = "wrapped: my error"

	if handle.Error() != want {
		t.Errorf("Error message %v differ from expected value %v", handle.Error(), want)
	}
}

func TestHandleAPIFaultGenericErrorMessageNotFound(t *testing.T) {
	CommandName = "payment"
	defer restoreOriginalErrorMessages()

	errorReasonMessage = messages{}

	errorReasonCommandMessageOverrides = map[string]messages{
		"payment": messages{
			"invalidCredential": "Invalid credential",
		},
	}

	var err = apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  404,
		Message: "Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "documentNotFound",
				Context: apihelper.APIFaultErrorContext{
					"message": "Document not found",
				},
			},
		},
	}

	var got = Handle(err)

	var want = "Document not found"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestHandleAPIFaultErrorPrintedInside(t *testing.T) {
	CommandName = "payment"
	defer restoreOriginalErrorMessages()

	errorReasonMessage = messages{
		"innerError": "This message contains an error inside: {{.Err}}",
	}

	var err = apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  404,
		Message: "Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "innerError",
				Context: apihelper.APIFaultErrorContext{
					"message": "message in 123",
				},
			},
		},
	}

	var got = Handle(err)

	var want = "This message contains an error inside: message in 123"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestHandleWrappedAPIFaultGenericErrorMessageNotFound(t *testing.T) {
	CommandName = "payment"
	defer restoreOriginalErrorMessages()

	errorReasonMessage = messages{}

	errorReasonCommandMessageOverrides = map[string]messages{
		"payment": messages{
			"invalidCredential": "Invalid credential",
		},
	}

	var err = errwrap.Wrapf("wrapped: {{err}}", apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  404,
		Message: "Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "documentNotFound",
				Context: apihelper.APIFaultErrorContext{
					"message": "Document not found",
				},
			},
		},
	})

	var got = Handle(err)

	var want = "wrapped: Document not found"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestHandleAPIFaultGenericErrorFound(t *testing.T) {
	CommandName = "payment"
	defer restoreOriginalErrorMessages()

	errorReasonMessage = messages{
		"documentNotFound": "Document not found",
	}

	errorReasonCommandMessageOverrides = map[string]messages{
		"payment": messages{
			"documentNotFound":  "Payment not found",
			"invalidCredential": "Credential not valid for payment",
		},
	}

	var err = apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  404,
		Message: "Payment Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "invalidCredential",
				Context: apihelper.APIFaultErrorContext{
					"message": "Credential not valid",
				},
			},
		},
	}

	var got = Handle(err)

	var want = "Credential not valid for payment"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestHandleAPIWrappedFaultGenericErrorFound(t *testing.T) {
	CommandName = "payment"
	defer restoreOriginalErrorMessages()

	errorReasonMessage = messages{
		"documentNotFound": "Document not found",
	}

	errorReasonCommandMessageOverrides = map[string]messages{
		"payment": messages{
			"documentNotFound":  "Payment not found",
			"invalidCredential": "Credential not valid for payment",
		},
	}

	var err = errwrap.Wrapf("wrapped error: {{err}}", apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  404,
		Message: "Payment Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "invalidCredential",
				Context: apihelper.APIFaultErrorContext{
					"message": "Credential not valid",
				},
			},
		},
	})

	var got = Handle(err)

	var want = "wrapped error: Credential not valid for payment"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestHandleAPIWrappedFaultGenericErrorFoundNested(t *testing.T) {
	CommandName = "payment credit"
	defer restoreOriginalErrorMessages()

	errorReasonMessage = messages{
		"documentNotFound": "Document not found",
	}

	errorReasonCommandMessageOverrides = map[string]messages{
		"payment": messages{
			"documentNotFound":  "Payment not found",
			"invalidCredential": "Credential not valid for payment",
		},
		"payment credit": messages{
			"documentNotFound":  "Payment not found",
			"invalidCredential": "Credit card not valid for payment",
		},
	}

	var err = errwrap.Wrapf("wrapped error: {{err}}", apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  404,
		Message: "Payment Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "invalidCredential",
				Context: apihelper.APIFaultErrorContext{
					"message": "Credential not valid",
				},
			},
		},
	})

	var got = Handle(err)

	var want = "wrapped error: Credit card not valid for payment"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}

	if reflect.TypeOf(got).String() != "*errwrap.wrappedError" {
		t.Errorf("Expected error to be wrapped")
	}

	var af = errwrap.GetType(got, apihelper.APIFault{})

	if af == nil {
		t.Errorf("Expected error to be of apihelper.APIFault{} time")
	}
}

func TestHandleAPIWrappedFaultGenericErrorFoundDeepNested(t *testing.T) {
	CommandName = "payment credit amex"
	defer restoreOriginalErrorMessages()

	errorReasonMessage = messages{
		"documentNotFound": "Document not found",
	}

	errorReasonCommandMessageOverrides = map[string]messages{
		"payment": messages{
			"documentNotFound":  "Payment not found",
			"invalidCredential": "Credential not valid for payment",
		},
		"payment credit": messages{
			"documentNotFound":  "Payment not found",
			"invalidCredential": "Credit card not valid for payment",
		},
		"payment credit amex": messages{
			"documentNotFound":  "Payment not found",
			"invalidCredential": "Amex credit card is not accepted",
		},
	}

	var err = errwrap.Wrapf("wrapped error: {{err}}", apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  404,
		Message: "Payment Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "invalidCredential",
				Context: apihelper.APIFaultErrorContext{
					"message": "Credential not valid",
				},
			},
		},
	})

	var got = Handle(err)

	var want = "wrapped error: Amex credit card is not accepted"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestHandleAPIWrappedFaultGenericErrorFoundDeepNestedNoOverride(t *testing.T) {
	CommandName = "payment credit amex"
	defer restoreOriginalErrorMessages()

	errorReasonMessage = messages{
		"documentNotFound":  "Document not found",
		"invalidCredential": "Invalid credential",
	}

	errorReasonCommandMessageOverrides = map[string]messages{
		"payment": messages{
			"documentNotFound": "Payment not found",
		},
		"payment credit": messages{
			"documentNotFound": "Payment not found",
		},
		"payment credit amex": messages{
			"documentNotFound": "Payment not found",
		},
	}

	var err = errwrap.Wrapf("wrapped error: {{err}}", apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  404,
		Message: "Payment Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "invalidCredential",
				Context: apihelper.APIFaultErrorContext{
					"message": "Credential not valid",
				},
			},
		},
	})

	var got = Handle(err)

	var want = "wrapped error: Invalid credential"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestHandleAPIFaultCommandOverridesErrorMessage(t *testing.T) {
	CommandName = "payment"
	defer restoreOriginalErrorMessages()

	errorReasonMessage = messages{
		"invalidCredential": "Invalid credential for all",
	}

	errorReasonCommandMessageOverrides = map[string]messages{
		"payment": messages{
			"documentNotFound":  "Payment not found",
			"invalidCredential": "Invalid credential for payment",
		},
		"other": messages{
			"documentNotFound":  "Not this payment not found",
			"invalidCredential": "Not this invalid credential for payment",
		},
	}

	var err = apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  403,
		Message: "Invalid credential",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "documentNotFound",
				Context: apihelper.APIFaultErrorContext{
					"message": "Document not found",
				},
			},
		},
	}

	var got = Handle(err)

	var want = "Payment not found"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestHandleInvalidParameter(t *testing.T) {
	CommandName = "payment credit"

	var err = errwrap.Wrapf("wrapped error: {{err}}", apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  http.StatusBadRequest,
		Message: "Payment Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "invalidParameter",
				Context: apihelper.APIFaultErrorContext{
					"param": "creditcard",
					"value": "@",
				},
			},
		},
	})

	var got = Handle(err)

	var want = `wrapped error: Invalid value "@" for parameter "creditcard"`

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}

	if reflect.TypeOf(got).String() != "*errwrap.wrappedError" {
		t.Errorf("Expected error to be wrapped")
	}

	var af = errwrap.GetType(got, apihelper.APIFault{})

	if af == nil {
		t.Errorf("Expected error to be of apihelper.APIFault{} time")
	}
}

func TestHandleInvalidParameterContextPriority(t *testing.T) {
	CommandName = "payment credit"

	var err = errwrap.Wrapf("wrapped error: {{err}}", apihelper.APIFault{
		Method:  "GET",
		URL:     "http://example.com/",
		Status:  http.StatusBadRequest,
		Message: "Payment Not Found",
		Errors: apihelper.APIFaultErrors{
			apihelper.APIFaultError{
				Reason: "invalidParameter",
				Context: apihelper.APIFaultErrorContext{
					"param":   "creditcard",
					"value":   "@",
					"message": "Credit card parameter is not valid",
				},
			},
		},
	})

	var got = Handle(err)

	var want = "wrapped error: Credit card parameter is not valid"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}

	if reflect.TypeOf(got).String() != "*errwrap.wrappedError" {
		t.Errorf("Expected error to be wrapped")
	}

	var af = errwrap.GetType(got, apihelper.APIFault{})

	if af == nil {
		t.Errorf("Expected error to be of apihelper.APIFault{} time")
	}
}

func TestHandleURLError(t *testing.T) {
	err := &url.Error{
		Op:  "OPTIONS",
		URL: "http://example.com/",
		Err: errors.New("not timeout"),
	}

	got := Handle(err)

	if want := "network connection error"; err == nil || !strings.Contains(got.Error(), want) {
		t.Errorf("Expected error to contain %v, got %v instead", want, got)
	}
}

// from https://github.com/golang/go/blob/da0d1a4/src/net/url/url_test.go#L1538-L1543
type timeoutError struct {
	timeout bool
}

func (e *timeoutError) Error() string { return "timeout error" }
func (e *timeoutError) Timeout() bool { return e.timeout }

func TestHandleURLErrorTimeout(t *testing.T) {
	err := &url.Error{
		Op:  "OPTIONS",
		URL: "http://example.com/",
		Err: &timeoutError{
			timeout: true,
		},
	}

	got := Handle(err)

	if want := "network connection timed out"; got == nil || !strings.Contains(got.Error(), want) {
		t.Errorf("Expected error to contain %v, got %v instead", want, got)
	}
}

func restoreOriginalErrorMessages() {
	errorReasonMessage = originalErrorReasonMessage
	errorReasonCommandMessageOverrides = originalErrorReasonCommandMessageOverrides
}
