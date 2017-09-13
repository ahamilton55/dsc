FROM golang:1.9.0
COPY . src/github.com/ahamilton55/dsc
WORKDIR src/github.com/ahamilton55/dsc
RUN go test -v -bench . && go build -o nginx-stats

FROM scratch
COPY --from=0 /go/src/github.com/ahamilton55/dsc/nginx-stats .
CMD ["/nginx-stats"]