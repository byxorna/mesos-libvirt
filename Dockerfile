FROM golang:1.5
MAINTAINER Gabe Conradi <gabe.conradi@gmail.com>

COPY . /go/src/github.com/byxorna/mesos-libvirt
WORKDIR /go/src/github.com/byxorna/mesos-libvirt
RUN make setup
RUN make
# expects mesos master on master:5050
EXPOSE 8080
CMD ["./framework","-master","master:5050","-logtostderr","-artifactPort","8080"]



