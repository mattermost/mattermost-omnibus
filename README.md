# Mattermost Omnibus

> **⚠️ DEPRECATION NOTICE**  
> This project has been deprecated starting with Mattermost v11. The last release of mattermost-omnibus is v10.12. Please consider migrating to alternative deployment methods such as Docker, Kubernetes, or the official deployment guide available at [docs.mattermost.com/deployment-guide/server/server-deployment-planning.html](https://docs.mattermost.com/deployment-guide/server/server-deployment-planning.html).

## Migration Guide

If you're currently using mattermost-omnibus and need to migrate to an officially supported deployment method, follow these steps:

### 1. Backup Your Data

Before migrating, create a complete backup using the built-in backup command:

```bash
# Create a complete backup (includes config, database, and data directory)
sudo mmomni backup
```

This will create a tarball containing:
- Configuration file (`mmomni.yml`) - contains omnibus-specific settings
- PostgreSQL database dump  
- Data directory with file uploads
- All necessary files for migration

**Important:** The `mmomni.yml` file contains omnibus-specific configuration that you'll need to translate to your new deployment:
- `db_user` & `db_password` - Database credentials for accessing your existing PostgreSQL instance
- `fqdn` - Your server's domain name
- `email` - Admin email for SSL certificates
- `https` - SSL/TLS configuration
- `data_directory` - Path to file uploads
- `enable_plugin_uploads` - Plugin upload settings
- `enable_local_mode` - Local development mode settings
- `client_max_body_size` - File upload size limits

These settings will help you configure your new deployment's database connection, SSL certificates, nginx settings, and Mattermost server configuration.

Also back up your server configuration files located at `/opt/mattermost/config`.

### 2. Choose Your New Deployment Method

Select one of the officially supported deployment methods:

- **Docker**: Recommended for single-server deployments
- **Kubernetes**: Best for scalable, cloud-native deployments  
- **Binary Installation**: Direct server installation with manual configuration

Refer to the [deployment planning guide](https://docs.mattermost.com/deployment-guide/server/server-deployment-planning.html) to choose the best option for your needs.

### 3. Install New Deployment

Follow the installation guide for your chosen method:

- [Docker deployment](https://docs.mattermost.com/install/install-docker.html)
- [Kubernetes deployment](https://docs.mattermost.com/install/install-kubernetes.html)
- [Binary installation](https://docs.mattermost.com/install/install-ubuntu.html)

### 4. Restore Your Data

After setting up the new deployment:

1. Stop the new Mattermost service
2. Extract the backup tarball created by `mmomni backup`
3. Use the database credentials from `mmomni.yml` (`db_user`, `db_password`) to restore the PostgreSQL database dump to your new database instance
4. Configure your new deployment using settings from `mmomni.yml`:
   - Update database connection settings with the restored database
   - Configure the site URL using the `fqdn` value
   - Set up SSL/TLS certificates (use `email` and `https` settings as reference)
   - Configure nginx with appropriate `client_max_body_size`
   - Enable plugin uploads based on `enable_plugin_uploads` setting
5. Restore file uploads from the backup's data directory to your new deployment's data directory (referenced by `data_directory` in `mmomni.yml`)
6. Start the Mattermost service

### 5. Verify Migration

- Test user login and functionality
- Verify file uploads and downloads work correctly
- Check that integrations and plugins are functioning
- Update any external integrations to point to the new server

### 6. Remove Old Installation

Once you've verified the migration was successful:

```bash
# Remove omnibus package
sudo apt remove -y mattermost-omnibus --purge
sudo apt autoremove

# Clean up and check old files are removed (be careful - ensure backups are safe first)
sudo rm -rf /opt/mattermost
sudo rm -rf /var/opt/mattermost
```

For assistance with migration, consult the [Mattermost community forums](https://forum.mattermost.com/) or [official documentation](https://docs.mattermost.com/).

---

# Original Project Documentation

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
