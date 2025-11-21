package api

import "viiper/pkg/apitypes"

// Factory helpers returning *apitypes.ApiError (single canonical error type).
func ErrBadRequest(detail string) *apitypes.ApiError {
	return &apitypes.ApiError{Status: 400, Title: "Bad Request", Detail: detail}
}
func ErrNotFound(detail string) *apitypes.ApiError {
	return &apitypes.ApiError{Status: 404, Title: "Not Found", Detail: detail}
}
func ErrConflict(detail string) *apitypes.ApiError {
	return &apitypes.ApiError{Status: 409, Title: "Conflict", Detail: detail}
}
func ErrInternal(detail string) *apitypes.ApiError {
	return &apitypes.ApiError{Status: 500, Title: "Internal Server Error", Detail: detail}
}

// WrapError normalizes any error into *apitypes.ApiError.
func WrapError(err error) *apitypes.ApiError {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*apitypes.ApiError); ok {
		return ae
	}
	// Default wrap as internal error
	return ErrInternal(err.Error())
}
