package domain

import "uuid"

type LearningRecord struct {
	UserID           uuid.UUID
	CourseID         uuid.UUID
	Completed        bool
	Started          bool
	LessonsCompleted int
}

type RatingRecord struct {
	UserID       uuid.UUID
	EntityID     uuid.UUID
	EntityTypeID uuid.UUID
	Value        int
}

type HistoryProfile struct {
	ID               string
	Title            string
	Description      string
	CompletedCourses int
	LearningCourses  int
	PositiveRatings  int
	NegativeRatings  int
	History          History
}
