cmdio
---

Run an executable and provide it input/output files via io.Reader/io.Writer.

### Usage

```go
package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

	cmdio "github.com/rikonor/go-cmdio"
)

func main() {
	r := strings.NewReader("Hello!")
	w := os.Stdout

	execPath := "./text-doubler"
	execArgs := []string{"INPUT", "OUTPUT"}

	tmpArgs, closeFn, err := cmdio.Wrap(r, w, execArgs)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFn()

	cmd := exec.Command(execPath, tmpArgs...)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
```

### License

MIT
