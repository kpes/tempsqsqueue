GOOS=

.PHONY: build-examples
build-examples:
	@printf "Building lambda to lambda example ... "
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -tags lambda.norpc -o examples/bin/lambda-to-lambda/client/bootstrap examples/lambda-to-lambda/client/main.go
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -tags lambda.norpc -o examples/bin/lambda-to-lambda/server/bootstrap examples/lambda-to-lambda/server/main.go
	@printf "Done\n"