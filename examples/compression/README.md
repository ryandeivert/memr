# `memr` Example With Compression

This demonstrates the preferred method of compressing memory read from a
system versus page-level compression.

The resulting archive will be in a format that can be decompressed by common tools.
Once decompressed, the resulting image will be in a format (LiME) that is already well adopted.

See also: [examples/decompressor](../decompressor)

Each command below can also use `--output=<FILE>` to specify the output file (default=`output.memr`)

## snappy (default)

    go run . --compression=snappy

## lz4 (comparable to snappy)

    go run . --compression=lz4

## zlib (not recommended; demo purposes only)

    go run . --compression=zlib

## gzip (not recommended; demo purposes only)

    go run . --compression=gzip

## none (explicitly disabled)

    go run . --compression=none
