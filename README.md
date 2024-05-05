# cmpconflash
This tool allows you to **compare and contrast** multiple flash
firmware image binaries with respect to different flash page sizes.

The idea is to be able to identify how many block are common
between firmware binaries.

# Obligatory Install Line

```sh
go install github.com/linux4life798/cmpconflash
```

---

# Usage

```
$ cmpconflash --help
Compare multiple files in sequence at different block sizes.

Usage:
  cmpconflash <file1> <file2> [files...] [flags]

Flags:
      --all           When specified, all pairing of files will be compared
      --bsizes ints   List of block sizes to compare against (default [8192,4096,2048,1024,512,256,1])
  -h, --help          help for cmpconflash
      --offset int    Offset to start comparing in byte indices.
      --size int      Size of region to compare in bytes. A size of -1 means unbounded. (default -1)
```

# Example

1. When provided random input, ony some random byte positions might match.

    ```sh
    dd if=/dev/urandom bs=2M count=1 of=file1
    dd if=/dev/urandom bs=2M count=1 of=file2

    cmpconflash file1 file2
    ```
    *Output*
    ```
    # Compare file1 vs. file2
    Block Size    Blocks-Mismatched    Blocks-Matched    Blocks-Total    Percent Matched
    1             2088965              8187              2097152         0.390387%
    256           8192                 0                 8192            0.000000%
    512           4096                 0                 4096            0.000000%
    1024          2048                 0                 2048            0.000000%
    2048          1024                 0                 1024            0.000000%
    4096          512                  0                 512             0.000000%
    8192          256                  0                 256             0.000000%
    ```

2. If we look for smaller block sizes, like 2 consecutive bytes, we can find a
   few more matches.

    ```sh
    cmpconflash --bsizes "1,2,3,4" file1 file2
    ```
    *Output*
    ```
    # Compare file1 vs. file2
    Block Size    Blocks-Mismatched    Blocks-Matched    Blocks-Total    Percent Matched
    1             2088965              8187              2097152         0.390387%
    2             1048561              15                1048576         0.001431%
    3             699051               0                 699051          0.000000%
    4             524288               0                 524288          0.000000%
    ```

3. When we deliberately copy the last 1MB of the file1 to file2, we have 1024
   1k blocks that match, where we had none before.

    ```sh
    dd if=file1 of=file2 bs=1M skip=1 seek=1

    cmpconflash file1 file2
    ```
    *Output*
    ```
    # Compare file1 vs. file2
    Block Size    Blocks-Mismatched    Blocks-Matched    Blocks-Total    Percent Matched
    1             1044530              1052622           2097152         50.192928%
    256           4096                 4096              8192            50.000000%
    512           2048                 2048              4096            50.000000%
    1024          1024                 1024              2048            50.000000%
    2048          512                  512               1024            50.000000%
    4096          256                  256               512             50.000000%
    8192          128                  128               256             50.000000%
    ```
