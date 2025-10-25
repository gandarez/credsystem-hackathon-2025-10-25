package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dyammarcano/crew-das-closures/internal/client/openrouter"
	"github.com/dyammarcano/crew-das-closures/internal/model"
	"github.com/dyammarcano/crew-das-closures/internal/prompt"
)

type Core struct {
	*openrouter.Client
	*prompt.PromptManager
}

var findReqPool = sync.Pool{New: func() any { return new(model.FindServiceRequest) }}

func NewCore(urlStr string, opts openrouter.Option) (*Core, error) {
	return &Core{
		Client:        openrouter.NewClient(urlStr, opts),
		PromptManager: prompt.NewPromptManager(),
	}, nil
}

// AskQuestion decodes the request, prepares a (mock) service response, and analyzes
// coherence issues between input and output. Diagnostics are returned to help
// detect problems early when integrating with external APIs.
func (c *Core) AskQuestion(question []byte) (*model.FindServiceResponse, error) {
	obj := findReqPool.Get().(*model.FindServiceRequest)
	// reset fields (only one field now, but future-proof)
	*obj = model.FindServiceRequest{}
	if err := json.Unmarshal(question, obj); err != nil {
		findReqPool.Put(obj)
		return nil, err
	}
	defer findReqPool.Put(obj)

	// montar o prompt para o OpenRouter com base no obj.Intent
	result, err := c.PromptManager.GenerateModelSpecificPrompt(obj.Intent)
	if err != nil {
		return nil, err
	}

	var msgs []openrouter.PromptMessage

	msg := openrouter.PromptMessage{
		Role:    "user",
		Content: result,
	}

	msgs = append(msgs, msg)

	oRequest := &openrouter.OpenRouterRequest{
		Model:    "gpt-4o",
		Messages: msgs,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	response, err := c.Client.ChatCompletion(ctx, oRequest)
	if err != nil {
		return nil, err
	}

	// Normalize/validate the model output against the canonical service registry
	normalizedID, normalizedName, normDiag := c.normalizeServicePair(response.ServiceID, response.ServiceName)

	sData := &model.ServiceData{
		ServiceID:   normalizedID,
		ServiceName: normalizedName,
	}

	diagnostics := analyzeCoherence(obj, sData)
	if normDiag != "" {
		diagnostics = append(diagnostics, normDiag)
	}

	if len(diagnostics) > 0 {
		// logar os diagnósticos para análise futura
		for _, diag := range diagnostics {
			log.Printf("Coherence issue detected: %s", diag)
		}
	}

	state := len(diagnostics) == 0

	if !state {
		sData.ServiceID = 0
		sData.ServiceName = ""

		return &model.FindServiceResponse{
			Success: false,
			Data:    sData,
			Error:   "coherence issues detected between request and response",
		}, nil
	}

	return &model.FindServiceResponse{
		Success: state,
		Data:    sData,
	}, nil
}

// analyzeCoherence performs lightweight checks to surface coherence issues
// between the incoming request and the produced service data. This is useful
// when consuming web/API outputs where format/content can drift.
func analyzeCoherence(req *model.FindServiceRequest, data *model.ServiceData) []string {
	issues := make([]string, 0, 4)
	if req == nil {
		issues = append(issues, "request is nil")
		return issues
	}

	if req.Intent == "" {
		issues = append(issues, "request.intent is empty")
	}

	if data == nil {
		issues = append(issues, "response data is nil")
		return issues
	}

	if data.ServiceID <= 0 {
		issues = append(issues, "response.service_id must be > 0")
	}

	if data.ServiceName == "" {
		issues = append(issues, "response.service_name is empty")
	}

	// Basic semantic check: ensure the intent appears related (very naive heuristic)
	// This is intentionally simple to avoid heavy dependencies.
	if req.Intent != "" && data.ServiceName != "" {
		lowerIntent := strings.ToLower(req.Intent)
		lowerName := strings.ToLower(data.ServiceName)
		if !strings.Contains(lowerIntent, "segur") && strings.Contains(lowerName, "segur") {
			issues = append(issues, "potential mismatch: intent may not relate to returned service name")
		}
	}

	return issues
}

// normalizeServicePair validates and corrects the (id,name) pair returned by the model
// against the canonical service registry. It returns the normalized id/name and an
// optional diagnostic string when a correction is applied.
func (c *Core) normalizeServicePair(id uint8, name string) (uint8, string, string) {
	services := c.PromptManager.GetServiceDefinitions()
	nameToID := make(map[string]uint8, len(services))
	idToName := make(map[uint8]string, len(services))

	for _, s := range services {
		idToName[uint8(s.ID)] = s.Name
		nameToID[s.Name] = uint8(s.ID)
	}

	trimmed := strings.TrimSpace(name)
	if trimmed != "" {
		// Exact name match (case-sensitive first)
		if correctID, ok := nameToID[trimmed]; ok {
			if correctID != id {
				return correctID, idToName[correctID],
					fmt.Sprintf("normalized: corrected service_id from %d to %d based on service_name", id, correctID)
			}
			// id matches name; ensure canonical casing/name
			return id, trimmed, ""
		}
		// Case-insensitive name match: find canonical name
		lower := strings.ToLower(trimmed)
		for canonName, canonID := range nameToID {
			if strings.ToLower(canonName) == lower {
				if canonID != id || canonName != trimmed {
					return canonID, canonName,
						fmt.Sprintf("normalized: corrected service to (id=%d,name=%q) based on case-insensitive name match", canonID, canonName)
				}
				return id, canonName, ""
			}
		}
	}

	// If name didn't help, try id
	if canonName, ok := idToName[id]; ok {
		if canonName != trimmed {
			return id, canonName,
				fmt.Sprintf("normalized: corrected service_name from %q to %q based on service_id", trimmed, canonName)
		}
		return id, canonName, ""
	}

	// Unknown pair: fall back
	fallback := c.PromptManager.GetFallbackService()
	return uint8(fallback.ID), fallback.Name,
		fmt.Sprintf("normalized: unknown service pair (id=%d,name=%q), falling back to %q", id, name, fallback.Name)
}
