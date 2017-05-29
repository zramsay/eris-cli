---

layout: single
title: "Tutorials | Trouble Shooting Your Installation"
aliases:
  - /docs/install-troubleshooting
menu:
  tutorials:
    weight: 5

---

<div class="note">
{{% data_sites rename_docs %}}
</div>

## I'm On macOS or Windows and It's All Wonky!

Never fear, the marmots are here. See [Section 1 of our docker-machine tutorial](/docs/) and come back to your installation. All will be well.

## No `monax` Command Found

If you get a "monax: command not found" error then (if you built it from source) you need to make sure that your `$GOBIN` variable value is in your `$PATH` (see [Getting Started](/docs/getting-started/) and then do:

```irc
cd $GOPATH/src/github.com/monax/monax/cmd/monax
go install
cd ~1
monax init
```

If you received that error but you performed the binary installation, then you will need to make sure that the zip or tarball which was extracted from the [Github Releases](https://github.com/monax/monax/releases) page was installed into a place in your `$PATH` which the shell can use. Please see the documentation for your operating system, or ask the Google for help.

## No Output At All

If you type `monax init` or `monax init --debug` and you get **no** output, this is almost always because your current user is not added to the docker group. To fix:

```bash
sudo usermod -a -G docker $USER
```

From the user who will be using Monax.

You will need to close the terminal window and open a new terminal for the changes to take effect. If you are `ssh`-ing into a cloud based development machine, then log out and log back in so that the changes will take effect.

Double check that your changes have taken hold (after you log back in or in a new terminal window) by:

```bash
groups $USER
```

From the user who will be using Monax.

Confirm that the line output includes `docker` and you will be good to go!

## Can't See What's Happening?

By default, `monax` is a fairly quiet tool. If you would like to have more output you can add `-v` (for verbose) **or** `-d` (for debug) to any command in order to see more output. In general, there is no need to use *both* of these flags. The `--verbose` flag will give a bit more output than the command will by default and the `--debug` flag will give *much* more output than the either the `--verbose` flag or the command by default, but will be directed primarily at Monax developers.

If you are reporting a bug, please rerun the command which caused the issue with the debug flag (`-d` or `--debug`) and send us the output to a [Github Issue](https://github.com/monax/monax/issues/new) or via Community Driven [Support Forums](https://support.monax.io).

## I'm Behind a Firewall

Docker itself needs to be given a proxy. You can do this by updating your `/etc/default/docker` with the following line.

```
export http_proxy="http://myproxy.com"
```

Some organizations do not allow connections to [quay.io](https://quay.io) (which is an equivalent to [Docker Hub](https://hub.docker.com) but with many more security features). At Monax we keep most of our Docker images on quay.io. At this time we are working on an automatic bridge script which will mirror our quay.io images to [Docker Hub](https://hub.docker.com) but that is still a work in progress. As such you will likely need to ensure your computer can access [quay.io](https://quay.io).


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Tutorials](/docs/)
