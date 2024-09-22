FROM public.ecr.aws/docker/library/golang:1.23-alpine AS gobuild

WORKDIR /app

COPY main.go .
COPY go.mod .
COPY go.sum .
COPY internal/ internal
COPY vendor/ vendor

RUN go build -mod vendor -o app *.go

FROM alpine:3.18

RUN apk --no-cache add ca-certificates
RUN apk --no-cache add tzdata

COPY --from=gobuild /app/app .

CMD ./app