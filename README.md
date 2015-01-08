xkg - X Keyboard Grabber
========================

[![GoDoc](https://godoc.org/github.com/henriquemenezes/xkg?status.png)](https://godoc.org/github.com/henriquemenezes/xkg)

## Installation

    go get github.com/henriquemenezes/xkg

## Usage example:

```go
package main

import (
	"fmt"

	"github.com/henriquemenezes/xkg"
)

func main() {
	var keys = make(chan int, 100)

	go xkg.StartXGrabber(keys)

	for {
		keycode := <-keys

		if key, ok := xkg.KeyMap[keycode]; ok {
			fmt.Printf("[%s]", key)
		}
	}
}
```

## License

The license of the project is [The MIT License (MIT)](https://github.com/henriquemenezes/xkg/blob/master/LICENSE).
