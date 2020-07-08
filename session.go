package hutil

//go:generate msgp -io=false -tests=false

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const sessionTimeout = 3600 * 24 * 30

// SessionInfo is type stored to redis
type SessionInfo struct {
	UID int64 `msg:"uid"`
}

// SessionNew returns id of newly created session
func SessionNew(data interface{}, w http.ResponseWriter) (string, error) {
	id, err := uuid.NewUUID()
	if nil != err {
		return "", err
	}

	key := id.String()
	err = CacheSetEncoded("session_"+key, data, sessionTimeout)
	if nil != err {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{Name: "authtoken", Value: key, Expires: time.Now().AddDate(0, 0, 30), Path: "/"})

	return key, nil
}

// SessionLogin creates session and saves uid
func SessionLogin(uid int64, w http.ResponseWriter) (string, error) {
	return SessionNew(SessionInfo{uid}, w)
}

// SessionLogout removed sid cookie and data from redis
func SessionLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("authtoken")
	if nil != err {
		// no login cookie present
		// TODO: sid in get
		return
	}

	CacheDelete("session_" + cookie.Value)
	http.SetCookie(w, &http.Cookie{Name: "authtoken", Value: "", Expires: time.Unix(0, 0), MaxAge: -1, Path: "/"})
}

// SessionGetUser returns all user Id based on session id
func SessionGetUser(r *http.Request) (int64, error) {
	cookie, err := r.Cookie("authtoken")
	if nil != err {
		return 0, err
	}

	if !IsUUID(cookie.Value) {
		return 0, errors.New("Invalid session key")
	}

	sessinfo := SessionInfo{}
	if err := CacheGetEncoded("session_"+cookie.Value, &sessinfo); nil != err {
		return 0, errors.New("Invalid session id: " + err.Error())
	}

	return sessinfo.UID, nil
}
