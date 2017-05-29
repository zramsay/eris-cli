package maker

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
)

func Tarball(do *definitions.Do) error {
	paths, err := filepath.Glob(filepath.Join(config.ChainsPath, do.Name, strings.ToLower(do.Name)+"_*_*"))
	if err != nil {
		return err
	}

	for _, path := range paths {
		fileP := fmt.Sprintf("%s.tar.gz", path)
		log.WithFields(log.Fields{
			"path": path,
			"file": fileP,
		}).Debug("Making A Tarball")

		dir, err := os.Open(path)
		if err != nil {
			return err
		}
		defer dir.Close()
		files, err := dir.Readdir(0) // grab the files list
		if err != nil {
			return err
		}
		tarfile, err := os.Create(fileP)
		if err != nil {
			return err
		}
		defer tarfile.Close()

		var fileWriter io.WriteCloser = tarfile
		fileWriter = gzip.NewWriter(tarfile)
		defer fileWriter.Close()

		tarfileWriter := tar.NewWriter(fileWriter)
		defer tarfileWriter.Close()

		for _, fileInfo := range files {
			if fileInfo.IsDir() {
				continue
			}

			file, err := os.Open(dir.Name() + string(filepath.Separator) + fileInfo.Name())
			if err != nil {
				return err
			}
			defer file.Close()

			// log.WithField("file", fileInfo.Name()).Debug("Adding File Info to Tarball")
			header := new(tar.Header)
			header.Name = filepath.Base(file.Name())
			header.Size = fileInfo.Size()
			header.Mode = int64(fileInfo.Mode())
			header.ModTime = fileInfo.ModTime()

			if err := tarfileWriter.WriteHeader(header); err != nil {
				return err
			}

			_, err = io.Copy(tarfileWriter, file)
			if err != nil {
				return err
			}
		}

		log.WithField("dir", path).Debug("Removing Directory.")
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

func Zip(do *definitions.Do) error {

	paths, err := filepath.Glob(filepath.Join(config.ChainsPath, do.Name, "*"))
	if err != nil {
		return err
	}

	for _, path := range paths {
		fileP := fmt.Sprintf("%s.zip", path)
		log.WithFields(log.Fields{
			"path": path,
			"file": fileP,
		}).Debug("Making A ZipFile")

		dir, err := os.Open(path)
		if err != nil {
			return err
		}
		defer dir.Close()

		files, err := dir.Readdir(0) // grab the files list
		if err != nil {
			return err
		}

		newfile, err := os.Create(fileP)
		if err != nil {
			return err
		}
		defer newfile.Close()

		zipit := zip.NewWriter(newfile)
		defer zipit.Close()

		for _, fileInfo := range files {
			// log.WithField("file", fileInfo.Name()).Debug("Adding File Info to ZipFile")
			if fileInfo.IsDir() {
				continue
			}

			file, err := os.Open(dir.Name() + string(filepath.Separator) + fileInfo.Name())
			if err != nil {
				return err
			}
			defer file.Close()

			header, err := zip.FileInfoHeader(fileInfo)
			if err != nil {
				return err
			}
			header.Method = zip.Deflate

			writer, err := zipit.CreateHeader(header)
			if err != nil {
				return err
			}
			_, err = io.Copy(writer, file)
			if err != nil {
				return err
			}
		}

		log.WithField("dir", path).Debug("Removing Directory.")
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}
