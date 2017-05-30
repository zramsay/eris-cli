package log

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
)

// MonaxFormatter is a custom logger implementation.
type MonaxFormatter struct {
	// Set to true to ignore TTY checks for color highlights.
	Color bool

	// Set to true to ignore level set by the log.SetLevel() function.
	IgnoreLevel bool
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

// Format implements the logger.Formatter interface. It returns a formatted
// log line as a slice of bytes.
func (f MonaxFormatter) Format(entry *Entry) (out []byte, err error) {
	// Check if output is necessary.
	if !f.IgnoreLevel && entry.Level > GetLevel() {
		return []byte{}, nil
	}

	// Sort tag names in alphabetical order.
	var keys []string
	for key := range entry.Data {
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
func (f MonaxFormatter) Highlight(tag, comment string) (adjustedOffset int, text string) {
	tagDecorated := tag
	commentDecorated := comment

	// Use color formatting if specified and if connected to the terminal.
	if f.Color && IsTerminal() {
		tagDecorated = fmt.Sprintf("%s%s%s", escTag, tag, escReset)
		commentDecorated = fmt.Sprintf("%s%s%s", escBold, comment, escReset)
	}

	if tag == arrowTag {
		return offset + 2, commentDecorated
	}
	return offset - len(tag) + 1, fmt.Sprintf("%s=%s", tagDecorated, commentDecorated)
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
