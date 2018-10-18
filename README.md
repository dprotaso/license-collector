# license-collector

`license-collector` is a tool for gathering up a collection of licenses that are embedded
in Debian Stretch (ie. gcr.io/distroless/base) based container image.

The resulting output from `license-collector` is the concatenation of
all the licenses discovered

### Basic Usage

```shell
$ license-collector gcr.io/cloud-builders/gcs-fetcher
===========================================================
image: gcr.io/cloud-builders/gcs-fetcher:latest@sha256:9b9e80737a9c890c63d4487b08803cf3c0328b0f4c1c5f0e76f94acdf13bd4a7
file:  /usr/share/doc/ca-certificates/copyright
contents:

        Format: http://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
        Source: http://ftp.debian.org/debian/pool/main/c/ca-certificates/

        Files: debian/*
               examples/*
               Makefile
               mozilla/*
               sbin/*
...

```