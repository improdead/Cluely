package answer

import (
	"log"
	"strings"
)

type Answer struct {
	Answer     string  `json:"answer"`
	FollowUp   string  `json:"followUp"`
	Confidence float64 `json:"confidence,omitempty"`
}

type Service struct{}

func NewServiceFromEnv() *Service {
	return &Service{}
}

func (s *Service) Micro(text string, ocr []string, firstOCR []string, lastOCR []string) *Answer {
	transcript := strings.TrimSpace(text)
	if transcript == "" {
		log.Println("[answer] empty text, skipping")
		return nil
	}

	contextTokens := contextualTokens(ocr, firstOCR, lastOCR)
	answer := craftAnswer(transcript, contextTokens)
	followUp := craftFollowUp(transcript, contextTokens)

	if answer == "" {
		answer = "Reflect key concern and reassure next steps"
	}
	if followUp == "" {
		followUp = "Ask: 'What would make you comfortable to proceed?'"
	}

	return &Answer{
		Answer:     answer,
		FollowUp:   followUp,
		Confidence: 0.4,
	}
}

func contextualTokens(ocr []string, firstOCR []string, lastOCR []string) []string {
	merged := append([]string{}, ocr...)
	merged = append(merged, firstOCR...)
	merged = append(merged, lastOCR...)

	seen := make(map[string]struct{}, len(merged))
	uniq := make([]string, 0, len(merged))
	for _, token := range merged {
		token = strings.TrimSpace(strings.ToLower(token))
		if token == "" {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		uniq = append(uniq, token)
	}
	return uniq
}

func craftAnswer(transcript string, tokens []string) string {
	lower := strings.ToLower(transcript)

	if strings.Contains(lower, "budget") {
		return "Highlight fiscal upside and ask who approves budget"
	}
	if strings.Contains(lower, "timeline") || strings.Contains(lower, "schedule") {
		return "Lock a concrete timeline before momentum fades"
	}
	if strings.Contains(lower, "concern") || strings.Contains(lower, "worried") {
		return "Acknowledge concern and offer next step to de-risk"
	}
	if strings.Contains(lower, "not sure") || strings.Contains(lower, "confused") {
		return "Clarify the core value prop in plain language"
	}

	for _, token := range tokens {
		switch token {
		case "revenue", "forecast", "q4":
			return "Frame the revenue story around leading indicators"
		case "architecture", "design", "diagram":
			return "Translate the architecture into executive outcomes"
		case "hiring", "headcount":
			return "Clarify hiring impact on roadmap commitments"
		}
	}

	return "Reinforce their goal and propose a decisive next step"
}

func craftFollowUp(transcript string, tokens []string) string {
	lower := strings.ToLower(transcript)

	if strings.Contains(lower, "budget") {
		return "Ask: 'Who signs off on the numbers?'"
	}
	if strings.Contains(lower, "timeline") {
		return "Ask: 'What deadline should we plan against?'"
	}
	if strings.Contains(lower, "risk") {
		return "Ask: 'Which risk matters most right now?'"
	}
	if strings.Contains(lower, "decision") {
		return "Ask: 'What do you need to decide today?'"
	}

	for _, token := range tokens {
		switch token {
		case "roadmap":
			return "Ask: 'Which milestone is most critical?'"
		case "contract":
			return "Ask: 'Any blockers before we finalize?'"
		case "metrics":
			return "Ask: 'Which metric should we steer toward?'"
		}
	}

	return "Ask: 'What would help you move forward?'"
}
