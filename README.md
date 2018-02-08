stdio-wrapper
---

Run an executable that requires input/output files via stdin/stdout.

### Installation

```bash
$ go get github.com/rikonor/stdio-wrapper
```

### Usage

```bash
$ echo Hello | stdio-wrapper <executable> <args>

# Example
$ echo hello | stdio-wrapper text-doubler INPUT OUTPUT
hellohello
```

### License

MIT
