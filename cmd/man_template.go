package commands

import (
	"regexp"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// Rendering Cobra commands directly to groff_mdoc(7).
// This method is preferred here to using Cobra's GenMan() and GenManTree()
// functions. Those unjustifiably use intermediary Markdown format and in turn
// convert it to the old-fashioned and less flexible groff_man(7)
// macros using the "github.com/cpuguy83/go-md2man/md2man" package.

const manPrologue = `.Dd {{date "January 2, 2006"}}
.Os
.Dt MONAX 1
.Sh NAME
.Nm monax
.Nd {{.Short}}
.Sh SYNOPSIS
.Nm
.Cm COMMAND Op FLAG Ns ...
.Sh DESCRIPTION
{{escape .Long}}
`

const manMidsection = `
{{define "commands"}}.Sh COMMANDS
.Bd -filled
To list available commands, ether run
.Cm monax
with no parameters or run
.Cm monax help .
.Ed
{{range .Commands}}.Pp
{{template "command prologue" .}}{{end}}{{end}}

{{define "command prologue"}}.Ss {{.Name}}
{{if .HasSubCommands}}.Pp
.Bd -filled
{{escape .Long}}
.Ed
{{range .Commands}}.D1 \~
{{template "command" .}}{{end}}{{else}}.Pp
{{template "command" .}}{{end}}{{end}}

{{define "command"}}.Pp
.Bd -ragged
.Cm {{.UseLine}}
.Pp
{{escape .Long}}{{if .Aliases}}{{template "aliases" .Aliases}}{{end}}{{if .HasFlags}}
{{template "flags" .NonInheritedFlags}}{{end}}{{if .HasExample}}{{template "examples" .Example}}{{end}}
.Ed
{{end}}

{{define "flags"}}.Bl -ohang
{{range escapeFlags .}}.It {{if .Shorthand}}Fl {{.Shorthand}} , Fl Ns Fl {{.Name}}{{else}}Fl Ns Fl {{.Name}}{{end}}{{if .DefValue}} Ns Op ={{.DefValue}}{{end}}
.D1 {{escape .Usage}}
{{end}}.El{{end}}

{{define "examples"}}
.Pp
.Em Examples :
.Bl -ohang -offset indent
{{range splitExamples .}}.It {{escape .}}
{{end}}.El{{end}}

{{define "aliases"}}
.Pp
(Alternative command name:
.Cm {{index . 0}} . ){{end}}

{{define "global flags"}}.Sh GLOBAL FLAGS
There are a few flags that are available for every single command. Global flags,
just as regular command flags, can start with one or two dashes. Default
values, if meaningful, are provided in square brackets.
{{template "flags" .NonInheritedFlags}}
{{end}}
`

const manEnvironment = `.Sh ENVIRONMENT
.Nm
uses the following environment variables:
.Bl -tag -width "MONAX_PULL_APPROVE"
.It Ev MONAX
.Nm
home directory. Supersedes the default
.Pa $HOME/.monax
path.
.It Ev MONAX_PULL_APPROVE
If set, answers
.Em yes
to the confirmation of whether to pull a Docker image,
potentially a long-running operation. Used, for example, by
.Cm monax services start
and
.Cm monax chains exec
commands.
.It Ev DOCKER_HOST
Docker service host, which
.Nm
connects to.
.It Ev DOCKER_CERT_PATH
Docker secure connection key and certificate. See
.Cm docker-machine env
command or similar.
.It Ev GOPATH
.Xr go 1
source and binary packages path. Used by the
.Cm monax update
self-update command.
.It Ev EDITOR
Default interactive text editor. Used by the
.Cm monax services edit
and
.Cm monax chains edit
commands. If not set, the
.Xr vim 1
editor is used by default.
.El
`

const manEpilogue = `.Sh WWW
http://monax.io
.Sh SEE ALSO
.Xr docker 1 ,
.Xr docker-machine 1 ,
.Xr go 1 ,
.Xr git 1 ,
.Xr http://ipfs.io ,
.Xr http://www.tutum.co ,
.Xr http://quay.io
.Sh AUTHORS
Team of marmots.
.Sh COPYRIGHT
Copyright \[co] 2014-{{date "2006"}} by Monax Industries, Ltd.
.Pp
The software is distributed under the terms of the
.Em GNU General Public License Version 3 .
.Sh BUGS
http://github.com/monax/monax/issues`

var manHelpers = map[string]interface{}{
	"escape": func(text string) string {
		// Replace empty strings with an empty string formatter.
		//
		// Regexp flags:
		//  (?m) - match begin of line and end of line in a multiline buffer.
		//  (?s) - match newlines in a long buffer.
		//
		text = regexp.MustCompile(`(?m)^$`).ReplaceAllString(text, ".Pp")

		// Some Ad-hoc formatting:

		// Reformat example description (as in "$ monax command -- description").
		text = strings.Replace(text, " -- ", "\n.D1 ", -1)

		// Highlight "[monax command ...]".
		text = regexp.MustCompile(`\[(monax\ [^]]+)\]([.,;:]?)[[:space:]]*`).ReplaceAllString(text, "\n.Cm $1 $2\n")

		// Highlight "NOTE:".
		text = regexp.MustCompile(`(NOTE):[[:space:]]*`).ReplaceAllString(text, "\n.Em $1  :\n")

		// Insert a line break before a line which starts with a number and a period
		// (prevent numbered lists to be reformatted).
		text = regexp.MustCompile(`(?m)^(\ ?[[:digit:]]+\..+)`).ReplaceAllString(text, ".Bd -ragged -compact\n$1\n.Ed\n")

		// Replace double new lines with single new lines (if any).
		text = regexp.MustCompile(`(?s)\n\n`).ReplaceAllString(text, "\n")

		// See groff_char(7).
		text = strings.NewReplacer(
			`\`, `\\`,
			`'`, `\[aq]`,
			`"`, `\[dq]`,
			`<`, `\[la]`,
			`>`, `\[ra]`,
		).Replace(text)

		return text
	},
	// Cobra package doesn't provide a way to get a slice of flags;
	// it requires "visiting" them.
	"escapeFlags": func(flagSet *pflag.FlagSet) []*pflag.Flag {
		flags := []*pflag.Flag{}

		flagSet.VisitAll(func(flag *pflag.Flag) {
			// Don't show empty lists or "false" as default flag values.
			if flag.DefValue == "[]" || flag.DefValue == "false" {
				flag.DefValue = ""
			}

			// Don't include the "--help" flag in the man page.
			if flag.Name == "help" {
				return
			}

			flags = append(flags, flag)
		})

		return flags
	},
	"splitExamples": func(examples string) []string {
		return strings.Split(examples, "\n")
	},
	"date": func(format string) string {
		return time.Now().Format(format)
	},
}
