FROM golang as builder
COPY vendor/ /go/src/github.com/parkr/docker-kill-unhealthy-containers/vendor
COPY cmd/ /go/src/github.com/parkr/docker-kill-unhealthy-containers/cmd
RUN CGO_ENABLED=0 GOOS=linux \
    go install -v -ldflags '-w -extldflags "-static"' \
    github.com/parkr/docker-kill-unhealthy-containers/...

FROM scratch
COPY --from=builder /go/bin/docker-kill-unhealthy-containers /bin/docker-kill-unhealthy-containers
ENTRYPOINT [ "/bin/docker-kill-unhealthy-containers" ]
