
Casey Kuhlman (0c814baeff 86e81c9b5e d1599792f3 8f90a15b79) [01:24]
projects will be a collection of services (basically a docker-compose file) which we'll use to start whether running in containers or locally.

Casey Kuhlman (0c814baeff 86e81c9b5e d1599792f3 8f90a15b79) [01:25]
services will be wrapped smaller services configured and started according to the "terms" of the project file (which probably will be in the package.json of the dapp).

Casey Kuhlman (0c814baeff 86e81c9b5e d1599792f3 8f90a15b79) [01:25]
services need not be started via containers

Casey Kuhlman (0c814baeff 86e81c9b5e d1599792f3 8f90a15b79) [01:26]
services > chains will bring the chain management work out of epm and into this tool but utilized via child processes or containers.

Casey Kuhlman (0c814baeff 86e81c9b5e d1599792f3 8f90a15b79) [01:28]
so chain checkout, chain new, all that stuff will come out of epm and into services > chains

Casey Kuhlman (0c814baeff 86e81c9b5e d1599792f3 8f90a15b79) [01:31]01:31
one thing i'm thinking through is whether we can isolate the configs for the blockchains as much as possible (like their config structs) .. this way we can pull them into the cli without having to worry about pulling in the crypto. need to look how tendermint and eth currently doing it.