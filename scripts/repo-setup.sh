#!/bin/bash
set -euo pipefail

if [[ ! $# -eq 0 ]]; then
    ARGUMENT_1="$1"
else
    ARGUMENT_1="all"
fi

# receives a gpg key and returns its fingerprint
getFingerprint() {
    gpg --with-colons --import-options show-only --import 2>/dev/null | sed -n 2p | cut -d ':' -f 10
}

getFingerprintFromFile() {
    file=$1
    cat "$file" | getFingerprint
}

validateNginxKey() {
    nginxFingerprint=$1
    # source: http://nginx.org/en/linux_packages.html
    expectedNginxFingerprint=573BFD6B3D8FBC641079A6ABABF5BD827BD9BF62

    if [[ "$nginxFingerprint" != "$expectedNginxFingerprint" ]]; then
        printf "ERROR: Invalid nginx key fingerprint\n" >&2
        apt-key del ABF5BD827BD9BF62
        exit 1
    fi
}

validatePgKey() {
    pgFingerprint=$1
    # source: https://wiki.postgresql.org/wiki/Apt
    expectedPgFingerprint=B97B0AFCAA1A47F044F244A07FCC7D46ACCC4CF8

    if [[ "$pgFingerprint" != "$expectedPgFingerprint" ]]; then
        printf "ERROR: Invalid postgresql key fingerprint\n" >&2
        apt-key del 7FCC7D46ACCC4CF8
        exit 1
    fi
}

validateAndAddPgKey() {
    pgFingerprint=$1
    pgKey=$2
    # source: https://www.postgresql.org/about/news/pgdg-apt-repository-for-debianubuntu-1432/
    expectedPgFingerprint=B97B0AFCAA1A47F044F244A07FCC7D46ACCC4CF8

    if [[ "$pgFingerprint" != "$expectedPgFingerprint" ]]; then
        printf "ERROR: Invalid PostgreSQL key fingerprint\n" >&2
        rm "$pgKey"
        exit 1
    fi

    cat "$pgKey" | sudo apt-key add -
    rm "$pgKey"
}

validateMmKey() {
    mmFingerprint=$1
    expectedMmFingerprint=A1B31D46F0F3A10B02CF2D44F8F2C31744774B28

    if [[ "$mmFingerprint" != "$expectedMmFingerprint" ]]; then
        printf "ERROR: Invalid mattermost key fingerprint\n" >&2
        apt-key del F8F2C31744774B28
        exit 1
    fi
}

validateAndAddMmKey() {
    mmFingerprint=$1
    mmKey=$2
    expectedMmFingerprint=A1B31D46F0F3A10B02CF2D44F8F2C31744774B28

    if [[ "$mmFingerprint" != "$expectedMmFingerprint" ]]; then
        printf "ERROR: Invalid Mattermost key fingerprint\n" >&2
        rm "$mmKey"
        exit 1
    fi

    cat "$mmKey" | apt-key add -
    rm "$mmKey"
}

########################################################
# Repo setup script for Mattermost Omnibus installation.
########################################################

release=$(lsb_release -cs)
architecture=$(dpkg-architecture -q DEB_HOST_ARCH)
curl_binary=$(which curl)

if [[ "$release" != "focal" && "$release" != "jammy" ]]; then
    printf "ERROR: Unsupported ubuntu release: \"%s\"\n" "$release" >&2
    exit 1
fi

# check root or sudo usage
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit
fi


# Install Nginx,Certbot,PostgreSQL repositories in case ARGUMENT_1 == all
if [[ $ARGUMENT_1 == "all" ]]; then

    case "$release" in
        focal )
            # Nginx
            apt-key adv --keyserver keyserver.ubuntu.com --recv-keys ABF5BD827BD9BF62
            nginxFingerprint=$(apt-key export ABF5BD827BD9BF62 2>/dev/null | getFingerprint)
            validateNginxKey "$nginxFingerprint"
            add-apt-repository -y "deb [arch=$architecture] https://nginx.org/packages/ubuntu/ ${release} nginx"
            # PostgreSQL
            pgKey=$(mktemp)
            "$curl_binary" -s https://www.postgresql.org/media/keys/ACCC4CF8.asc -o "$pgKey"
            pgFingerprint=$(getFingerprintFromFile "$pgKey")
            validateAndAddPgKey "$pgFingerprint" "$pgKey"
            add-apt-repository -y "deb [arch=$architecture] http://apt.postgresql.org/pub/repos/apt ${release}-pgdg main"
            ;;
        jammy)
            # Nginx
            "$curl_binary" -s https://nginx.org/keys/nginx_signing.key | gpg --dearmor \
            | sudo tee /usr/share/keyrings/nginx-archive-keyring.gpg >/dev/null
            nginxFingerprint=$(getFingerprintFromFile "/usr/share/keyrings/nginx-archive-keyring.gpg")
            validateNginxKey "$nginxFingerprint"
            echo "deb [signed-by=/usr/share/keyrings/nginx-archive-keyring.gpg] \
            http://nginx.org/packages/ubuntu ${release} nginx" | tee /etc/apt/sources.list.d/nginx.list  &>/dev/null
            # PostgreSQL
            "$curl_binary" -s https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor \
            | sudo tee /usr/share/keyrings/postgresql-archive-keyring.gpg >/dev/null
            pgFingerprint=$(getFingerprintFromFile "/usr/share/keyrings/postgresql-archive-keyring.gpg")
            validatePgKey "$pgFingerprint"
            echo "deb [signed-by=/usr/share/keyrings/postgresql-archive-keyring.gpg] \
            http://apt.postgresql.org/pub/repos/apt ${release}-pgdg main" | tee /etc/apt/sources.list.d/pgdg.list  &>/dev/null
            ;;
    esac
    # Certbot
    case "$release" in
        focal | jammy )
            add-apt-repository -y universe
            ;;
    esac
fi

if [[ $ARGUMENT_1 == "all" || $ARGUMENT_1 == "mattermost" ]] ; then
    case "$release" in
        focal )
            # Mattermost Omnibus
            mmKey=$(mktemp)
            "$curl_binary" -s https://deb.packages.mattermost.com/pubkey.gpg -o "$mmKey"
            mmFingerprint=$(getFingerprintFromFile "$mmKey")
            validateAndAddMmKey "$mmFingerprint" "$mmKey"
            apt-add-repository -y "deb [arch=$architecture] https://deb.packages.mattermost.com ${release} main"
            ;;
        jammy)
            # Mattermost Omnibus
            "$curl_binary" -s  https://deb.packages.mattermost.com/pubkey.gpg | gpg --dearmor \
            | sudo tee /usr/share/keyrings/mattermost-archive-keyring.gpg >/dev/null
            mmFingerprint=$(getFingerprintFromFile "/usr/share/keyrings/mattermost-archive-keyring.gpg")
            validateMmKey "$mmFingerprint"
            echo "deb [signed-by=/usr/share/keyrings/mattermost-archive-keyring.gpg] \
            https://deb.packages.mattermost.com ${release} main" | tee /etc/apt/sources.list.d/mattermost.list  &>/dev/null
            ;;
    esac
fi
# Update to retrieve the newly added repositories.
apt update
