package main

import (
	"errors"
	"fmt"
)

var ErrIllegalMove = errors.New("illegalMove")
var ErrNotSet = errors.New("notSet")

type HTTPError struct {
	Status    int
	Err       string
	CachedMsg []byte
}

func (h HTTPError) Error() string {
	return h.Err
}

var (
	ErrRateLimit     = &HTTPError{Status: 429, Err: "rateLimit"}
	ErrNotAuthorized = &HTTPError{Status: 401, Err: "notAuthorized"}
	ErrBadBody       = &HTTPError{Status: 400, Err: "badBody"}
	ErrNotFound      = &HTTPError{Status: 404, Err: "notFound"}
	// ErrBadLength = &HTTPError{Status: 400, Err: "badLength"}
	// ErrProfanity = &HTTPError{Status: 400, Err: "profane"}
	// ErrBodyMissing = &HTTPError{Status: 400, Err: "bodyMissing"}
	// ErrOAuth2Code = &HTTPError{Status: 400, Err: "noCode"}
	// ErrBadEmail = &HTTPError{Status: 401, Err: "badEmailDomain"}
	// ErrBadLimit = &HTTPError{Status: 400, Err: "badLimit"}
	// ErrBanned = &HTTPError{Status: 401, Err: "banned"}
)

func init() {
	allErrors := []*HTTPError{
		ErrRateLimit,
		ErrNotAuthorized,
		ErrBadBody,
		ErrNotFound,
	}
	for _, err := range allErrors {
		err.CachedMsg = []byte(fmt.Sprintf(`{"error":"%v"}`, err.Err))
	}
}
