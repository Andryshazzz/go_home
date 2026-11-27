package apperrors

type FieldError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

type AppError struct {
	HTTPCode int
	Code     string
	Message  string
	Details  []FieldError
}

func (e AppError) WithDetails(details ...FieldError) AppError {
	e.Details = details

	return e
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
		HTTPCode: 400,
		Code:     "invalid_argument",
		Message:  "invalid argument provided",
	}
	ErrInvalidRequest = AppError{
		HTTPCode: 400,
		Code:     "invalid_request",
		Message:  "invalid request",
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
	ErrFileTooLarge = AppError{
		HTTPCode: 413,
		Code:     "file_too_large",
		Message:  "file is too large, the limit is 10 MB",
	}
	ErrUnsupportedFileType = AppError{
		HTTPCode: 415,
		Code:     "unsupported_file_type",
		Message:  "unsupported file type, use jpeg or png",
	}
	ErrAvatarLimitReached = AppError{
		HTTPCode: 409,
		Code:     "avatar_limit_reached",
		Message:  "maximum number of active avatars reached",
	}
	ErrNotFoundAvatar = AppError{
		HTTPCode: 404,
		Code:     "avatar_not_found",
		Message:  "avatar not found",
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
