package genetics

import (
	"fmt"
	"strings"
)

// FeatureSet represents a collection of genetic features
type FeatureSet []Feature

// MergeWith объединяет текущий FeatureSet с другим и возвращает новый набор
func (fs FeatureSet) MergeWith(other FeatureSet) FeatureSet {
	result := make(FeatureSet, len(fs)+len(other))
	copy(result, fs)
	copy(result[len(fs):], other)
	return result
}

// BuildLLMContext creates a formatted context string from a set of Features
// that can be sent to an LLM for interpretation
func (fs FeatureSet) BuildLLMContext() string {
	var builder strings.Builder

	builder.WriteString("# Genetic Analysis Data\n\n")

	// Process each feature
	for i, feature := range fs {
		builder.WriteString(fmt.Sprintf("## Feature %d: %s\n\n", i+1, feature.Name))

		// Add genes information
		if len(feature.Genes) > 0 {
			builder.WriteString("### Genes:\n")
			for _, gene := range feature.Genes {
				builder.WriteString(fmt.Sprintf("#### %s\n", gene.Name))

				if len(gene.Interpretations) > 0 {
					for _, interp := range gene.Interpretations {
						builder.WriteString(fmt.Sprintf("- %s\n", interp))
					}
				}
				builder.WriteString("\n")
			}
		}

		// Add conclusions
		if len(feature.Conclusions) > 0 {
			builder.WriteString("### Conclusions:\n")
			for _, conclusion := range feature.Conclusions {
				builder.WriteString(fmt.Sprintf("- %s\n", conclusion))
			}
			builder.WriteString("\n")
		}

		// Add nutritional recommendations
		if len(feature.Nutrition) > 0 {
			builder.WriteString("### Nutrition:\n")
			for _, nutrition := range feature.Nutrition {
				builder.WriteString(fmt.Sprintf("- %s\n", nutrition))
			}
			builder.WriteString("\n")
		}

		// Add additional recommendations
		if len(feature.Additional) > 0 {
			builder.WriteString("### Additional:\n")
			for _, additional := range feature.Additional {
				builder.WriteString(fmt.Sprintf("- %s\n", additional))
			}
			builder.WriteString("\n")
		}

		// Add checklist items
		if len(feature.Checklist) > 0 {
			builder.WriteString("### Checklist:\n")
			for _, item := range feature.Checklist {
				builder.WriteString(fmt.Sprintf("- %s\n", item))
			}
			builder.WriteString("\n")
		}

		// Add clear separator between features
		builder.WriteString("\n===========================================\n\n")
	}

	return builder.String()
}
