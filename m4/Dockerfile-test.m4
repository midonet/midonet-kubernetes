include(Dockerfile.pre-build)
RUN go test -v ./... | tee go_test.output
RUN go2xunit -input go_test.output -output /tmp/junit.xml

FROM builder as test_fail_checker
RUN go2xunit -input go_test.output -fail
