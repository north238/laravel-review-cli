package review

import (
	"encoding/json"
	"fmt"
	"os"
)

type Findings struct {
	Content []Content `json:"findings"`
}

type Content struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Category    string `json:"category"`
	Confidence  string `json:"confidence"`
	Message     string `json:"message"`
	CodeSnippet string `json:"code_snippet"`
}

func ParseFindings(content string, aspect Aspect) ([]Finding, error) {
	var responseFindings Findings
	err := json.Unmarshal([]byte(content), &responseFindings)
	if err != nil {
		return nil, fmt.Errorf("【ERROR】failed to json parse: %w", err)
	}

	// []Contentを[]Findingに変換
	findings := make([]Finding, 0)
	for _, finding := range responseFindings.Content {
		// confidenceのフォールバック処理
		var confidence Confidence = Confidence(finding.Confidence)
		switch confidence {
		case ConfidenceHigh, ConfidenceMedium, ConfidenceLow:
			// 値をそのまま返却（何もしない）
		default:
			confidence = ConfidenceMedium
			fmt.Fprintf(os.Stderr, "【WARNING】invalid to confidence value: %s\n", finding.Confidence)
		}

		findings = append(findings, Finding{
			File:        finding.File,
			Line:        finding.Line,
			Aspect:      aspect,
			Category:    finding.Category,
			Confidence:  confidence,
			Message:     finding.Message,
			CodeSnippet: finding.CodeSnippet,
		})
	}

	return findings, nil
}
