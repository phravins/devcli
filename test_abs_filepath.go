package main

import (
	"fmt"
	"path/filepath"
)

func main() {
    base := "/app"
    testPaths := []string{
        "foo",
        "/etc/passwd",
        "../foo",
    }

    for _, p := range testPaths {
        if filepath.IsAbs(p) {
            fmt.Printf("%q is absolute. Cannot allow this as a relative path.\n", p)
        } else {
            fmt.Printf("%q joined with %q is %q\n", p, base, filepath.Join(base, p))
        }
    }
}
