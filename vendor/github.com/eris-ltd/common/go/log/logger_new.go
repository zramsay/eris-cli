package log

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"

	"github.com/docker/docker/pkg/term"

	log "github.com/Sirupsen/logrus"
)

type ErisFormatter struct {
	// Override the logging level to be able to collect logger messages
	// at a level lower (more verbose) than the one is used for the console.
	Level log.Level

	// Provide color formatting for log messages.
	Color bool
}

const (
	arrowTag = "=>"

	// Where to place `tag=comment` on screen.
	offset = 44

	// Space between a log message and a tag name
	spacing = 4
)

var (
	// See terminfo(5) for the list of commands.
	escReset = tput("sgr0")
	escBold  = tput("bold")
	// http://worldwidemann.com/content/images/2015/03/finalterm-colors.png
	escTag = tput("setaf", 241)
)

// RemoteFormatter returns the ErisFormatter with settings
// suitable for the remote logging.
func RemoteFormatter(level log.Level) ErisFormatter {
	return ErisFormatter{Level: level}
}

// ConsoleFormatter returns the ErisFormatter with settings
// suitable for the console media.
func ConsoleFormatter(level log.Level) ErisFormatter {
	return ErisFormatter{Level: level, Color: true}
}

// Format implements the logrus.Formatter interface. It returns a formatted
// log line as a slice of bytes.
func (f ErisFormatter) Format(entry *log.Entry) (out []byte, err error) {
	// Check if output is necessary.
	if entry.Level > f.Level {
		return []byte{}, nil
	}

	// Sort tag names in alphabetical order.
	var keys []string
	for key, _ := range entry.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Display a message and the first tag.
	if len(entry.Data) > 0 {
		tag, comment := keys[0], fmt.Sprintf("%v", entry.Data[keys[0]])

		// Highlight the tag.
		adjustedOffset, text := f.Highlight(tag, comment)

		if len(entry.Message) < adjustedOffset-spacing {
			// Message with the tag inline.
			out = append(out, fmt.Sprintf("%-*s%s\n", adjustedOffset, entry.Message, text)...)
		} else {
			// Message with the tag on a separate line.
			out = append(out, fmt.Sprintf("%s\n%-*s%s\n", entry.Message, adjustedOffset, "", text)...)
		}

		// Remove the used tag name.
		keys = keys[1:]
	} else {
		// Message without tags.
		out = append(out, fmt.Sprintln(entry.Message)...)
	}

	// Display every other tag on a separate line.
	for _, key := range keys {
		// Highlight the tag.
		adjustedOffset, text := f.Highlight(key, fmt.Sprintf("%v", entry.Data[key]))

		out = append(out, fmt.Sprintf("%-*s%s\n", adjustedOffset, "", text)...)
	}

	return out, nil
}

// Highlight emphasizes a tag and a comment. It returns the highlighted
// text along with an offset where to place it on screen.
func (f ErisFormatter) Highlight(tag, comment string) (adjustedOffset int, text string) {
	tagDecorated := tag
	commentDecorated := comment

	// Use color formatting if specified and if connected to the terminal.
	if f.Color && term.IsTerminal(os.Stdout.Fd()) && term.IsTerminal(os.Stderr.Fd()) {
		tagDecorated = fmt.Sprintf("%s%s%s", escTag, tag, escReset)
		commentDecorated = fmt.Sprintf("%s%s%s", escBold, comment, escReset)
	}

	if tag == arrowTag {
		return offset + 2, fmt.Sprintf("%s", commentDecorated)
	} else {
		return offset - len(tag) + 1, fmt.Sprintf("%s=%s", tagDecorated, commentDecorated)
	}
}

// tput asks the terminfo database for a particular escape sequence.
func tput(command string, params ...interface{}) []byte {
	args := []string{command}

	if runtime.GOOS == "windows" {
		return []byte{}
	}

	for _, param := range params {
		switch param.(type) {
		case string:
			args = append(args, param.(string))
		case int:
			args = append(args, fmt.Sprintf("%d", param))
		}
	}

	out, err := exec.Command("tput", args...).Output()
	if err != nil {
		return []byte{}
	}

	return out
}
