FROM golang
RUN apt-get update && apt-get -y install iptables dbus
COPY . /go/src/github.com/eyedeekay/docker-i2p-plugin
WORKDIR /go/src/github.com/eyedeekay/docker-i2p-plugin
RUN go mod download
ENTRYPOINT ["docker-i2p-plugin"]
