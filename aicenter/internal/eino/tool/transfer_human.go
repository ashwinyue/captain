package tool

import (
	"context"
	"log"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// TransferHumanRequest is the input schema for transfer to human tool
type TransferHumanRequest struct {
	Reason string `json:"reason" jsonschema_description:"The reason for transferring to human agent, e.g., 'User requested human assistance', '转人工', '人工客服'"`
}

// TransferHumanResponse is the output schema for transfer to human tool
type TransferHumanResponse struct {
	Success bool   `json:"success" jsonschema_description:"Whether the transfer was successful"`
	Message string `json:"message" jsonschema_description:"A message describing the result"`
}

// TransferHumanCallback is the callback function type for transfer to human
type TransferHumanCallback func(ctx context.Context, reason string) error

// TransferHumanToolImpl implements the transfer to human logic
type TransferHumanToolImpl struct {
	callback TransferHumanCallback
}

const transferHumanDesc = `Transfer the conversation to a human customer service agent.
IMPORTANT: You MUST call this tool immediately when:
- User says "转人工", "人工客服", "human agent", "speak to agent", "transfer to human"
- User explicitly requests to speak with a human
- The issue is too complex for AI to handle

Do NOT ask for more details - just transfer them immediately.`

// NewTransferHumanTool creates a new transfer to human tool using eino's InferTool
func NewTransferHumanTool(callback TransferHumanCallback) tool.BaseTool {
	impl := &TransferHumanToolImpl{callback: callback}
	t, err := utils.InferTool("transfer_to_human", transferHumanDesc, impl.Invoke)
	if err != nil {
		log.Printf("[TransferHumanTool] Failed to create tool: %v", err)
		return nil
	}
	return t
}

// Invoke is the main function that will be called when the tool is invoked
func (t *TransferHumanToolImpl) Invoke(ctx context.Context, req *TransferHumanRequest) (*TransferHumanResponse, error) {
	log.Printf("[TransferHumanTool] Invoke called with reason: %s", req.Reason)

	if t.callback != nil {
		if err := t.callback(ctx, req.Reason); err != nil {
			log.Printf("[TransferHumanTool] Callback failed: %v", err)
			return &TransferHumanResponse{
				Success: false,
				Message: "Failed to transfer: " + err.Error(),
			}, nil
		}
	}

	log.Printf("[TransferHumanTool] Transfer successful")
	return &TransferHumanResponse{
		Success: true,
		Message: "Successfully transferred to human agent. The user will be connected to customer service shortly.",
	}, nil
}
