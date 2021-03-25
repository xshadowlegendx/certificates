package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/smallstep/certificates/acme"
	"github.com/smallstep/certificates/errs"
	"github.com/smallstep/certificates/logging"
)

// WriteError writes to w a JSON representation of the given error.
func WriteError(w http.ResponseWriter, err error) {
	switch err.(type) {
	case *acme.Error:
		w.Header().Set("Content-Type", "application/problem+json")
	default:
		w.Header().Set("Content-Type", "application/json")
	}
	cause := errors.Cause(err)
	if sc, ok := err.(errs.StatusCoder); ok {
		w.WriteHeader(sc.StatusCode())
	} else {
		if sc, ok := cause.(errs.StatusCoder); ok {
			w.WriteHeader(sc.StatusCode())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Write errors in the response writer
	if rl, ok := w.(logging.ResponseLogger); ok {
		logErr := err
		if u, ok := err.(*acme.Error); ok {
			logErr = u.Err
		}
		rl.WithFields(map[string]interface{}{
			"error": logErr,
		})
		if os.Getenv("STEPDEBUG") == "1" {
			if e, ok := logErr.(errs.StackTracer); ok {
				rl.WithFields(map[string]interface{}{
					"stack-trace": fmt.Sprintf("%+v", e),
				})
			} else {
				if e, ok := cause.(errs.StackTracer); ok {
					rl.WithFields(map[string]interface{}{
						"stack-trace": fmt.Sprintf("%+v", e),
					})
				}
			}
		}
	}

	if err := json.NewEncoder(w).Encode(err); err != nil {
		LogError(w, err)
	}
}
