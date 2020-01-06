# docker build -t wifinina .
# docker run wifinina -v "../build/wifinina:/src/build"

FROM debian:stable-slim AS esp
WORKDIR /src

RUN apt-get clean && apt-get update && \
    apt-get install -y sudo wget gcc git wget libncurses-dev flex bison gperf build-essential \
    python python-pip python-setuptools python-serial python-cryptography python-future python-pyparsing make

RUN mkdir /src/wifinina && \
	cd /src/wifinina && \
	wget https://dl.espressif.com/dl/xtensa-esp32-elf-linux64-1.22.0-80-g6c4433a-5.2.0.tar.gz && \
	mkdir -p /src/esp && \
	cd /src/esp && \
	tar -xzf /src/wifinina/xtensa-esp32-elf-linux64-1.22.0-80-g6c4433a-5.2.0.tar.gz

RUN cd /src/esp && \
    git clone --branch v3.3.1 --recursive https://github.com/espressif/esp-idf.git

FROM esp AS nina

RUN cd /src/esp && \
    git clone https://github.com/arduino/nina-fw.git

COPY ./firmware.sh /src
RUN chmod +x /src/firmware.sh
ENTRYPOINT ["/src/firmware.sh"]
