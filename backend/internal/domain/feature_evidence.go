package domain

type FeatureName string

const (
	FeatureTaskFit  FeatureName = "task_fit"
	FeatureIndustry FeatureName = "industry"
	FeatureHelpMode FeatureName = "help_mode"
)

type FeatureEvidence struct {
	Feature FeatureName
	Value   string
	Match   TaskMatch
	Source  string
	Rule    string
	Method  string
	Version string
}
