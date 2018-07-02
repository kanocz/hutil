package hutil

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
)

// JSON http Content-type header
var JSON = http.Header{"Content-type": {"application/json"}, "Cache-Control": {"no-cache, no-store, must-revalidate"}, "Pragma": {"no-cache"}}

// Error func ouputs error message in json format
func Error(w http.ResponseWriter, r *http.Request, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	encoded, err := json.Marshal(map[string]string{"error": message})
	if nil == err {
		w.Write(encoded)
	}
	_, file, no, _ := runtime.Caller(1)
	log.Printf("%s:%d %d %s %s %s", file, no, code, message, r.URL, r.RemoteAddr)
}

// ErrorLog func ouputs error message in json format and logs separate log comment
func ErrorLog(w http.ResponseWriter, r *http.Request, code int, message string, comment string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	encoded, err := json.Marshal(map[string]string{"error": message})
	if nil == err {
		w.Write(encoded)
	}
	_, file, no, _ := runtime.Caller(1)
	log.Printf("%s:%d %d %s / %s %s %s", file, no, code, message, comment, r.URL, r.RemoteAddr)
}

// ErrorLogErr func ouputs error message in json format and logs error
func ErrorLogErr(w http.ResponseWriter, r *http.Request, code int, message string, err error) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	encoded, err := json.Marshal(map[string]string{"error": message})
	if nil == err {
		w.Write(encoded)
	}
	_, file, no, _ := runtime.Caller(1)
	log.Printf("%s:%d %d %s / %v %s %s", file, no, code, message, err, r.URL, r.RemoteAddr)
}

// WriteRawJSON func ouputs message with 200 status
func WriteRawJSON(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(200)
	w.Write([]byte(message))
}

// OK func simple outputs json with 200 status
func OK(w http.ResponseWriter, r *http.Request, j interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	encoded, err := json.Marshal(j)
	if nil == err {
		w.Write(encoded)
	} else {
		_, file, no, _ := runtime.Caller(1)
		log.Printf("%s:%d error json encoding: %s %s %s", file, no, err.Error(), r.URL, r.RemoteAddr)
	}
}

// SleepyError function outputs json error and needed values for sleepy framework
func SleepyError(status int, msg string, request *http.Request) (int, interface{}, http.Header) {
	_, file, no, _ := runtime.Caller(1)
	log.Printf("%s:%d %d %s %s %s", file, no, status, msg, request.URL, request.RemoteAddr)
	return status, map[string]string{"error": msg}, JSON
}

// SleepyStError function outputs structured json error for sleepy framework
func SleepyStError(msg string, request *http.Request) (int, interface{}, http.Header) {
	_, file, no, _ := runtime.Caller(1)
	log.Printf("%s:%d sleepyStError: %s %s %s", file, no, msg, request.URL, request.RemoteAddr)
	return 200, map[string]string{"status": "error", "message": msg}, JSON
}

// SleepyStResult warps sigle value into {"status": status, key:result}
func SleepyStResult(result interface{}, key string) (int, interface{}, http.Header) {
	return 200, map[string]interface{}{"status": "ok", key: result}, JSON
}

// SleepyStOk return just "status:ok" result
func SleepyStOk() (int, interface{}, http.Header) {
	return 200, map[string]string{"status": "ok"}, JSON
}

// Sleepy1Result warps sigle value into array for rest/sleepy
func Sleepy1Result(result interface{}) (int, interface{}, http.Header) {
	return 200, [1]interface{}{result}, JSON
}

// Sleepy1in1Map warps sigle value into map for rest/sleepy
func Sleepy1in1Map(result interface{}, key string) (int, interface{}, http.Header) {
	return 200, map[string]interface{}{key: result}, JSON
}

// Sleepy1inNMap warps sigle value into array in map for rest/sleepy
func Sleepy1inNMap(result interface{}, key string) (int, interface{}, http.Header) {
	return 200, map[string]interface{}{key: [1]interface{}{result}}, JSON
}

// Request2json reads requests body and unmarshals json from it
func Request2json(request *http.Request, v interface{}) error {
	if nil == request.Body {
		return errors.New("Body is nil")
	}

	defer request.Body.Close()
	body, err := ioutil.ReadAll(request.Body)

	if nil != err {
		return err
	}

	err = json.Unmarshal(body, v)
	if nil != err {
		return err
	}

	return nil
}

// IsHexString allows to check hashes and so on
func IsHexString(str string) bool {
	if "" == str {
		return false
	}
	for _, x := range str {
		if ((x < '0') || (x > '9')) && ((x < 'a') || (x > 'f')) {
			return false
		}
	}
	return true
}

// IsUUID allows to check is string contains only hex symbols and dashes
func IsUUID(str string) bool {
	if "" == str {
		return false
	}
	for _, x := range str {
		if ((x < '0') || (x > '9')) && ((x < 'a') || (x > 'f')) && (x != '-') {
			return false
		}
	}
	return true
}

// IsLangID checks if argument is 2-char string and both of them are a-z
func IsLangID(str string) bool {
	if len(str) != 2 {
		return false
	}
	if (str[0] < 'a') || (str[0] > 'z') {
		return false
	}
	if (str[1] < 'a') || (str[1] > 'z') {
		return false
	}
	return true
}
