FROM golang:1.15

# Copy code resources
COPY ./cmd /root/hsbc-hw/cmd
COPY ./model /root/hsbc-hw/model
COPY ./serving /root/hsbc-hw/serving

COPY build.sh /root/hsbc-hw/build.sh

WORKDIR /root/hsbc-hw

RUN chmod a+x ./build.sh

# for unit test
EXPOSE 8083

RUN ./build.sh && cp /root/hsbc-hw/bin/server /bin/server

RUN chmod a+x /bin/server

ENTRYPOINT ["/bin/server"]
