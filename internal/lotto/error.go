package lotto

type UserInputError struct {
	msg string
}

func (e UserInputError) Error() string {
	return "[ERROR] " + e.msg
}

func NewUserInputError(msg string) error {
	return UserInputError{msg: msg}
}
