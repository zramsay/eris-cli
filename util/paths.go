package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/monax/monax/config"
	"github.com/monax/monax/log"
)

func GetFileByNameAndType(typ, name string) string {
	log.WithFields(log.Fields{
		"file": name,
		"type": typ,
	}).Debug("Looking for file")
	files := GetGlobalLevelConfigFilesByType(typ, true)

	for _, file := range files {
		fileBase := strings.Split(filepath.Base(file), ".")[0] // quick and dirty file root
		if fileBase == name {
			log.WithField("file", file).Debug("This file found")
			return file
		}
		log.WithField("file", file).Debug("Group file found")
	}

	return ""
}

// note this function fails silently.
func GetGlobalLevelConfigFilesByType(typ string, withExt bool) []string {
	var path string
	switch typ {
	case "services":
		path = config.ServicesPath
	case "chains":
		path = config.ChainsPath
	}

	files := []string{}
	fileTypes := []string{}

	// TODO [csk]: DRY up how we deal with file extensions
	for _, t := range []string{"*.json", "*.yaml", "*.toml"} {
		fileTypes = append(fileTypes, filepath.Join(path, t))
	}

	for _, t := range fileTypes {
		s, _ := filepath.Glob(t)
		for _, s1 := range s {
			if !withExt {
				s1 = strings.Split(filepath.Base(s1), ".")[0]
			}
			files = append(files, s1)
		}
	}
	return files
}

func MoveOutOfDirAndRmDir(src, dst string) error {
	log.WithFields(log.Fields{
		"from": src,
		"to":   dst,
	}).Info("Move all files/dirs out of a dir and `rm -fr` that dir")
	toMove, err := filepath.Glob(filepath.Join(src, "*"))
	if err != nil {
		return err
	}

	if len(toMove) == 0 {
		log.Debug("No files to move")
	}

	for _, f := range toMove {
		t := filepath.Join(dst, filepath.Base(f))
		log.WithFields(log.Fields{
			"from": f,
			"to":   t,
		}).Debug("Moving")

		// using a copy (read+write) strategy to get around swap partitions and other
		//   problems that cause a simple rename strategy to fail. it is more io overhead
		//   to do this, but for now that is preferable to alternative solutions.
		config.Copy(f, t)
	}

	log.WithField("=>", src).Info("Removing directory")
	return os.RemoveAll(src)
}

// CopyFile copies from src to dst until either EOF is reached on src or
// an error occurs. It verifies that src exists and removes the dst
// if it exists before copying. (Adapted from github.com/docker/pkg/fileutils.)
func CopyFile(src, dst string) (err error) {
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)
	if cleanSrc == cleanDst {
		return nil
	}
	sf, err := os.Open(cleanSrc)
	if err != nil {
		return err
	}
	defer sf.Close()
	if err := os.Remove(cleanDst); err != nil && !os.IsNotExist(err) {
		return err
	}
	df, err := os.Create(cleanDst)
	if err != nil {
		return err
	}
	defer df.Close()
	io.Copy(df, sf)

	return nil
}

// CopySymlink copies the src link to a location pointed to by the dst.
func CopySymlink(src, dst string) error {
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if err := os.Symlink(link, dst); err != nil {
		return err
	}
	return nil
}

// CopyTree copies a directory pointed by the src to a location pointed
// by dst. (Adapted from https://github.com/coreos/rkt/pkg/fileutil)
func CopyTree(src, dst string) error {
	cleanSrc := filepath.Clean(src)

	copyWalker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rootLess := path[len(cleanSrc):]
		if cleanSrc == "." {
			rootLess = path
		}
		target := filepath.Join(dst, rootLess)
		mode := info.Mode()

		switch {
		case mode.IsDir():
			err := os.MkdirAll(target, mode.Perm())
			if err != nil {
				return err
			}
		case mode.IsRegular():
			if err := CopyFile(path, target); err != nil {
				return err
			}
		case mode&os.ModeSymlink == os.ModeSymlink:
			if err := CopySymlink(path, target); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported mode: %v", mode)
		}

		// lchown(2) says that, depending on the linux kernel version, it
		// can change the file's mode also if executed as root. So call
		// os.Chmod after it.
		if mode&os.ModeSymlink != os.ModeSymlink {
			if err := os.Chmod(target, mode); err != nil {
				return err
			}
		}

		return nil
	}

	return filepath.Walk(cleanSrc, copyWalker)
}

// MoveTree copies the src directory to dst and removes the src
// afterwards.
func MoveTree(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}

	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	if err := CopyTree(src, dst); err != nil {
		return err
	}

	return os.RemoveAll(src)
}

// MoveFile copies the src file to a dst location and removes
// the src file afterwards.
func MoveFile(src, dst string) error {
	if err := CopyFile(src, dst); err != nil {
		return err
	}

	return os.Remove(src)
}

// Tilde converts the leading home directory in the path to the `~` symbol.
// Doesn't modify the path on Windows.
func Tilde(path string) string {
	if runtime.GOOS == "windows" {
		return path
	}

	home := config.HomeDir()
	if strings.HasPrefix(path, home) {
		return strings.Replace(path, home, "~", 1)
	}
	return path
}

// DoesDirExist returns true if the directory exists and readable,
// otherwise false.
func DoesDirExist(dir string) bool {
	f, err := os.Stat(dir)
	if err != nil {
		return false
	}
	if !f.IsDir() {
		return false
	}
	return true
}

// DoesFileExist returns true if the file exists and readable,
// otherwise false.
func DoesFileExist(file string) bool {
	if _, err := os.Stat(file); err != nil {
		return false
	}
	return true
}
