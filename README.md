# SketchQuest
A website to play a game where guessers have a time limit to guess what is a drawer's sketch is for points.

## Build
Guess the sketch provides a Dockerfile to build and run the server.

Build a docker image from the docker file using `docker build -t guessthesketch:1 .`

Run the dockerfile using `docker run -d -p 8080:8080 guessthesketch:1`
The argument `-p 8080:8080` will map the host port to the container's port.
