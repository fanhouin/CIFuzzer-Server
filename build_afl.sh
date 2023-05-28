#!/bin/sh
if [ -d "AFLplusplus" ]; then
    echo "AFLplusplus directory already exists." 
    exit 1
fi

sudo apt-get update -y
sudo apt-get install wget -y
sudo apt-get install -y unzip
sudo apt-get install -y build-essential python3-dev automake cmake git flex bison libglib2.0-dev libpixman-1-dev python3-setuptools cargo libgtk-3-dev
# try to install llvm 12 and install the distro default if that fails
sudo apt-get install -y lld-12 llvm-12 llvm-12-dev clang-12 || sudo apt-get install -y lld llvm llvm-dev clang
sudo apt-get install -y gcc-$(gcc --version|head -n1|sed 's/\..*//'|sed 's/.* //')-plugin-dev libstdc++-$(gcc --version|head -n1|sed 's/\..*//'|sed 's/.* //')-dev
sudo apt-get install -y ninja-build # for QEMU mode

wget https://github.com/AFLplusplus/AFLplusplus/archive/refs/tags/4.05c.zip  
unzip 4.05c.zip 
rm 4.05c.zip 
mv AFLplusplus-4.05c AFLplusplus  
cd AFLplusplus  
make distrib 
make install