FROM golang:1.12.0
ADD . ./src/
# EXPOSE 8080
RUN ["go", "get", "github.com/globalsign/mgo"]
RUN ["go", "get", "github.com/koron/go-dproxy"]
RUN ["go", "build", "./src/test_salesforce_api.go"]
CMD ["./test_salesforce_api"]
