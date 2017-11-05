#!/bin/bash

rm measurements.log

runtimes="./runtimes"
basedir=${PWD}
niters=30
ngo=2

for i in `find . -type d`; do
  cd $i
  go build 2> >(grep -v "no Go files")
  cd ${basedir}
done

cd ${basedir}

[ -d ${runtimes} ] && cp -r ${runtimes} ./runtimes_old && rm -rf ${runtimes}

mkdir ${runtimes}

files=`find ./gather ./scatter ./alltoall -type f -executable`

for ((j=0; j<${niters}; j++)); do
  echo "Iteration ${j} of ${niters}" >> "measurements.log"
  for ((c=1; c<=${ngo}; c++)); do
    echo "    ngo ${c} of ${ngo}" >> "measurements.log"
    for i in ${files}; do
      filename=`echo "${i%/*}"  | sed 's/\.\///g'| sed 's/\//\_/g'`
      ${i} -ncpu ${c} >> ${runtimes}/${filename}_${c}.time
    done
  done
done
