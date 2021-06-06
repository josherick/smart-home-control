# Build go server
FROM golang:1-alpine as gobuilder
RUN apk --no-cache add ca-certificates git
WORKDIR /gobuild
COPY go.mod ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build

# Build python-kasa
FROM python:3.10-rc-alpine3.13 as pythonbuilder
RUN apk --no-cache add ca-certificates git gcc musl-dev libffi-dev rust cargo libressl-dev
RUN pip install poetry
WORKDIR /pythonbuild
RUN git clone https://github.com/josherick/python-kasa.git
WORKDIR python-kasa
RUN git fetch origin conditional-emeter-check
RUN git checkout conditional-emeter-check
RUN poetry build

# Copy binaries and run
FROM python:3.10-rc-alpine3.13
COPY --from=gobuilder /gobuild/smart-home-control .
COPY --from=pythonbuilder /pythonbuild/python-kasa/dist/*.whl .
COPY ./config.yaml .
RUN pip install *.whl
EXPOSE 8080
CMD ["./smart-home-control"]
