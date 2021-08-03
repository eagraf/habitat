# Habitat

## Basic Architecture

Habitat runs as a service that accepts requests from `habitatctl` in a client server model. Habitat will start and stop apps using `start.sh` and `stop.sh` scripts held in subdirectories of the `procs` folder.

Right now, the flow of data is `notes frontend <- notes backend <- fs library <- ipfs`.


## Building & Running

To rebuild the notes app, run `make install-notes` from the root directory. This will rebuild both the frontend and backend, and then install them in the right directories in `procs`pp, run `make install-notes` from the root directory. This will rebuild both the frontend and backend, and then install them in the right directories in `procs`.

To start the Habitat service, run `make run-dev` in the root directory of this repository. Habitat will start listening for commands from `habitatctl`. `make run-dev` will rebuild the `habitat` and `habitatctl` binaries.

Right now, the two commands `habitatctl` supports are `habitatctl start` and `habitatctl stop` which are used to start and stop apps on Habitat. For example, to start the notes app, you would run the following:
```
habitatctl start ipfs
habitatctl start notes_backend
habitatctl start nginx
```

### Development on notes app

For development on the Notes app, its usually better to just have IPFS running via  `habitatctl start`, but then running the frontend and backend on your machine.

Frontend:
Run `npm start` in the frontend directory. This will run the frontend with hot reloading.

Backend:
Right now, just `./out/backend/notes-api` from the notes app directory. We should find a better way of running this.

