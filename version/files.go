package version
// DEPRECATION warning: to be replaced by pulling from toadserver
// && [eris services ls] which'll query the toadserver for available
// service definition files which can then be imported.
var (
	SERVICE_DEFINITIONS = []string{
		"bigchaindb.toml",
		"btcd.toml",
		"bitcoincore.toml",
		"bitcoinclassic.toml",
		"compilers.toml",
		"geth.toml",
		"ipfs.toml",
		//"keys.toml", now in eris-cli binary so we can version
		"logspout.toml",
		"logrotate.toml",
		"mindy.toml",
		"openbazaar.toml",
		"rethinkdb.toml",
		"toadserver.toml",
		"tinydns.toml",
		"tor.toml",
		"watchtower.toml",
		"do_not_use.toml",
	}

	ACTION_DEFINITIONS = []string{
		"chain_info.toml",
		"dns_register.toml",
		"keys_list.toml",
	}

	CHAIN_DEFINITIONS = []string{
		"default.toml",
		"config.toml",
		"server_conf.toml",
	}
)
