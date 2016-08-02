// +build darwin freebsd openbsd netbsd dragonfly

package logger

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA

// Termios is an exposed syscall structure.
type Termios syscall.Termios
