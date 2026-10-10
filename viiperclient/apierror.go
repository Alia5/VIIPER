package viiperclient

func ErrBadRequest(detail string) APIError {
	return APIError{Status: 400, Title: "Bad Request", Detail: detail}
}
func ErrNotFound(detail string) APIError {
	return APIError{Status: 404, Title: "Not Found", Detail: detail}
}
func ErrConflict(detail string) APIError {
	return APIError{Status: 409, Title: "Conflict", Detail: detail}
}
func ErrInternal(detail string) APIError {
	return APIError{Status: 500, Title: "Internal Server Error", Detail: detail}
}
func ErrUnauthorized(detail string) APIError {
	return APIError{Status: 401, Title: "Unauthorized", Detail: detail}
}

// WrapError normalizes any error into viipertypes.ApiError.
func WrapError(err error) APIError {
	if ae, ok := err.(*APIError); ok {
		return *ae
	}
	if ae, ok := err.(APIError); ok {
		return ae
	}
	// Default wrap as internal error
	return ErrInternal(err.Error())
}
