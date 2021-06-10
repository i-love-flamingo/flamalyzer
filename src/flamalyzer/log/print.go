package log

import "log"

// Println prints a log message.
// Should be used with the `isDebug()` function of the config so messages get logged only if the `--debugFlamalyzer`-flag is given
func Println(msg string, approved bool) {
	if approved {
		log.Println(msg)
	}
}
