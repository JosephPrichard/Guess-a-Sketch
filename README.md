# Skribbl
This is my attempt to make a clone of scribbl.io, a website to play a classic game of pictionary over the internet.
The project was created to learn about working with websockets in Go - and to learn the Solid.js framework.

## Build
Provides a Dockerfile to build and run the server.

Build a docker image from the docker file using `docker build -t myserver:1 .`

Run the dockerfile using `docker run -d -p 8080:8080 myserver:1`
The argument `-p 8080:8080` will map the host port to the container's port.
