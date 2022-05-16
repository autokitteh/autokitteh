#!/bin/bash

set -euo pipefail

p2 -t /proto/src/template-values.yaml -i /proto/src/patterns.yaml > /gen/proto/src/template-values.yaml

find /proto/src -mindepth 1 -type d | while read -r indir; do
  outdir="/gen/proto/src/$(basename "${indir}")"
  mkdir -p "${outdir}"

  for infile in "${indir}"/*; do
    outfile="${outdir}/$(basename "${infile}")"

    echo "${infile} -> ${outfile}"

    p2 -t "${infile}" -i /gen/proto/src/template-values.yaml > "${outfile}"
  done
done
