FROM golang:1.10.1
ADD server /
CMD ["/server"]
#EXPOSE 9090