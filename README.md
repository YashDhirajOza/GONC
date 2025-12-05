# gonc — Pure-Go minimal NetCDF reader

**gonc** is a small, pure-Go (no CGO) library that implements the initial parts of a NetCDF Classic/64-bit-offset reader (header parsing and dimension listing). It is being developed as a learning-first, production-minded project to read NetCDF `.nc` files using only Go's standard library.

> **Status:** Early prototype — reads magic header, `numrecs`, and `dim_list` (dimension names + lengths). Does **not** yet parse attributes, variables, or variable data. Does not support netCDF-4/HDF5. See *Future milestones* below for planned work.

---

## Why a pure-Go NetCDF reader?

During development we ran into CGO-related tooling problems on Windows (32-bit GCC vs 64-bit Go toolchain). Instead of solving platform-specific CGO issues, this project explores a pure-Go approach that:

* avoids external C dependencies and linking
* produces single static Go binaries
* is easier to build and ship across platforms
* serves as a learning and research project for binary format parsing and robust IO in Go

Note: implementing full netCDF-4 (HDF5-backed) functionality in pure Go is a massive effort. This project targets the **Classic** (CDF-1) format and the 64-bit-offset variant (CDF-2) initially — netCDF-4/HDF5 is out of scope for now.

---

## Current features

* Read NetCDF magic bytes (`CDF`) and format byte (Classic vs 64-bit-offset)
* Read `numrecs` (the 32‑bit record count)
* Parse `dim_list` and expose dimensions as `[]Dimension` on `File`
* Simple, idiomatic Go API (`Open`, `Close`, `File.Dims`)
* Pure Go — no CGO required

---

## Project layout (recommended)

```
floatchat-gopy/              # module root (example module name from development)
├── go.mod                   # module floatchat-gopy
├── main.go                  # example using gonc
└── gonc/
    └── netcdf.go           # library implementation (package gonc)
```

> Your module name may be different. Imports within your code should use the module path from your `go.mod` (for example `floatchat-gopy/gonc`).

---

## Quick usage example

Create an example `main.go` (in the module root):

```go
package main

import (
    "fmt"
    "log"
    "time"

    "floatchat-gopy/gonc" // replace with your module path
)

func main() {
    start := time.Now()

    nc, err := gonc.Open("path/to/yourfile.nc")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    if nc.Format == gonc.ClassicFormat {
        fmt.Println("Format: Classic (CDF-1)")
    } else {
        fmt.Println("Format: 64-bit Offset (CDF-2)")
    }

    fmt.Println("NumRecs:", nc.NumRecs)
    fmt.Println("Dimensions:")
    for _, d := range nc.Dims {
        fmt.Printf("  %s = %d\n", d.Name, d.Length)
    }

    fmt.Println("Total execution time:", time.Since(start))
}
```

Run it:

```bash
go run .
```

---

## API (so far)

```go
package gonc

type File struct {
    f      *os.File
    Format byte
    NumRecs uint32
    Dims   []Dimension
}

type Dimension struct {
    Name   string
    Length uint32
}

func Open(path string) (*File, error)
func (nc *File) Close() error
```

`Open` reads the header (magic, format, `numrecs`) and calls the internal `readDimList()` helper to populate `File.Dims`.

---

## Limitations & Notes

* **Read-only header parsing**: currently the library only parses header information (dimensions). It does not parse global or variable attributes or read variable data.
* **No netCDF-4/HDF5 support**: netCDF-4 files (HDF5 backend) are *not* supported — implementing HDF5 in pure Go is a large task and out of scope for the initial work.
* **CDF-2 support**: the code recognizes the format byte for 64-bit offset files but has not been fully validated on CDF-2 files. Further testing and fixes may be required.
* **No external C dependencies**: the library intentionally avoids CGO and `libnetcdf` to simplify building on Windows and other platforms.

---

## Development and testing

* Use `go test` (we will add unit tests in a subsequent milestone)
* Use `go vet`, `go fmt`, and `golangci-lint` during development
* Example debugging commands on Windows:

  * Use `xxd` or a hex-dump utility to inspect file headers
  * Ensure your file path is correct (avoid stray `\n` inside string literals)

---

## Future milestones

This section lists planned work and milestones. Each item can become its own issue/PR.

### Short-term (next 1–2 sprints)

1. **Global attributes (`gatt_list`) parser** — parse and expose file-level attributes
2. **Variable metadata (`var_list`) parser** — read variable names, types, dimension refs, offsets, and attributes
3. **Variable data reading** — support reading basic data types: `int8`, `int16`, `int32`, `float32`, `float64`, fixed-length strings
4. **Support unlimited/record variables** (record section parsing)
5. **Add unit tests** covering classic CDF files and common edge cases
6. **Add example CLI utility** (tiny `gondump` / `gonc-dump`) to print header/vars like `ncdump -h`

### Mid-term (bigger features)

1. **Slicing/subsetting API** — read hyperslabs or slices from variables without loading all data
2. **Memory-efficient read modes** — support streaming, `io.ReaderAt` and optional mmap backends
3. **Benchmarking & profiling** — measure hot paths and optimize IO
4. **Cross-platform CI** — GitHub Actions to build/test on Linux/macOS/Windows

### Long-term / Research (very large scope)

1. **Partial netCDF-4 (HDF5) support** — only if we either bind to HDF5 C libs or implement a minimal HDF5 reader in Go (very large)
2. **Write support** — allow creating and writing classic NetCDF files in pure Go
3. **Advanced features** — support chunking, compression filters, virtual datasets if HDF5 support added

---

## Contribution guidelines

Contributions are welcome. Suggested workflow:

1. Fork the repo
2. Create a branch: `git switch -c feat/var-list`
3. Add tests for the new feature
4. Open a PR with description, test results, and benchmarks

Please keep changes small and add tests for any bugfix/feature.

---

## About this implementation

This README and the current code were developed interactively while prototyping steps to parse NetCDF headers in Go. Development notes and decisions (including the choice to avoid CGO due to compiler problems on Windows) are documented in the project's issues and future milestones above.

---

## License

Choose a license for the project (MIT recommended for small libraries). Add a `LICENSE` file if you want permissive reuse.

---

If you want, I can:

* generate a `README.md` file in your repo (this file is already created for you in the canvas),
* create example unit tests for `readString` and `readDimList`, or
* scaffold a small CLI `gonc-dump` that prints header information.

Which of these should I do next?
