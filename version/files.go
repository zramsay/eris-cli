// +build !arm

package version

// DEPRECATION warning: to be replaced by pulling from toadserver
// && [eris services ls] which'll query the toadserver for available
// service definition files which can then be imported.
var (
	SERVICE_DEFINITIONS = []string{
		"bigchaindb",
		"btcd",
		"bitcoincore",
		"bitcoinclassic",
		"compilers",
		"geth",
		"ipfs",
		"keys",
		"logspout",
		"logrotate",
		"mindy",
		"openbazaar",
		"rethinkdb",
		"toadserver",
		"tinydns",
		"tor",
		"watchtower",
		"do_not_use",
	}

	ACTION_DEFINITIONS = []string{
		"chain_info",
		"dns_register",
		"keys_list",
	}

	CHAIN_DEFINITIONS = []string{
		"default",
		"config",
		"server_conf",
	}
)
