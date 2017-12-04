#!/bin/bash

runtimes="./runtimes"
results="./avgs"
basedir=${PWD}
niters=40
ngo=16

files=`find ./gather ./scatter ./alltoall -type f -executable`

for i in ${files}; do
  echo $i
  filename=`echo "${i%/*}"  | sed 's/\.\///g'| sed 's/\//\_/g'`
  rm ${results}/${filename}.avg
  touch ${results}/${filename}.avg
  for ((c=1; c<=${ngo}; c++)); do
    awk '{s+=$1; ss+=$1^2}END{print "'$c'",(sqrt(ss/NR-(s/NR)^2)),(s/NR)}' ${runtimes}/${filename}_${c}.time >> ${results}/${filename}.avg
  done
done
