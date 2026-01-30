package apierror

import "github.com/Alia5/VIIPER/apitypes"

func ErrBadRequest(detail string) apitypes.ApiError {
	return apitypes.ApiError{Status: 400, Title: "Bad Request", Detail: detail}
}
func ErrNotFound(detail string) apitypes.ApiError {
	return apitypes.ApiError{Status: 404, Title: "Not Found", Detail: detail}
}
func ErrConflict(detail string) apitypes.ApiError {
	return apitypes.ApiError{Status: 409, Title: "Conflict", Detail: detail}
}
func ErrInternal(detail string) apitypes.ApiError {
	return apitypes.ApiError{Status: 500, Title: "Internal Server Error", Detail: detail}
}
func ErrUnauthorized(detail string) apitypes.ApiError {
	return apitypes.ApiError{Status: 401, Title: "Unauthorized", Detail: detail}
}

// WrapError normalizes any error into apitypes.ApiError.
func WrapError(err error) apitypes.ApiError {
	if ae, ok := err.(*apitypes.ApiError); ok {
		return *ae
	}
	if ae, ok := err.(apitypes.ApiError); ok {
		return ae
	}
	// Default wrap as internal error
	return ErrInternal(err.Error())
}
