# `memr` Example Streaming Directly to S3

## basic upload usage

    go run . --bucket=<BUCKET> --key=<KEY>

## upload with lz4 compression (default=snappy)

    go run . --bucket=<BUCKET> --key=<KEY> --compression=lz4

## upload with 10 goroutines (default=5)

**Warning**: This will use more memory on a system because of the extra buffers allocated in memory.

Setting `--concurrency` to a smaller value (like 1) will result in a minimal memory footprint.

    go run . --bucket=<BUCKET> --key=<KEY> --concurrency=10
