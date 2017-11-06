#!/bin/sh

LIST="k-nucleotide-time \
      orig-time"

INPUT="input1.fasta input2.fasta input3.fasta input4.fasta"

for i in $LIST; do
  echo > ./$i.time
  for n in $INPUT; do
    for ncpu in $(seq 1 16); do
      for iters in $(seq 1 20); do
        echo "./$i/$i -ncpu $ncpu < ./input/$n >> ./$i.time"
        ./$i/$i -ncpu $ncpu < ./input/$n >> ./$i.time
        tail -1 ./$i.time
      done
    done
  done
done
