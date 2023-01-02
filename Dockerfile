FROM golang
RUN apt-get update && apt-get -y install iptables dbus
RUN go get github.com/tools/godep
COPY . /go/src/github.com/eyedeekay/docker-i2p-plugin
WORKDIR /go/src/github.com/eyedeekay/docker-i2p-plugin
RUN godep go install -v
ENTRYPOINT ["docker-i2p-plugin"]
