package converter

import (
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

var (
	publishedStatusID = uuid.MustParse("61f12dbc-3cd4-4c14-b64b-1709548b904e")
	draftStatusID     = uuid.MustParse("eefe018b-487a-454b-a5ba-bcb094816e2c")

	openAccessTypeID         = uuid.MustParse("ed085df2-b9bf-43fc-9a29-e2140346e333")
	subscriptionAccessTypeID = uuid.MustParse("1069fd62-8920-417b-b6c1-1bfb7c3f57fb")
	paidAccessTypeID         = uuid.MustParse("9cda1148-8db6-4d29-98fc-6d67ea538fff")
)

func accessModel(sourceTypeID uuid.UUID, settingsTypeID *uuid.UUID) domain.AccessModel {
	if settingsTypeID == nil || sourceTypeID != *settingsTypeID {
		return domain.AccessModelUnknown
	}

	return accessTypeModel(sourceTypeID)
}

func accessTypeModel(typeID uuid.UUID) domain.AccessModel {
	switch typeID {
	case openAccessTypeID:
		return domain.AccessModelFree
	case subscriptionAccessTypeID:
		return domain.AccessModelSubscription
	case paidAccessTypeID:
		return domain.AccessModelPaid
	default:
		return domain.AccessModelUnknown
	}
}
