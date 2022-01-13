# Various `memr` Examples

## [compression](./compression)
- Demonstrates how to read and compress memory with various algorithms
- Supported formats: `snappy`, `lz4`, `zlib`, `gzip`
- See also: [decompressor](./decompressor)

## [custom header](./custom_header)
- Demonstrates how to prepend a custom header to memory pages
- Particularly useful if custom page handling is being performed (see also: [page compression](./page_compression))

## [decompressor](./decompressor)
- Simple example taking a compressed input file and decompressing it using the specified algorithm
- Supported input formats: `snappy`, `lz4`, `zlib`, `gzip`
- See also: [compression](./compression)

## [page compression](./page_compression) [**warning: this is not recommended**]
- Example showing how to perform page-level compression using `memr`
- Note: this is not advisable, and the method defined in the [compression](./compression) example should be used instead

## [progress](./progress)
- Simple example showing how to enable/disable progress reporting during read

## [source](./source)
- Demonstrates how to read from a specific memory source (`/proc/kcore`, `/dev/crash`, `/dev/mem`)
- The alternative to this is to use `memr.Probe()`, which attempts to load a valid ready from any source

## [stream s3](./stream_s3)
- Demonstrates how to read memory and stream it directly to an S3 bucket
- This is particularly useful in AWS environments when disk space may be limited and network I/O is very high
- A similar approach could be applied to other cloud environments using their respective SDKs
