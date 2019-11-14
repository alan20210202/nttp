#!/bin/sh

echo Installing NTTP...

if ! [ "$(id -u)" = 0 ]; then
  echo Please run this script as root!
  exit 1
fi

if ! [ -x "$(command -v go)" ]; then
  echo Cannot detect golang, trying to install...
  if [ -x "$(command -v apt-get)" ]; then
    add-apt-repository ppa:longsleep/golang-backports
    apt-get update
    apt-get install -y golang-go
  fi
  if ! [ -x "$(command -v go)" ]; then
    echo Cannot detect apt-get or installation failed, please install go manually!
    exit 1
  fi
fi

gopath="$(go env GOPATH)"
nttp_root="$gopath/src/nttp"
src_path="$nttp_root/main/main.go"
bin_path="$nttp_root/bin/nttp"
install_path="/usr/local/bin/nttp"

echo Detected GOPATH: "$gopath"
mkdir -p "$gopath/src"
rm -rf "$nttp_root"

echo Copying source files...
cp -r . "$nttp_root"

echo Fetching dependencies...
go get github.com/urfave/cli

echo Building...
go build -o "$bin_path" "$src_path"

echo Creating symbolic link...
rm $install_path
ln -s "$bin_path" "$install_path"

# shellcheck disable=SC2039
read -r -p "Register systemd service? [Y/n] " yn
yn="${yn:-Y}"
case $yn in
  [Yy]* )
    # shellcheck disable=SC2039
    read -r -p "Listen on: [:44353]" listen
    listen="${listen:-:44353}"
    # shellcheck disable=SC2039
    read -r -p "Public address: [0.0.0.0]" pubaddr
    pubaddr="${pubaddr:-0.0.0.0}"

    echo Listening on "$listen" with public address "$pubaddr"
    mkdir -p "/usr/lib/systemd/system"
    service_path="/usr/lib/systemd/system/nttp.service"
    echo "[Unit]" > "$service_path"
    echo "Description=NTTP Server Service" >> "$service_path"
    echo >> "$service_path"
    echo "[Service]" >> "$service_path"
    echo "ExecStart=$install_path server -l $listen -s $pubaddr" >> "$service_path"
    echo >> "$service_path"
    echo "[Install]" >> "$service_path"
    echo "WantedBy=multi-user.target" >> "$service_path"
    ;;
  [Nn]* )
    echo Cancelled systemd service registration!
    ;;
esac

echo Done!