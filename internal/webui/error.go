package webui

const errorPrefix = "[ERROR] "

func errorText(msg string) string {
	return errorPrefix + msg
}

func errorMsg(err error) string {
	if err == nil {
		return ""
	}
	return errorPrefix + err.Error()
}
