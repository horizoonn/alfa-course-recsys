package catalog

import (
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

type labelRule struct {
	source  string
	id      uuid.UUID
	kind    domain.MaterialKind
	feature domain.FeatureName
	value   string
}

type ReviewedRule struct {
	Key       domain.MaterialKey
	TitleHash string
	Fits      []domain.TaskFit
}

var labelRules = []labelRule{
	{
		source:  "Topics",
		id:      uuid.MustParse("7864ae13-5634-4426-a136-09c32e84de91"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "first_steps",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("3aa8b98e-d50e-4f2a-a684-b287ddb45e0f"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "first_steps",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("de8477f7-a5f9-4bd3-b08a-f462befe1975"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "first_steps",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("4a88cbb5-01af-4d4e-8372-75e03aadba16"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "finance",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("4a88cbb5-01af-4d4e-8372-75e03aadba16"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "accounting",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("bfb81716-1041-4f7f-8ad2-f7eb5a1300a8"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "finance",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("eaf862c8-e07d-4de0-9d0e-c1e8175ab2a1"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "accounting",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("5fbd94ce-7ffa-4915-83df-0eb8fcb7930d"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "team",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("1fbcfe68-06cf-4760-ab57-b9b130c31b9e"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "promotion",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("1fbcfe68-06cf-4760-ab57-b9b130c31b9e"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "attract_clients",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("483045c7-218f-4d6e-9e1f-ebeb852b538c"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "sales",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("7eb7e98f-69a4-480a-b754-0c078daf7c50"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "sales",
	},
	{
		source:  "Topics",
		id:      uuid.MustParse("eca5d7cc-321c-451d-930c-784540d105a0"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "export",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("7c694604-9953-4f90-92ef-bbb0ea559682"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "first_steps",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("6bd69142-8d73-4a07-8b5b-550e1f264f04"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "first_steps",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("e410fe16-f016-45d4-b062-78a6cc6575f9"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "first_steps",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("9137d741-6bf7-4542-adcf-9c9b5592f747"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "accounting",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("16c416a3-a720-496b-8b09-79bb9a75b362"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "promotion",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("16c416a3-a720-496b-8b09-79bb9a75b362"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "attract_clients",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("fb0dfb3d-b547-4747-b625-d3da8e527db2"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "sales",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("aa2bbf41-800f-46e1-a1f6-08f88604e80b"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "sales",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("aa2bbf41-800f-46e1-a1f6-08f88604e80b"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "attract_clients",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("2f3b2c1b-3131-423d-8661-e9fa842420cb"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "finance",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("a04b5fe8-04ed-41cd-8cb1-2ffbd89ed01c"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "team",
	},
	{
		source:  "Tags",
		id:      uuid.MustParse("230e35b7-22ae-4a54-9084-39d58d6fbe9a"),
		kind:    "",
		feature: domain.FeatureTaskFit,
		value:   "team",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("0d65fe33-f7f4-4255-9281-56ad99726253"),
		kind:    "article",
		feature: domain.FeatureTaskFit,
		value:   "finance",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("e0aa7ae2-f56d-48d6-8698-28a5be72be33"),
		kind:    "article",
		feature: domain.FeatureTaskFit,
		value:   "accounting",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("005becd4-8bc7-487e-acd7-14f705173555"),
		kind:    "article",
		feature: domain.FeatureTaskFit,
		value:   "team",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("dfbc2d39-3a10-47b2-b59d-6e9d57687be8"),
		kind:    "article",
		feature: domain.FeatureTaskFit,
		value:   "promotion",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("dfbc2d39-3a10-47b2-b59d-6e9d57687be8"),
		kind:    "article",
		feature: domain.FeatureTaskFit,
		value:   "attract_clients",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("63424bc3-0767-427e-bbdb-b11a725571ab"),
		kind:    "article",
		feature: domain.FeatureTaskFit,
		value:   "sales",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("6e1771b3-204d-4aac-bcbb-aa9251d3d24c"),
		kind:    "document",
		feature: domain.FeatureTaskFit,
		value:   "finance",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("97f35e44-94dc-41dc-89ad-e0929e006ad1"),
		kind:    "document",
		feature: domain.FeatureTaskFit,
		value:   "accounting",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("92b16f7d-87a7-4906-ad5a-ab9a6b066d01"),
		kind:    "document",
		feature: domain.FeatureTaskFit,
		value:   "accounting",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("df88a4ac-aaad-4276-845f-01214385b67f"),
		kind:    "document",
		feature: domain.FeatureTaskFit,
		value:   "team",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("d25b8808-e453-4781-afb6-39bf47cf4af5"),
		kind:    "document",
		feature: domain.FeatureTaskFit,
		value:   "team",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("11c53f05-706d-4228-8b61-8f76d7f1e20d"),
		kind:    "document",
		feature: domain.FeatureTaskFit,
		value:   "first_steps",
	},
	{
		source:  "Metadata.AdditionalLabels",
		id:      uuid.MustParse("7934e04c-5831-4fff-9feb-483b464d6d92"),
		kind:    "",
		feature: domain.FeatureIndustry,
		value:   "services",
	},
	{
		source:  "Metadata.AdditionalLabels",
		id:      uuid.MustParse("146751ed-d436-4124-a7f8-109a71feb002"),
		kind:    "",
		feature: domain.FeatureIndustry,
		value:   "retail",
	},
	{
		source:  "Metadata.AdditionalLabels",
		id:      uuid.MustParse("07776a5e-37e2-4487-b3fa-a14bd51951f6"),
		kind:    "",
		feature: domain.FeatureIndustry,
		value:   "food",
	},
	{
		source:  "Metadata.AdditionalLabels",
		id:      uuid.MustParse("67133f7b-590c-4080-a916-03c147f59155"),
		kind:    "",
		feature: domain.FeatureIndustry,
		value:   "production",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("db9cd40f-a26d-44ea-8feb-5ebe727485dc"),
		kind:    "article",
		feature: domain.FeatureIndustry,
		value:   "services",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("c2a4dfd7-c086-4ccb-bf97-084c5ece2b3b"),
		kind:    "article",
		feature: domain.FeatureIndustry,
		value:   "retail",
	},
	{
		source:  "Metadata.Category",
		id:      uuid.MustParse("be3bb5c9-4360-481d-9171-cd376ee35d06"),
		kind:    "article",
		feature: domain.FeatureIndustry,
		value:   "food",
	},
}
