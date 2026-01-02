package trace

import (
	"context"
	"log"
	"os"

	clc "github.com/cloudwego/eino-ext/callbacks/cozeloop"
	"github.com/cloudwego/eino/callbacks"
	"github.com/coze-dev/cozeloop-go"
)

// CloseFn is a function to close the trace client
type CloseFn func(ctx context.Context)

// InitCozeLoop initializes CozeLoop tracing if configured via environment variables.
// Required env vars:
//   - COZELOOP_WORKSPACE_ID: CozeLoop workspace ID
//   - COZELOOP_API_TOKEN: CozeLoop API token
//
// Returns a close function to flush and close the client.
func InitCozeLoop(ctx context.Context) CloseFn {
	wsID := os.Getenv("COZELOOP_WORKSPACE_ID")
	apiToken := os.Getenv("COZELOOP_API_TOKEN")

	if wsID == "" || apiToken == "" {
		log.Println("[trace] CozeLoop not configured, skipping initialization")
		return func(ctx context.Context) {}
	}

	client, err := cozeloop.NewClient(
		cozeloop.WithWorkspaceID(wsID),
		cozeloop.WithAPIToken(apiToken),
	)
	if err != nil {
		log.Printf("[trace] Failed to create CozeLoop client: %v", err)
		return func(ctx context.Context) {}
	}

	// Register as global callback handler
	handler := clc.NewLoopHandler(client)
	callbacks.AppendGlobalHandlers(handler)

	log.Printf("[trace] CozeLoop initialized successfully, workspace: %s", wsID)
	log.Println("[trace] View traces at: https://loop.coze.cn")

	return client.Close
}
