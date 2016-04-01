# varnishadm-ninja

Basic command line tool to communicate with Varnish Administration/Management interface over TCP

Based on [GVA](https://github.com/kreuzwerker/gva)

## Use

This utility could be useful for next cases when you need to communicate with Varnish but:

* not able to install Varnish package to obtain `varnishadm` tool
* not able to read Varnish secret file on file-system, but know it content

Almost of all command works as expected, but different Varnish versions has different syntax and number of parameters, so check documentation before:

* [Varnish CLI 3.x](https://www.varnish-cache.org/docs/3.0/reference/varnish-cli.html)
* [Varnish CLI 4.0](https://www.varnish-cache.org/docs/4.0/reference/varnish-cli.html)
* [Varnish CLI 4.1+](https://www.varnish-cache.org/docs/trunk/reference/varnish-cli.html)

## Installation

### Download precompiled binary

Grab corresponding binary from [releases](https://github.com/akuznecov/varnishadm-ninja/releases)

### Build from source

`go get -u github.com/akuznecov/varnishadm-ninja`

## Running

```
$ varnishadm-ninja -T ADDRESS:PORT -S SECRET <command> [<arguments> <flags>]
```