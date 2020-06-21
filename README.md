# SchemaDiff: API verification from JSON to Go

It's not uncommon to get a JSON value that you might not have a fully
specified schema for. `SchemaDiff` is used to verify if your Go structs have
complete and correct field descriptions for a given JSON value.

Our operating example will be a Reddit response.

```sh
curl http://reddit.com/r/Longreads/.json > longreads.json
```

```go

import (
    "fmt"

    "github.com/nathanwiegand/schemadiff"
)

type Reddit struct {

}

func main() {

}
```

We start with an empty struct. Running the program we see:

```txt

```

// TODO: