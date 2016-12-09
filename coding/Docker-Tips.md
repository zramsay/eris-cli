[Eris Docker Hub](https://registry.hub.docker.com/repos/eris/) || [EI Docker Hub](https://registry.hub.docker.com/repos/erisindustries/)

[Intro to Docker](https://medium.freecodecamp.com/a-beginner-friendly-introduction-to-containers-vms-and-docker-79a9e3e119b#.6zfg8vnq6)

## Install the Tools

### Docker

The Eris stack does not work well with the Ubuntu `apt-get install docker.io` due to the conservative nature of Ubuntu packagers. `apt-get` will install version 1.2, and the Eris stack works well on Docker version >= 1.4. 

Ubuntu:

```bash
curl -sSL https://get.docker.com/ubuntu/ | sudo sh
```

Other Linux:

Please see the [Docker documentation](https://docs.docker.com/installation/) for your platform.

OSX && Windows:

You will need `boot2docker` installed. This tool uses a very light kernel and VirtualBox to provide the base framework which Docker needs. boot2docker then automates docker commands. 

Boot2Docker can be installed from [here](https://github.com/boot2docker/osx-installer/releases/latest)

## Containers

Containers are what are ran. They are *used* by docker. 

You can list running containers with `docker ps` (all containers with `docker ps -a`).

You can remove containers with `docker rm CONTAINER`.

## Images

Images are what Dockerfiles build; they are also what `docker pull` and `docker push` work with. Images become containers when they are used. 

You can list images with `docker images` and remove images with `docker rmi IMAGE`. 

## Dockerfile

The Dockerfile determines how an image will be built. 

A nice checklist for security of Dockerfiles is available [here](http://linux-audit.com/security-best-practices-for-building-docker-images/)

## Volumes

Docker volumes are used for persistence. So they store data from the container on the host and can be re-used. docker creates folders on the host in /var/lib/docker/vfs (or similar) that store the exported volumes. If you're not careful, you may end up creating hundreds of these. [This](https://github.com/cpuguy83/docker-volumes) tool is super helpful for managing docker volumes. Highly recommended.

Docker also let's you *mount* a directory from the host, by mapping it to a volume in the container. The problem here is you can *only* mount directories as root, so if you are running as non-root in the container, this doesn't help. Instead of mounting a volume from the host, you should use a *data-only container*. This means you run an instance of your container that simply exits (for the sake of your logs, do something like `docker run --name myapp-data myapp echo "Data-only container for myapp"`). Now when you run your app container, instead of mounting a volume from the host, do `docker run --name myapp --volumes-from myapp-data -d myapp`. When `myapp` containers writes anything to the exported volume, it will be saved in the `myapp-data` container. Don't delete this container, or you'll delete the data (which again is in /var/lib/docker/<something>. But it doesn't matter where it is, use `docker exec --volumes-from myapp-data myapp ls /path/to/volume` to inspect it, or `docker inspect myapp` for more info).

Finally, you may want to copy data from a directory on the host into your container. This is typically where you might mount the directory, with `docker run -v host_dir:image_dir myapp`. Again, it will only mount as root. Instead, copy it into the data only container using `tar`:

```
tar cf - . | docker run -i --volumes-from myapp-data myapp tar xvf - -C /data/myapp
```

This should become easier once the `docker cp` command allows you to copy from host to container (there's an open PR).

Finally, note that docker-compose seems to have some trouble with volumes-from and the above expects you to use shell scripts instead. See for example https://github.com/tendermint/tendermint/tree/godep/DOCKER

## Build

Remember the `--no-cache` flag. Use it when needed.

## Its Like Git

Docker doesn't bring in any changes unless you tell it to bring in changes; like git. So if an image exists locally, say `eris/decerver` (soon to be `eris/erisserver`) or `eris/erisdb` then if you build on top of that image (with fig or whatever) the image will not be updated even if later changes are sent to docker hub. This is for stability and efficiency reasons. To bring in new changes from the docker hub just `docker pull eris/decerver` or whatever. 

## Tags

Tags are great. They can work both like channels (such as a `:latest` tag or a `:stable` tag or whatever) and with semantic versioning (such as a `:1.4` tag like our golang images use). Use them!

## Helpful Tips

```bash
docker_rmv () {
        docker-volumes rm $(docker-volumes list -q)
}
```

Clears **all** the volumes. Handle with care. 

```bash
#docker remove all untagged images
function docker_rmun()
{
	docker rmi $(docker images | grep "^<none>" | awk '{print $3; }')
}
function docker_rma()
{
	docker rm -v -f $(docker ps -a -q)
}
function docker_rmv()
{
	docker-volumes rm $(docker-volumes list -q)
}
function docker_rmiall()
{
	docker rmi $(docker images -q)
}

#docker-machine functions
function dmx()
{
	dm_clear_current
}
function dm_which()
{
	mach_host=$(echo $DOCKER_HOST | sed -r 's/tcp:\/\///' | sed -r 's/:2376//')
  mach_name=$(basename $DOCKER_CERT_PATH)
	echo "$mach_name":"$mach_host"
}
function dm_clear_current()
{
	unset DOCKER_TLS_VERIFY
	unset DOCKER_HOST
	unset DOCKER_CERT_PATH
	unset DOCKER_MACHINE_NAME
}

#docker-machine eris
function dme_use()
{
	ping_times=0
	machine="eris"
	if [ "$#" -eq 0 ]; then dm_clear_current; return 0; fi
	for x; do; machine="$machine-$x";	done
  docker-machine start $machine &> /dev/null
  until [[ $(docker-machine status $machine) == "Running" ]] || [ $ping_times -eq 5 ]
  do
     ping_times=$[$ping_times +1]
     sleep 3
  done
  if [[ $(docker-machine status $machine) != "Running" ]]
  then
    echo "Could not start the machine."
    return 1
  fi
  echo "Machine Started."
	echo "Connecting to machine => $machine"
  eval "$(docker-machine env $machine)" &>/dev/null
  echo "Connected to Machine."
}

function dme_clear()
{
	if [ "$#" -eq 0 ]
	then 
		cur=$(docker-machine active) 2>/dev/null
	else
		cur="eris"
	  for x; do cur="$cur-$x"; done
	fi
  docker-machine kill $cur 2>/dev/null
	dm_clear_current
}

#docker-machine erispaas
function dmp_use()
{
	mach="erispaas"
	if [ "$#" -eq 0 ]
	then 
		dm_clear_current
		return 0
	elif [ "$#" -eq 2 ]
	then
		for x; do; mach="$mach-$x";	done
		mach="$mach-00"
	  echo "Connecting to machine => $mach"
		eval $(docker-machine env --swarm $mach)
	else
	  for x; do; mach="$mach-$x";	done
	  echo "Connecting to machine => $mach"
	  eval $(docker-machine env $mach)
	fi
}

function dmp_clear()
{
	if [ "$#" -eq 0 ]
	then 
		cur=$(docker-machine active)
	else
		cur="erispaas"
	  for x; do cur="$cur-$x"; done
	fi
  docker-machine kill $cur
	dm_clear_current
}

function dmp_wipe()
{
	mach="erispaas"
	for x; do; mach="$mach-$x";	done
	fin=4
	for ((i=0;i<=fin;i++))
	do
		my_machine="$mach"-0$i
		echo "Removing machine => $my_machine"
		docker-machine rm $my_machine 1>/dev/null
	done
}
```

Clears **all** the **untagged** containers. Handle with care. 

# Good Links:

http://www.carlboettiger.info/2014/08/29/docker-notes.html - tips on efficiency and workflow

http://about.archilogic.com/index.html@p=675.html - tips on CI and CD with etcd

# Docker-Machine

DM is **wonderful**

Some functions to get this section started

```bash
function dm_clear_current()
{
        unset DOCKER_TLS_VERIFY
        unset DOCKER_HOST
        unset DOCKER_CERT_PATH
        unset DOCKER_MACHINE_NAME
}
function dm_use()
{
        if [ "$#" -eq 0 ]; then dm_clear_current; return 0; fi
        mach="eris"
        for x; do mach="$mach-$x"; done
        docker-machine start $mach
        eval $(docker-machine env $mach)
}
function dm_clear()
{
        if [ "$#" -eq 0 ]
        then 
                cur=$(docker-machine active)
        else
                cur="eris"
          for x; do cur="$cur-$x"; done
        fi
        docker-machine kill $cur
        dm_clear_current
}
```