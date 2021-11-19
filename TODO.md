# TODO List

* Dynamically create the plugins needed by CNI and put them in the docker image.

* Connect the container to the "veth cable."

* Decide whether to create a JSON config for the network/veth/bridge.

* Pull all this together inside a multi-stage dockerfile example.

* Decide on the following options:

1. Create a Makefile and put the all recipe to include the testing phase first.

2. Create a Dockerfile specifically for testing which docker-compose can start up.

3. Figure out how to connect/start up the grpc server normally running on /run/containerd/containerd.sock.
