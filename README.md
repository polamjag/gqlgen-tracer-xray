# gqlgen-tracer-xray

[AWS X-Ray][xray] Tracer for [gqlgen][]

## Synopsis

```go
package main

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/aereal/gqlgen-tracer-xray/gqlgentracerxray"
)

func main() {
  handler.GraphQL(
    NewExecutableSchema(...),
    gqlhandler.ComplexityLimit(1000), // Recommended for record complexity
    gqlgentracerxray.New(),
  )
}
```

## See also

- [99designs/gqlgen-contrib][gqlgen-contrib]

[xray]: https://aws.amazon.com/xray/
[gqlgen]: https://gqlgen.com/
[gqlgen-tracer]: https://github.com/99designs/gqlgen/blob/master/graphql/tracer.go
[gqlgen-contrib]: https://github.com/99designs/gqlgen-contrib
