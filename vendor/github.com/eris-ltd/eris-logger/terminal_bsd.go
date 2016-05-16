// +build darwin freebsd openbsd netbsd dragonfly

package logger

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA

type Termios syscall.Termios
