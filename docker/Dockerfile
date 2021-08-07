FROM karalabe/xgo-base

ENV GO_VERSION 11603
ENV ROOT_DIST https://storage.googleapis.com/golang/go1.16.3.linux-amd64.tar.gz
ENV ROOT_DIST_SHA 951a3c7c6ce4e56ad883f97d9db74d3d6d80d5fec77455c6ada6c1f7ac4776d2

ADD bootstrap_pure.sh /bootstrap_pure.sh
RUN chmod +x /bootstrap_pure.sh
RUN /bootstrap_pure.sh

ENV ALSA_LIB_VERSION 1.2.4

RUN apt-get update && apt-get install -y \
    curl

RUN mkdir /alsa && \
    curl "ftp://ftp.alsa-project.org/pub/lib/alsa-lib-${ALSA_LIB_VERSION}.tar.bz2" -o /alsa/alsa-lib-${ALSA_LIB_VERSION}.tar.bz2

# https://www.programering.com/a/MTN0UDMwATk.html
# https://stackoverflow.com/questions/36195926/alsa-util-1-1-0-arm-cross-compile-issue
RUN cd /alsa && \
    tar -xvf alsa-lib-${ALSA_LIB_VERSION}.tar.bz2 && \
    cd alsa-lib-${ALSA_LIB_VERSION} && \
    CC=arm-linux-gnueabihf-gcc-5 ./configure --host=arm-linux && \
    make && \
    make install