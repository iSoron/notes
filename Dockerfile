FROM golang:1.12-alpine as builder
RUN apk add --no-cache git make rsync nodejs npm
WORKDIR /go/notes
COPY . .
RUN make install-deps all

FROM alpine:latest 
VOLUME /data
EXPOSE 8050
COPY --from=builder /go/notes/build/ /
ENTRYPOINT ["/notes"]
CMD ["--data","/data","--allow-file-uploads","--max-upload-mb","10","--host","0.0.0.0"]
