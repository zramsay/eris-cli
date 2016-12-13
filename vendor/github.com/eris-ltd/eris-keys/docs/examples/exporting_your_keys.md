Unless you have a different configuration than our standard, admittedly opinionated, pathway then you will be running the eris-keys signing server inside of a container. This means that you need to be able to import and export your keys. This tutorial covers the existing `eris keys` commands and working with keys vis-a-vis containers on the eris platform.

**Note** -- This is a reminder to treat Docker containers as ephemeral. Even with data containers, you do not want to keep important things inside a container without backups. Just as you would not keep any other important files in one location without a backup.

# Start keys container

If you don't have one already running...

```bash
eris services start keys
```

# Generate a key

```bash
eris keys gen
```

That will create a **non safe** (but easy for development) key. You'll see an address output that looks like: `ECD053462FF7B4B6C5003AA5E3549C3A93DB41E6`. Since the key is in the keys data container, you'll need to export it to host. See it in the container with:

```bash
eris keys ls --container
```

The same address as above should be output.

# Exporting a key

**Note:** If you have many keys to export (or import), the `--all` flag is now available for both commands.

Exporting your key to the host is quite easy:

```bash
eris keys export ADDR
```

What this does is take the contents (a key) from `/home/eris/.eris/keys/data/ADDR` in the container and copies it to `~/.eris/keys/data/ADDR`. This is the simplest way to backup your key to the default keys path. Check that it is there with:

```bash
eris keys ls --host
```
or

```bash
ls ~/.eris/keys/data
```
That's it. Now it's in the right position on your host. Opptionally, to export all keys from the keys container you can use the data command:

```bash
eris data export keys /home/eris/.eris/keys/data ~/.eris/keys
```

To import a key we do the reverse.

# Importing a key

```bash
eris keys import ADDR
```

This command will take `~/.eris/keys/data/ADDR` from the host and copy it to `/home/eris/.eris/keys/data/ADDR` in an existing container. It is useful to "loading" backed-up keys into a container that will be used in deployement. Note: `--all` can be used here as well.

## Get pubkey

Returns a pubkey; used for making genesis files.

```bash
eris keys pub ADDR
```

## Convert key to tendermint format

This command will soon be deprecated in favour of adding a pubkey to `config.toml` rather than loading a `priv_validator.json` on `eris chains new`. In the meantime, it takes an `eris-keys` format key and converts it to a tendermint format priv validator.

```bash
eris keys convert ADDR
```

The latter two commands require `ADDR` in the running keys container.

# Under the hood
All the above commands wrap either `eris services exec keys "CMD"` or `eris data import/export keys SRC DEST`, which executes the necessary commands within a docker container. The following is a comparison.

```bash
eris services exec keys "eris-keys gen --no-pass"
eris services exec keys "ls /home/eris/.eris/keys/data"
eris data export keys /home/eris/.eris/keys/data/ADDR ~/.eris/keys/data/ADDR
eris data import keys ~/.eris/keys/data/ADDR /home/eris/.eris/keys/data/ADDR
eris services exec keys "eris-keys pub --addr ADDR"
eris services exec keys "mintkey mint ADDR"
```
compared to:

```bash
eris keys gen
eris keys ls --container
eris keys export ADDR
eris keys import ADDR
eris keys pub ADDR
eris keys convert ADDR
```

Yay Docker.
