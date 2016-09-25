package keys

import (
	"bytes"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/definitions"
	srv "github.com/eris-ltd/eris-cli/services"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

func GenerateKey(do *definitions.Do) error {
	do.Name = "keys"

	if err := srv.EnsureRunning(do); err != nil {
		return err
	}
	// TODO implement
	// if do.Password {}

	buf, err := srv.ExecHandler(do.Name, []string{"eris-keys", "gen", "--no-pass"})
	if err != nil {
		return err
	}

	if do.Save {
		addr := new(bytes.Buffer)
		addr.ReadFrom(buf)

		doExport := definitions.NowDo()
		doExport.Address = strings.TrimSpace(addr.String())

		log.WithField("=>", doExport.Address).Warn("Saving key to host")
		if err := ExportKey(doExport); err != nil {
			return err
		}
	}

	io.Copy(config.Global.Writer, buf)

	return nil
}

func ExportKey(do *definitions.Do) error {
	do.Name = "keys"
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}

	if do.All && do.Address == "" {
		do.Destination = common.KeysPath
		do.Source = path.Join(common.KeysContainerPath)
	} else {
		do.Destination = common.KeysDataPath
		do.Source = path.Join(common.KeysContainerPath, do.Address)
	}
	return data.ExportData(do)
}

func ImportKey(do *definitions.Do) error {
	do.Name = "keys"
	if err := srv.EnsureRunning(do); err != nil {
		return err
	}

	if do.All && do.Address == "" {
		doLs := definitions.NowDo()
		doLs.Container = false
		doLs.Host = true
		doLs.Quiet = true
		result, err := ListKeys(doLs)
		if err != nil {
			return err
		}

		for _, addr := range result {
			do.Source = filepath.Join(common.KeysDataPath, addr)
			do.Destination = path.Join(common.KeysContainerPath, addr)
			if err := data.ImportData(do); err != nil {
				return err
			}
		}
	} else {
		do.Source = filepath.Join(common.KeysDataPath, do.Address)
		do.Destination = path.Join(common.KeysContainerPath, do.Address)
		if err := data.ImportData(do); err != nil {
			return err
		}
	}

	return nil
}
