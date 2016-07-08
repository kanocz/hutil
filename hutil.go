package hutil

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
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
	}
}

// SleepyError function outputs json error and needed values for sleepy framework
func SleepyError(status int, msg string, request *http.Request) (int, interface{}, http.Header) {
	log.Println(status, msg, request.URL, request.RemoteAddr)
	return status, map[string]string{"error": msg}, JSON
}

// SleepyStError function outputs structured json error for sleepy framework
func SleepyStError(msg string, request *http.Request) (int, interface{}, http.Header) {
	log.Println("sleepyStError:", msg, request.URL, request.RemoteAddr)
	return 200, map[string]string{"status": "error", "message": msg}, JSON
}

// SleepyStResult warps sigle value into {"status": status, key:result}
func SleepyStResult(result interface{}, key string) (int, interface{}, http.Header) {
	return 200, map[string]interface{}{"status": "ok", key: result}, JSON
}

// SleepyStOk return just "status:ok" result
func SleepyStOk(result interface{}, key string) (int, interface{}, http.Header) {
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
