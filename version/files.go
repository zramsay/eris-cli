package version

var (
	SERVICE_DEFINITIONS = []string{
		"bigchaindb.toml",
		"btcd.toml",
		"bitcoincore.toml",
		"bitcoinclassic.toml",
		"compilers.toml",
		"eth.toml",
		"ipfs.toml",
		"keys.toml",
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
