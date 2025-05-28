package handlers

import "net/http"

var AuthKey string

func SetAuthKey(k string) {
	AuthKey = k
}

func IsAuthorized(r *http.Request) bool {
	return r.Header.Get("Authorization") == "Bearer "+AuthKey
}
