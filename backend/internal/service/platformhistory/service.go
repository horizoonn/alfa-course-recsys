package platformhistory

import (
	"context"
	"fmt"
	"slices"
	"uuid"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/domain"
)

type Repository interface {
	Load(context.Context) ([]domain.LearningRecord, []domain.RatingRecord, error)
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

type aggregate struct {
	history                          domain.History
	completed                        map[domain.MaterialKey]bool
	learning                         map[domain.MaterialKey]bool
	positiveRatings, negativeRatings int
}

var entityKinds = map[uuid.UUID]domain.MaterialKind{
	uuid.MustParse("8ede4179-0faf-407d-a4fc-1d94f9a63089"): domain.MaterialKindArticle,
	uuid.MustParse("8b440b1a-00c3-423a-9bae-3376475eb203"): domain.MaterialKindCourse,
	uuid.MustParse("da1ad865-8605-4ab1-9f6c-4489a26fbc35"): domain.MaterialKindDocument,
	uuid.MustParse("ece4acf3-5dd3-47b9-9f08-99c0d6203cc7"): domain.MaterialKindWebinar,
	uuid.MustParse("142c23e2-bf71-4a73-b12e-6b1b5e4d47cc"): domain.MaterialKindPodcast,
	uuid.MustParse("7ac1c8c3-18b8-42ea-92ea-fcc10a6d496f"): domain.MaterialKindCalculator,
	uuid.MustParse("72dfaf58-4e43-4b92-a1cd-5f8b5de739b4"): domain.MaterialKindCalculator,
}

func (s *Service) Load(ctx context.Context, materials []domain.Material) ([]domain.HistoryProfile, error) {
	learning, ratings, err := s.repository.Load(ctx)
	if err != nil {
		return nil, err
	}
	allowed := map[domain.MaterialKey]bool{}
	lessonCourses := map[uuid.UUID][]domain.MaterialKey{}
	for _, material := range materials {
		if material.Visibility != domain.VisibilityListed {
			continue
		}
		allowed[material.Key] = true
		if material.Metadata.Course != nil {
			for _, lesson := range material.Metadata.Course.Program {
				if lesson.Visibility == domain.VisibilityListed {
					lessonCourses[lesson.ID] = append(lessonCourses[lesson.ID], material.Key)
				}
			}
		}
	}
	users := map[uuid.UUID]*aggregate{}
	get := func(id uuid.UUID) *aggregate {
		if users[id] == nil {
			users[id] = &aggregate{completed: map[domain.MaterialKey]bool{}, learning: map[domain.MaterialKey]bool{}}
		}
		return users[id]
	}
	for _, record := range learning {
		key := domain.MaterialKey{Kind: domain.MaterialKindCourse, ID: record.CourseID}
		if !allowed[key] || record.UserID == uuid.Nil() {
			continue
		}
		user := get(record.UserID)
		if record.Completed {
			user.completed[key] = true
		} else if record.Started || record.LessonsCompleted > 0 {
			user.learning[key] = true
		}
	}
	for _, record := range ratings {
		kind, known := entityKinds[record.EntityTypeID]
		if !known || record.UserID == uuid.Nil() {
			continue
		}
		keys := []domain.MaterialKey{{Kind: kind, ID: record.EntityID}}
		lessonRating := false
		if kind == domain.MaterialKindCourse && !allowed[keys[0]] {
			keys = lessonCourses[record.EntityID]
			lessonRating = true
		}
		for _, key := range keys {
			if !allowed[key] {
				continue
			}
			user := get(record.UserID)
			if record.Value >= 4 {
				user.history.Positive = append(user.history.Positive, key)
				user.positiveRatings++
			}
			if record.Value <= 2 && !lessonRating {
				user.history.Dismissed = append(user.history.Dismissed, key)
				user.negativeRatings++
			}
		}
	}
	for _, user := range users {
		for key := range user.completed {
			user.history.Completed = append(user.history.Completed, key)
			delete(user.learning, key)
		}
		for key := range user.learning {
			user.history.Positive = append(user.history.Positive, key)
		}
		user.history.Positive = slices.DeleteFunc(user.history.Positive, func(key domain.MaterialKey) bool {
			return slices.Contains(user.history.Dismissed, key)
		})
		for _, keys := range []*[]domain.MaterialKey{&user.history.Completed, &user.history.Positive, &user.history.Dismissed} {
			slices.SortFunc(*keys, func(a, b domain.MaterialKey) int {
				if a.String() < b.String() {
					return -1
				}
				if a.String() > b.String() {
					return 1
				}
				return 0
			})
			*keys = slices.Compact(*keys)
		}
	}
	profiles := []domain.HistoryProfile{}
	selected := map[uuid.UUID]bool{}
	for index, category := range []string{"completed", "learning", "ratings"} {
		score := func(user *aggregate) int {
			if category == "completed" {
				return len(user.completed)
			}
			if category == "learning" {
				return len(user.learning)
			}
			return user.positiveRatings + user.negativeRatings
		}
		ids := []uuid.UUID{}
		for id, user := range users {
			if !selected[id] && score(user) > 0 {
				ids = append(ids, id)
			}
		}
		slices.SortFunc(ids, func(a, b uuid.UUID) int {
			if diff := score(users[b]) - score(users[a]); diff != 0 {
				return diff
			}
			if a.String() < b.String() {
				return -1
			}
			if a.String() > b.String() {
				return 1
			}
			return 0
		})
		if len(ids) == 0 {
			continue
		}
		id := ids[0]
		selected[id] = true
		user := users[id]
		title := map[string]string{"completed": "История завершённых курсов", "learning": "История начатого обучения", "ratings": "История оценок материалов"}[category]
		profiles = append(profiles, domain.HistoryProfile{ID: fmt.Sprintf("H%02d", index+1), Title: title,
			Description:      fmt.Sprintf("Обезличенная история из БД Альфа-Курса: завершено курсов — %d, начато — %d, оценок 4–5 — %d, оценок 1–2 — %d.", len(user.completed), len(user.learning), user.positiveRatings, user.negativeRatings),
			CompletedCourses: len(user.completed), LearningCourses: len(user.learning), PositiveRatings: user.positiveRatings, NegativeRatings: user.negativeRatings, History: user.history})
	}
	return profiles, nil
}
