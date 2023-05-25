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
    fmt.Printf("failed to create interop: %s]\n", err)
    os.Exit(1)
  }

  defer i.Shutdown()

  syncer, err := sync.New(i)
  if err != nil {
    fmt.Printf("sync failed: %s\n", err)
    os.Exit(2)
  }

  err = syncer.Sync()
  if err != nil {
    fmt.Printf("sync failed: %s\n", err)
    os.Exit(3)
  }
}
