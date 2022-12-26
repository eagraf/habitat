# Habitat

Habitat is a personal server for you and your friends. 

## Build System
Habitat uses GNU Make as a build system. The majority of the build framework is expressed in `common.mk`, and recursive Makefiles are used to build each subdirectory. On MacOS, `gmake` should be used instead of `make`.

## Building and Running
The two main binaries that can be built from this repository are `habitat` and `habitatctl`. The `habitat` program is a server, and `habitatctl` is a command line client for interacting with it. To build these, run
```
make -C cmd build
```
For local development, these binaries are placed in the `./dist/bin` folder relative to the project directory. To run the `habitat` server in development mode, run:
```
make run
```
Then, you can use the `habitatctl` client to interact with the server. For example:
```
./dist/bin/habitatctl inspect
```
This repository also contains a number of application projects, which are stored under the `./apps` directory. To build and install them for local development, run:
```
make -C apps all install
```
This will build all application binaries, web files, etc necesarry for `habitat` to start these apps as subprocesses. 

## Source Directories
We do our best to follow the [standard Go project layout](https://github.com/golang-standards/project-layout). This is a work in progress, and many things should be moved into an `internal` folder that does not yet exist.

## Runtime Directories
All files generated at runtime are placed under the `HABITAT_PATH` directory. This includes all object store, Raft log stable storage, and node configuration. In development, `HABITAT_PATH` is set to the `.habitat` direectory under the root project directory. On a typical deployed instance of Habitat, it might be found at `~/.habitat` for the user running it.

When Habitat is asked to start an application, it searches for an installation of the application using the `HABITAT_APP_PATH` environment variable. `HABITAT_APP_PATH` works similar to the `PATH` variable on Unix like systems. Multiple directores can be listed, separated by semicolons. This arrangement makes it easy for apps in other repositories to be found. Each entry in `HABITAT_APP_PATH` should be a directory with one or more subdirectories, each representing an application. 

At the bare minimum, application directories need to provide a `habitat.yaml` file, and a `bin` directory. The `habitat.yaml` acts as an application manifest. Additionally, all binaries produced by the application should be placed under `bin/<arch_os>/<bin_name>` paths. All web content that the app serves should be placed under a directory named `web`. Both `bin` and `web` should sit at the top level in the application directory.
