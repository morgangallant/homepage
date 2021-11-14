FROM golang as build
ADD . /src
WORKDIR /src/
RUN go get .
RUN go build -o homepage .

FROM gcr.io/distroless/base
WORKDIR /mg
COPY --from=build /src/homepage /mg/
COPY --from=build /src/site /mg/
ENTRYPOINT [ "/mg/homepage" ]