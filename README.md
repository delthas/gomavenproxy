# gomavenproxy [![builds.sr.ht status](https://builds.sr.ht/~delthas/gomavenproxy.svg)](https://builds.sr.ht/~delthas/gomavenproxy?)

A small HTTP server to let users transparently upload Maven artifacts to an FTP server.

*Project stability: successfully tested in a production environment.*

## Setup

- copy `gomavenproxy.example.yml` and edit it
- run `gomavenproxy -config gomavenproxy.yml`

## Usage

For a `build.gradle` file:
```
repositories {
    maven {
        url 'http://gomaven.proxy:12345'
        credentials {
            username "ftp_username"
            password "ftp_password"
        }
    }
}
```

For a `build.gradle.kts` file:
```
repositories {
    maven {
        url = uri("http://gomaven.proxy:12345")
        credentials {
            username = "ftp_username"
            password = "ftp_password"
        }
    }
}
```

## Rationale

Maven does have a Wagon plugin to support deploying artifacts by FTP.

Gradle supports publishing modules by FTP with the deprecated `maven` plugin, which in particular does not support [Gradle Module Metadata](https://docs.gradle.org/current/userguide/publishing_gradle_module_metadata.html).

For the new `maven-publish` plugin, only [a few protocols are supported](https://docs.gradle.org/current/userguide/declaring_repositories.html#sec:supported_transport_protocols), but not FTP.

This HTTP to FTP proxy lets you use `maven-publish` on an FTP repository by proxying it as an HTTP repository.

## Builds

| OS | URL |
|---|---|
| Linux x64 | https://delthas.fr/gomavenproxy/linux/gomavenproxy |
| Mac OS X x64 | https://delthas.fr/gomavenproxy/mac/gomavenproxy |
| Windows x64 | https://delthas.fr/gomavenproxy/windows/gomavenproxy.exe |
