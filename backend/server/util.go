package server

import (
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"bytes"
	"reflect"

	"encoding/json"

	"net/http"

	"github.com/gorilla/schema"
	"github.com/heedy/heedy/backend/assets"
	"github.com/sirupsen/logrus"
)

var queryDecoder = schema.NewDecoder()

// apiHeaders writes headers that need to be present in all API requests
func apiHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // All API requests return json
}

// RequestLogger generates a basic logger that holds relevant request info
func RequestLogger(r *http.Request) *logrus.Entry {
	raddr := r.RemoteAddr
	if fwdFor := r.Header.Get("X-Forwarded-For"); fwdFor != "" {
		raddr = fwdFor
	}
	fields := logrus.Fields{"addr": raddr, "path": r.URL.Path, "method": r.Method}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		fields["realip"] = realIP
	}
	return logrus.WithFields(fields)
}

//UnmarshalRequest unmarshals the input data to the given interface
func UnmarshalRequest(request *http.Request, unmarshalTo interface{}) error {
	defer request.Body.Close()

	//Limit requests to the limit given in configuration
	data, err := ioutil.ReadAll(io.LimitReader(request.Body, *assets.Config().RequestBodyByteLimit))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, unmarshalTo)
}

// ErrorResponse is the response given by the server upon an error
type ErrorResponse struct {
	ErrorName        string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ID               string `json:"id,omitempty"`
}

func (er *ErrorResponse) Error() string {
	return er.ErrorName + ":" + er.ErrorDescription
}

// WriteJSONError writes an error message as json. It is assumed that the resulting
// status code is not StatusOK, but rather 4xx
func WriteJSONError(w http.ResponseWriter, r *http.Request, status int, err error) {
	c := CTX(r)

	es := ErrorResponse{
		ErrorName:        "internal_error",
		ErrorDescription: err.Error(),
	}
	myerr := err

	// We can have error types encoded in the error, split with a :
	errs := strings.SplitN(err.Error(), ":", 2)
	if len(errs) > 1 && !strings.Contains(errs[0], " ") {
		es.ErrorName = errs[0]
		es.ErrorDescription = strings.TrimSpace(errs[1])
	}

	if c != nil {
		es.ID = c.RequestID
	}
	jes, err := json.Marshal(&es)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server_error", "error_description": "Failed to create error message"}`))
		if c != nil {
			c.Log.Errorf("Failed to write error message: %s", err)
		} else {
			logrus.Errorf("Failed to write error message: %s", err)
		}
	}

	if c != nil {
		c.Log.Warn(myerr)
	} else {
		logrus.Warn(myerr)
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(jes)))
	w.WriteHeader(status)
	w.Write(jes)
}

// WriteJSON writes response as JSON, or writes the error if such is given
func WriteJSON(w http.ResponseWriter, r *http.Request, data interface{}, err error) {
	if err != nil {
		// By default, an error returns 400
		WriteJSONError(w, r, 400, err)
		return
	}
	jdata, err := json.Marshal(data)
	if err != nil {
		WriteJSONError(w, r, http.StatusInternalServerError, err)
		return
	}

	// golang json unmarshal encodes empty arrays as null
	// https://github.com/golang/go/issues/27589
	// This detects that, and fixes the problem.
	if bytes.Equal(jdata,[]byte("null")) {
		if k := reflect.TypeOf(data).Kind(); (k== reflect.Slice || k==reflect.Array) && reflect.ValueOf(data).Len()==0 {
			jdata = []byte("[]")
		} 
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(jdata)))
	w.WriteHeader(http.StatusOK)
	w.Write(jdata)
}

// WriteResult writes "ok" if the command succeeded, and outputs an error if it didn't
func WriteResult(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		// By default, an error returns 400
		WriteJSONError(w, r, 400, err)
		return
	}
	// success :)
	w.Header().Set("Content-Length", "4")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`"ok"`))

}
