# `memr` Example With Specific Source

Each command below can also use `--output=<FILE>` to specify the output file (default=`output.memr`)

## /proc/kcore

    go run . --source /proc/kcore

## /dev/crash

    go run . --source /dev/crash

## /dev/mem

    go run . --source /dev/mem
