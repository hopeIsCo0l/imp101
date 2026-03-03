package services

import (
	"math"
	"regexp"
	"sort"
	"strings"
)

var tokenRegex = regexp.MustCompile(`[a-zA-Z0-9_]+`)

func normalizeTokens(text string) []string {
	parts := tokenRegex.FindAllString(strings.ToLower(text), -1)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if len(p) > 2 {
			out = append(out, p)
		}
	}
	return out
}

func frequency(tokens []string) map[string]float64 {
	m := make(map[string]float64)
	if len(tokens) == 0 {
		return m
	}
	for _, t := range tokens {
		m[t] += 1
	}
	total := float64(len(tokens))
	for k := range m {
		m[k] = m[k] / total
	}
	return m
}

func cosineScore(a, b map[string]float64) float64 {
	var dot, na, nb float64
	seen := map[string]struct{}{}
	for k, av := range a {
		seen[k] = struct{}{}
		dot += av * b[k]
		na += av * av
	}
	for k, bv := range b {
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
		}
		nb += bv * bv
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func ComputeTFIDFApproximation(cvText, jdText, requiredSkills string) CVAnalysis {
	cvTokens := normalizeTokens(cvText)
	jdTokens := normalizeTokens(jdText + " " + requiredSkills)
	cvFreq := frequency(cvTokens)
	jdFreq := frequency(jdTokens)
	score := cosineScore(cvFreq, jdFreq) * 100

	type pair struct {
		term  string
		score float64
	}
	var overlaps []pair
	for term, cvW := range cvFreq {
		jdW := jdFreq[term]
		contrib := cvW * jdW
		if contrib > 0 {
			overlaps = append(overlaps, pair{term: term, score: contrib})
		}
	}
	sort.Slice(overlaps, func(i, j int) bool {
		return overlaps[i].score > overlaps[j].score
	})

	topSkills := make([]string, 0, 10)
	commonTerms := make([]string, 0, 10)
	for i := 0; i < len(overlaps) && i < 10; i++ {
		topSkills = append(topSkills, overlaps[i].term)
		commonTerms = append(commonTerms, overlaps[i].term)
	}

	explanation := "Low match"
	if score >= 80 {
		explanation = "High match based on strong overlap in key terms."
	} else if score >= 60 {
		explanation = "Moderate match with several relevant terms."
	}

	rawText := cvText
	if len(rawText) > 4000 {
		rawText = rawText[:4000]
	}

	return CVAnalysis{
		Score:       math.Round(score*100) / 100,
		TopSkills:   topSkills,
		CommonTerms: commonTerms,
		Explanation: explanation,
		RawText:     rawText,
	}
}
