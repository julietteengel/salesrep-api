package insights

import (
	"strings"
)

// CallTranscript represents a transcript segment (duplicate to avoid circular import)
type CallTranscript struct {
	ID         uint    `json:"id"`
	CallID     uint    `json:"call_id"`
	SpeakerID  string  `json:"speaker_id"`
	Text       string  `json:"text"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	Confidence float64 `json:"confidence"`
}

type IInsightsService interface {
	GenerateInsights(transcripts []*CallTranscript) *CallInsights
}

type InsightsService struct{}

type CallInsights struct {
	Summary   string   `json:"summary"`
	Sentiment string   `json:"sentiment"`
	Score     float64  `json:"score"`
	Keywords  []string `json:"keywords"`
}

func NewInsightsService() IInsightsService {
	return &InsightsService{}
}

// Generate basic insights from transcripts
func (s *InsightsService) GenerateInsights(transcripts []*CallTranscript) *CallInsights {
	if len(transcripts) == 0 {
		return &CallInsights{
			Summary:   "No transcript available",
			Sentiment: "neutral",
			Score:     0.0,
			Keywords:  []string{},
		}
	}

	// Combine all transcript text
	var allText string
	for _, transcript := range transcripts {
		allText += transcript.Text + " "
	}

	// Basic sentiment analysis (simple keyword-based for now)
	sentiment, score := s.analyzeSentiment(allText)
	
	// Extract keywords (simple implementation)
	keywords := s.extractKeywords(allText)

	// Generate summary (first 200 characters for now)
	summary := allText
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	return &CallInsights{
		Summary:   summary,
		Sentiment: sentiment,
		Score:     score,
		Keywords:  keywords,
	}
}

// Simple sentiment analysis
func (s *InsightsService) analyzeSentiment(text string) (string, float64) {
	positiveWords := []string{"great", "excellent", "good", "amazing", "fantastic", "perfect", "love", "happy", "yes", "absolutely"}
	negativeWords := []string{"bad", "terrible", "awful", "hate", "no", "never", "disappointed", "frustrated", "angry", "problem"}
	
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		if contains(text, word) {
			positiveCount++
		}
	}
	
	for _, word := range negativeWords {
		if contains(text, word) {
			negativeCount++
		}
	}
	
	if positiveCount > negativeCount {
		return "positive", 0.7
	} else if negativeCount > positiveCount {
		return "negative", -0.7
	}
	
	return "neutral", 0.0
}

// Extract keywords (simple implementation)
func (s *InsightsService) extractKeywords(text string) []string {
	commonKeywords := []string{"price", "cost", "budget", "timeline", "deadline", "meeting", "demo", "proposal", "contract", "deal"}
	var foundKeywords []string
	
	for _, keyword := range commonKeywords {
		if contains(text, keyword) {
			foundKeywords = append(foundKeywords, keyword)
		}
	}
	
	return foundKeywords
}

// Helper function to check if text contains a word (case-insensitive)
func contains(text, word string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(word))
}