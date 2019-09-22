
# Gelf Multi Docker logging plugin
This Docker plugin ships container logs to multiple gelf endpoints. These instructions are for Linux host systems. For other platforms, see the [Docker Engine managed plugin system documentation](https://docs.docker.com/engine/extend/).

# Prerequisites
Here's what you need to run the plugin:

* Docker Engine version 17.05 or later. If you plan to configure this plugin using `daemon.json`, you need Docker Community Edition (Docker-ce) 18.03 or later.

# Install and configure the Gelf Multi Docker logging plugin

## Step 1: Install the plugin
Choose how you want to install the plugin:
* [Option 1: Install from the Docker Store](#option-1-install-from-the-docker-store)
* [Option 2: Install from source](#option-2-install-from-source)


### Option 1: Install from the Docker Store

1. Pull the plugin from the Docker Store:
  ```
  $ docker plugin install stefansaftic/gelf-multi:<version> --alias gelf-multi
  ```

2. Enable the plugin, if needed:
  ```
  $ docker plugin enable gelf-multi
  ```

Continue to [Step 2: Set configuration variables](#step-2-set-configuration-variables)

### Option 2: Install from source

1. Clone the repository and check out release branch:
  ```
  $ cd docker-gelf-multi-log-driver
  $ git checkout release
  ```

2. Build the plugin:
  ```
  $ make all
  ```

3. Enable the plugin:
  ```
  $ docker plugin gelf-multi:<version> enable
  ```

4. Restart the docker daemon for the changes to apply:
  ```
  $ service docker restart
  ```

Continue to [Step 2: Set configuration variables](#step-2-set-configuration-variables)

## Step 2: Set configuration variables
Choose how you want to configure the plugin parameters:
* [Option 1: Configure all containers with daemon.json](#option-1-configure-all-containers-with-daemonjson)
* [Option 2: Configure individual containers at run time](#option-2-configure-individual-containers-at-run-time)


### Option 1: Configure all containers with daemon.json
The `daemon.json` file allows you to configure all containers with the same options.

For example:
```
{
  "log-driver": "gelf-multi:<version>",
  "log-opts": {
    "gelf-count": "<count of gelf loggers>"
    "gelf-multi-gelf-address.0": "<gelf udp or tcp address >"
  }
}
```

1. _(Optional)_ Set any [environment variables](#advanced-options-environment-variables)
2. Include all [required variables](#required-variables) in your configuration and any [optional variables](#optional-variables).

Continue to [Step 3: Run containers](#step-3-run-containers)

### Option 2: Configure individual containers at run time

Configure the plugin separately for each container when using the docker run command. For example:
```
$ docker run --log-driver=gelf-multi:<version> --log-opt gelf-count=1 --log-opt gelf-address.0=<gelf udp or tcp address> <your_image>
```

1. _(Optional)_ Set any [environment variables](#advanced-options-environment-variables)
2. Include all [required variables](#required-variables) in your configuration and any [optional variables](#optional-variables).

Continue to [Step 3: Run containers](#step-3-run-containers)


#### Required Variables

| Variable | Description | Notes |
| --- | --- | --- |
| `gelf-count` | Count of gelf loggers. | |
| `gelf-multi-gelf-address.X` | Gelf address for gelf logger. | X(0..gelf-count) |

#### Optional Variables

| Variable | Description | Default value |
|---|---|---|
| `gelf-multi-gelf-compression-type.X` | Gelf compression type. | |
| `gelf-multi-gelf-compression-level.X` | Gelf compression level. | |
| `gelf-multi-gelf-tcp-max-reconnect.X` | Gelf max tcp reconnect. | |
| `gelf-multi-gelf-tcp-reconnect-delay.X` | Gelf tcp reconnect delay. | |
| `gelf-multi-tag.X` | Gelf tag. | |
| `gelf-multi-labels.X` | Gelf labels. | |
| `gelf-multi-labels-regex.X` | Gelf labels regex. | |
| `gelf-multi-env.X` | Gelf env. | |
| `gelf-multi-env-regex.X` | Gelf env-regex. | |
| `json-multi-max-file` | Json max-file. | |
| `json-multi-max-size` | Json max-size of the file. | |
| `json-multi-compress` | Json compress. | |
| `json-multi-tag` | Json tag. | |
| `json-multi-labels` | Json labels. | |
| `json-multi-labels-regex` | Json labels regex. | |
| `json-multi-env` | Json env. | |
| `json-multi-env-regex` | Json env-regex. | |

### Usage example

```
$ docker run --log-driver=gelf-multi:<version> \
             --log-opt gelf-count=2
             --log-opt gelf-multi-gelf-address.0=udp://127.0.0.1:12201 \
             --log-opt gelf-multi-gelf-address.1=,udp://127.0.0.1:12202 \
             --env "DEV=true" \
             --label region=us-east-1 \
             <docker_image>
```

## Step 3: Run containers

Now that the plugin is installed and configured, it will send the container while the container is running.

To run your containers, see [Docker Documentation](https://docs.docker.com/config/containers/logging/configure/).

## Credits
This plugin relies on the open source docker gelf logging driver and json-file logging driver

## Release Notes
- 1.0.0 - First version.