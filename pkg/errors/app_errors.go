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

func (e AppError) Is(target error) bool {
	t, ok := target.(AppError)
	return ok && t.Code == e.Code
}

var (
	ErrNotFoundUser = AppError{
		HTTPCode: 404,
		Code:     "user_not_found",
		Message:  "user not found",
	}
	ErrInvalidArgument = AppError{
		HTTPCode: 422,
		Code:     "invalid_argument",
		Message:  "invalid argument provided",
	}
	ErrUserAlreadyExists = AppError{
		HTTPCode: 409,
		Code:     "user_already_exists",
		Message:  "user already exists",
	}
	ErrInvalidCredentials = AppError{
		HTTPCode: 401,
		Code:     "invalid_credentials",
		Message:  "invalid login or password",
	}
	ErrInvalidToken = AppError{
		HTTPCode: 401,
		Code:     "invalid_token",
		Message:  "invalid or expired token",
	}
	ErrTokenMismatch = AppError{
		HTTPCode: 401,
		Code:     "token_mismatch",
		Message:  "access and refresh token mismatch",
	}
	ErrHashPassword = AppError{
		HTTPCode: 500,
		Code:     "password_hash_error",
		Message:  "failed to hash password",
	}
	ErrCreateUser = AppError{
		HTTPCode: 500,
		Code:     "user_creation_error",
		Message:  "failed to create user",
	}
	ErrGenerateTokens = AppError{
		HTTPCode: 500,
		Code:     "token_generation_error",
		Message:  "failed to generate tokens",
	}
	ErrInternal = AppError{
		HTTPCode: 500,
		Code:     "internal_error",
		Message:  "internal server error",
	}
)
