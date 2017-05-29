---

layout: single
title: "Tutorials | Building From Source"
aliases:
  - /docs/install-source
menu:
  tutorials:
    weight: 5

---

<div class="note">
{{% data_sites rename_docs %}}
</div>

You will need `go` and `git` installed to do this.

## Install Go

At the current time, `monax` requires `go` version >= {{< data_coding golang_min >}}. Go is not needed if you install `monax` as a binary.

An easy way to install Go (for macOS and Linux) is via the Travis-CI's [Gimme](https://github.com/travis-ci/gimme) tool. First you install Gimme; then running the `eval $(gimme {{< data_coding golang_auth >}})` command and you'll be all set up.

Otherwise, please see the documentation in [this link](https://golang.org/doc/install) on how to install it.

Make sure that Go is properly installed by running:

```
go version
```

Once you have Go installed, you will then want to check if you also have your `$GOBIN` value in your `$PATH`. Most gophers add the following lines to their `~/.bashrc`, `~/.profile`, `~/.zshrc` file or other relevant file.

```
export GOPATH=$HOME
export GOBIN=$GOPATH/bin
export PATH="$GOBIN:$PATH"
```

**Note** you will need to double check that you perform the above commands for the *user* which will be running `monax`.

If you do not add those lines to the relevant shell files then you can just type them into the shell each time you log in. You can check that this change was added by running the `echo $PATH|tr ':' '\n'` command and making sure that your path has been updated appropriately.

Now you're ready to install the components of the Monax platform.

## Building Monax from source

Go makes it very easy to build from source. Indeed, it is really only one command.

```
go get github.com/monax/monax/cmd/monax
monax init
```

Now you're ready to go (if you pardon the pun)!


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Tutorials](/docs/)
