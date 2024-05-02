# Mattermost Omnibus


Generates Mattermost debian packages easy that makes it trivial to
deploy the platform in an empty instance.

## Installation

Please refer to the [Omnibus Install
Documentation](https://docs.mattermost.com/install/mattermost-omnibus.html).

## Development environment

### Requirements

You need to have `make` and `docker` installed, and the following
repositories added to the apt sources:

 - [NGINX repository](https://www.nginx.com/resources/wiki/start/topics/tutorials/install/#official-debian-ubuntu-packages)
 - [Certbot repository](https://certbot.eff.org/lets-encrypt/ubuntubionic-nginx.html)
 - [PostgreSQL repository](https://wiki.postgresql.org/wiki/Apt)

### Usage

To build a package, just run:

```sh
make all version=X.Y.Z revision=N
```
