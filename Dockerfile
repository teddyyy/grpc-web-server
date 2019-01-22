FROM golang:1.10.1
ADD bin/server /
CMD ["/server"]
