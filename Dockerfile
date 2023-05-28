FROM golang:1.18

WORKDIR /app
COPY . /app

RUN apt-get update -y \ 
    && apt-get install vim -y \ 
    && apt-get install wget -y \
    && apt-get install -y unzip \
    && apt-get install -y build-essential python3-dev automake cmake git flex bison libglib2.0-dev libpixman-1-dev python3-setuptools cargo libgtk-3-dev \
    && apt-get install -y lld-12 llvm-12 llvm-12-dev clang-12 || apt-get install -y lld llvm llvm-dev clang \
    && apt-get install -y gcc-$(gcc --version|head -n1|sed 's/\..*//'|sed 's/.* //')-plugin-dev libstdc++-$(gcc --version|head -n1|sed 's/\..*//'|sed 's/.* //')-dev \
    && apt-get install -y ninja-build


RUN rm -rf AFL* \ 
    && wget https://github.com/AFLplusplus/AFLplusplus/archive/refs/tags/4.05c.zip \ 
    && unzip 4.05c.zip \
    && rm 4.05c.zip \
    && mv AFLplusplus-4.05c AFLplusplus \ 
    && cd AFLplusplus \ 
    && make distrib\ 
    && make install

RUN go mod tidy 

EXPOSE 8080
ENTRYPOINT ["go", "run", "CIServer.go"]

