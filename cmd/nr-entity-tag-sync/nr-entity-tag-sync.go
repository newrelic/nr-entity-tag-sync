package main

import (
	"fmt"
	"os"

	_ "github.com/newrelic/nr-entity-tag-sync/internal/provider/servicenow"

	"github.com/newrelic/nr-entity-tag-sync/internal/sync"
	"github.com/newrelic/nr-entity-tag-sync/pkg/interop"
)

func main() {
  i, err := interop.NewInteroperability()
  if err != nil {
    fmt.Printf("failed to create interop: %s", err)
    os.Exit(1)
  }

  defer i.Shutdown()

  err = sync.Sync(i)
  if err != nil {
    fmt.Println(err)
  }
}
