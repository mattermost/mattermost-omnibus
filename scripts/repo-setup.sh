#!/bin/bash
set -euo pipefail

# receives a gpg key and returns its fingerprint
getFingerprint() {
    gpg --with-colons --import-options show-only --import | sed -n 2p | cut -d ':' -f 10
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

if [[ "$release" != "bionic" && "$release" != "focal" ]]; then
    printf "ERROR: Unsupported ubuntu release: \"%s\"\n" "$release" >&2
    exit 1
fi

# check root or sudo usage

# Nginx
apt-key adv --keyserver keyserver.ubuntu.com --recv-keys ABF5BD827BD9BF62
nginxFingerprint=$(apt-key export ABF5BD827BD9BF62 2>/dev/null | getFingerprint)
validateNginxKey "$nginxFingerprint"
add-apt-repository -y "deb https://nginx.org/packages/ubuntu/ ${release} nginx"

# Certbot
case "$release" in
    bionic)
        add-apt-repository -y ppa:certbot/certbot
        ;;
    focal)
        add-apt-repository universe
        ;;
esac

# PostgreSQL
pgKey=$(mktemp)
curl https://www.postgresql.org/media/keys/ACCC4CF8.asc -o "$pgKey"
pgFingerprint=$(getFingerprintFromFile "$pgKey")
validateAndAddPgKey "$pgFingerprint" "$pgKey"
add-apt-repository -y "deb http://apt.postgresql.org/pub/repos/apt ${release}-pgdg main"

# Mattermost Omnibus
mmKey=$(mktemp)
curl https://deb.packages.mattermost.com/pubkey.gpg -o "$mmKey"
mmFingerprint=$(getFingerprintFromFile "$mmKey")
validateAndAddMmKey "$mmFingerprint" "$mmKey"
apt-add-repository -y "deb https://deb.packages.mattermost.com ${release} main"

# Update to retrieve all the newly added repositories.
apt update
