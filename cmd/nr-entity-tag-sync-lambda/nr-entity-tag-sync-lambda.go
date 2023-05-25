package main

import (
	"context"
	"fmt"

	_ "github.com/newrelic/nr-entity-tag-sync/internal/provider/servicenow"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/newrelic/nr-entity-tag-sync/internal/sync"
	"github.com/newrelic/nr-entity-tag-sync/pkg/interop"
)

type TagSyncResult struct {
  Success           bool
  Message           error
}

func HandleRequest(ctx context.Context) (TagSyncResult, error) {
  i, err := interop.NewInteroperability()
  if err != nil {
    retErr := fmt.Errorf("failed to create interop: %s", err)
    return TagSyncResult{false, retErr}, retErr
  }

  defer i.Shutdown()

  syncer, err := sync.New(i)
  if err != nil {
    retErr := fmt.Errorf("sync failed: %s", err)
    return TagSyncResult{false, retErr}, retErr
  }

  err = syncer.Sync()
  if err != nil {
    retErr := fmt.Errorf("sync failed: %s", err)
    return TagSyncResult{false, retErr}, retErr
  }

  return TagSyncResult{true, nil}, nil
}

func main() {
  lambda.Start(HandleRequest)
}
