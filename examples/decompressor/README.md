# Example Decompressor

This demonstrates how to decompress data from various compression formats.

See also: [examples/compression](../compression)

Each command below can also use `--output=<FILE>` to specify the output file (default=`output.memr`)

## snappy (default)

    go run . --input=<FILE> --compression=snappy

## lz4 (comparable to snappy)

    go run . --input=<FILE> --compression=lz4

## zlib (not recommended; demo purposes only)

    go run . --input=<FILE> --compression=zlib

## gzip (not recommended; demo purposes only)

    go run . --input=<FILE> --compression=gzip

## none (explicitly disabled)

    go run . --input=<FILE> --compression=none
