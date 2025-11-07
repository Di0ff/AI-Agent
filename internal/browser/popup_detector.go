package browser

import (
	"context"
	"encoding/json"
	"fmt"
)

type PopupDetector interface {
	DetectPopup(ctx context.Context, pageSnapshot *PageSnapshot) (*PopupInfo, error)
}

type PopupInfo struct {
	HasPopup         bool   `json:"has_popup"`
	CloseSelector    string `json:"close_selector"`
	PopupDescription string `json:"popup_description"`
	Reasoning        string `json:"reasoning"`
}

type LLMPopupDetector struct {
	llmClient LLMClient
}

type LLMClient interface {
	AnalyzePopup(ctx context.Context, elements string) (*PopupInfo, error)
}

func NewLLMPopupDetector(llmClient LLMClient) *LLMPopupDetector {
	return &LLMPopupDetector{
		llmClient: llmClient,
	}
}

func (d *LLMPopupDetector) DetectPopup(ctx context.Context, pageSnapshot *PageSnapshot) (*PopupInfo, error) {
	if pageSnapshot == nil || len(pageSnapshot.Elements) == 0 {
		return &PopupInfo{HasPopup: false}, nil
	}

	elementsJSON, err := json.Marshal(pageSnapshot.Elements)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal elements: %w", err)
	}

	return d.llmClient.AnalyzePopup(ctx, string(elementsJSON))
}
