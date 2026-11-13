package apperrors

type AppError struct {
	HTTPCode int
	Code     string
	Message  string
	Details  map[string]any
}

func (e AppError) Error() string {
	return e.Message
}

var (
	ErrNotFoundUser = AppError{
		HTTPCode: 404,
		Code:     "user_not_found",
		Message:  "user not found",
		Details:  nil,
	}

	ErrInvalidArgument = AppError{
		HTTPCode: 422,
		Code:     "invalid_argument",
		Message:  "invalid argument provided",
		Details:  nil,
	}

	ErrUserAlreadyExists = AppError{
		HTTPCode: 409,
		Code:     "user_already_exists",
		Message:  "user with this phone number already exists",
		Details:  nil,
	}

	ErrInvalidCredentials = AppError{
		HTTPCode: 401,
		Code:     "invalid_credentials",
		Message:  "invalid phone number or password",
		Details:  nil,
	}

	ErrInvalidToken = AppError{
		HTTPCode: 401,
		Code:     "invalid_token",
		Message:  "invalid or expired token",
		Details:  nil,
	}

	ErrTokenMismatch = AppError{
		HTTPCode: 401,
		Code:     "token_mismatch",
		Message:  "access and refresh token mismatch",
		Details:  nil,
	}
)
