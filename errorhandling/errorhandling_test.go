package errorhandling

import (
	"errors"
	"testing"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
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
	var err = errwrap.Wrapf("Wrapped: {{err}}", errors.New("my error"))
	var handle = Handle(err)
	var want = "Wrapped: my error"

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
	var err = errwrap.Wrapf("Wrapped: {{err}}", errors.New("my error"))
	var handle = Handle(err)
	var want = "Wrapped: my error"

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

	var err = &apihelper.APIFault{
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

	var want = "WeDeploy API error: 404 Not Found (GET http://example.com/)\n" +
		"\tDocument not found: documentNotFound"

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

	var err = errwrap.Wrapf("Wrapped: {{err}}", &apihelper.APIFault{
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

	var want = "WeDeploy API error: 404 Not Found (GET http://example.com/)\n" +
		"\tDocument not found: documentNotFound"

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

	var err = &apihelper.APIFault{
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

	var err = errwrap.Wrapf("Wrapped error: {{err}}", &apihelper.APIFault{
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

	var want = "Credential not valid for payment"

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

	var err = errwrap.Wrapf("Wrapped error: {{err}}", &apihelper.APIFault{
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

	var want = "Credit card not valid for payment"

	if want != got.Error() {
		t.Errorf("Wanted %v, got %v instead", want, got)
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

	var err = errwrap.Wrapf("Wrapped error: {{err}}", &apihelper.APIFault{
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

	var want = "Amex credit card is not accepted"

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

	var err = errwrap.Wrapf("Wrapped error: {{err}}", &apihelper.APIFault{
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

	var want = "Invalid credential"

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

	var err = &apihelper.APIFault{
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

func restoreOriginalErrorMessages() {
	errorReasonMessage = originalErrorReasonMessage
	errorReasonCommandMessageOverrides = originalErrorReasonCommandMessageOverrides
}
