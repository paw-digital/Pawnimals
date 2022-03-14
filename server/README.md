# natricon backend
[![CI](https://github.com/appditto/natricon/workflows/CI/badge.svg)](https://github.com/appditto/natricon/actions?query=workflow%3ACI) [![Twitter Follow](https://img.shields.io/twitter/follow/appditto?style=social)](https://twitter.com/intent/follow?screen_name=appditto)

The backend and business logic for [natricon](https://natricon.com)

natricon is built in [GOLang](http://golang.org/)

## Requirements

The natricon backend requires ImageMagick development libraries to be installed. ImageMagick should be compiled with librsvg, libxml2, libpng, and libwebp.

## Installing pre requisites on Ubuntu 20.04
### Installing GO
```bash
curl -OL https://golang.org/dl/go1.16.7.linux-amd64.tar.gz
sha256sum go1.16.7.linux-amd64.tar.gz
sudo tar -C /usr/local -xvf go1.16.7.linux-amd64.tar.gz

# edit profile
sudo nano ~/.profile
# add this to the end of the file
export PATH=$PATH:/usr/local/go/bin

# run this line
source ~/.profile
```

### Installing Redis
```bash
sudo apt install redis-server
```


### Install ImageMagick

```bash
sudo apt-get install build-essential
sudo apt install libpng16-16
sudo apt install librsvg2-dev
sudo apt install libxml2-dev
sudo apt install libwebp-dev
wget https://download.imagemagick.org/ImageMagick/download/ImageMagick.tar.gz
tar -axvf ImageMagick.tar.gz
cd ImageMagick-7.1.0-13/

./configure --with-rsvg=yes --with-xml=yes --with-png=yes --with-webp=yes
# the ouput should show all flags with yes
#--with-xml=yes              yes
#--with-png=yes              yes
#--with-webp=yes             yes
#--with-rsvg=yes             yes
make
make install
sudo ldconfig
# convert has to and should output png rsvg webp and xml
convert --version
```

## Natricon server build setup

```bash
# install dependencies
$ go get -u
# run in debugging mode
$ go run .

# build binary for production
$ go build . -o natricon
# execute natricon in production mode
$ GIN_MODE=release ./natricon

# For all options run
$ ./natricon -help
```

## WebAssembly (wasm) build setup

There is a WebAssembly reference implementation in the [wasm folder](https://github.com/appditto/natricon/tree/master/server/wasm)

This allows you to generate a natricon entirely on client-side from within the browser.

```bash
# To compile wasm
$ cd wasm
$ GOOS=js GOARCH=wasm go build -o main.wasm

# Running the sample
$ go get -u github.com/go-serve/goserve
$ ${GO_PATH}/bin/serve .
```

## Other Server Configuration

The server implements some mechanisms for tracking donations, such as a socket.io server and subscription to nano node websocket.

The account to listen for donations can be set in the environment

```
export DONATION_ACCOUNT=nano_1natrium1o3z5519ifou7xii8crpxpk8y65qmkih8e8bpsjri651oza8imdd
```

The websocket can be setup by specifying it as a parameter

```
./natricon -node-ws-url http://[::1]:7078
```

The website uses raw amounts to identify donations with specific clients and implements an automatic refund mechanism, to use this the WALLET_ID should be specified in the environment. For either Pippin or the Nano Node Wallet

```
export WALLET_ID=d897b5ec-1897-4e7e-8a90-4526f454c8de
```

All of these settings are optional, and don't need to be specified for the natricon server to run.
