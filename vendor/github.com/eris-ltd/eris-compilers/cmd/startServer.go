package cmd

import (
	server "github.com/eris-ltd/eris-compilers/perform"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

func BuildServerCommand() {
	CompilersCmd.AddCommand(serverCmd)
	addServerFlags()
}

var (
	serverPort uint64
	securePort uint64
	noSSL      bool
	secureOnly bool
	serverCert string
	serverKey  string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start a compiler server",
	Run: func(cmd *cobra.Command, args []string) {
		addrUnsecure := ""
		addrSecure := ""

		addrUnsecure += ":" + strconv.FormatUint(serverPort, 10)
		addrSecure += ":" + strconv.FormatUint(securePort, 10)

		if noSSL {
			addrSecure = ""
		} else {
			if secureOnly {
				addrUnsecure = ""
			}
			if _, err := os.Stat(serverKey); os.IsNotExist(err) {
				log.Error("Can't find ssl key %s. Use --no-ssl flag to disable", serverKey)
				os.Exit(1)
			}
			if _, err := os.Stat(serverCert); os.IsNotExist(err) {
				log.Error("Can't find ssl cert %s. Use --no-ssl flag to disable", serverCert)
				os.Exit(1)
			}
		}

		server.StartServer(addrUnsecure, addrSecure, serverCert, serverKey)
	},
}

func addServerFlags() {
	serverCmd.Flags().Uint64VarP(&serverPort, "port", "p", setServerPort(), "set the listening port for http")
	serverCmd.Flags().Uint64VarP(&securePort, "secure-port", "s", setSecurePort(), "set the listening port for https")
	serverCmd.Flags().BoolVarP(&noSSL, "no-ssl", "n", setSSL(), "use only http")
	serverCmd.Flags().BoolVarP(&secureOnly, "secure-only", "o", setSecureOnly(), "use only https")
	serverCmd.Flags().StringVarP(&serverCert, "cert", "c", setDefaultServerCert(), "set the https certificate")
	serverCmd.Flags().StringVarP(&serverKey, "key", "k", setDefaultServerKey(), "set the key to interact with the https certificate")
}

func setServerPort() uint64 {
	return 9099
}

func setSecurePort() uint64 {
	return 9098
}

func setSSL() bool {
	return false
}

func setSecureOnly() bool {
	return false
}

func setDefaultServerCert() string {
	return ""
}

func setDefaultServerKey() string {
	return ""
}
