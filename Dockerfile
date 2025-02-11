FROM golang:1.23.4

WORKDIR /avito-shop

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN cd cmd/merch && go build

EXPOSE 8080

CMD [ "./cmd/merch/merch" ]
