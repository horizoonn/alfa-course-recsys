package domain

import "slices"

type Profile struct {
	TaskID     TaskID
	Stage      BusinessStage
	Industry   Industry
	Experience ExperienceLevel
	HelpMode   HelpMode
}

type Option struct {
	ID    string
	Label string
}
type ProfileOptions struct{ Tasks, Stages, Industries, Experience, HelpModes []Option }

type DemoProfile struct {
	ID          string
	Title       string
	Description string
	Profile     Profile
}

type FieldError struct {
	Field   string
	Message string
}
type InputError struct {
	Fields []FieldError
}

func (e *InputError) Error() string {
	return "invalid recommendation request"
}

type History struct {
	Completed []MaterialKey
	Positive  []MaterialKey
}

func (p Profile) Validate() error {
	options := ProfileMetadata()
	fields := []struct {
		name, value string
		options     []Option
	}{
		{"profile.task_id", string(p.TaskID), options.Tasks},
		{"profile.stage", string(p.Stage), options.Stages},
		{"profile.industry", string(p.Industry), options.Industries},
		{"profile.experience", string(p.Experience), options.Experience},
		{"profile.help_mode", string(p.HelpMode), options.HelpModes},
	}
	errors := []FieldError{}
	for _, field := range fields {
		if field.value != "" && !slices.ContainsFunc(field.options, func(o Option) bool { return o.ID == field.value }) {
			errors = append(errors, FieldError{field.name, "Неизвестный код."})
		}
	}
	if len(errors) > 0 {
		return &InputError{Fields: errors}
	}
	return nil
}

func ProfileMetadata() ProfileOptions {
	return ProfileOptions{
		Tasks: []Option{
			{ID: "first_steps", Label: "Первые шаги в бизнесе"},
			{ID: "attract_clients", Label: "Привлечь клиентов"},
			{ID: "promotion", Label: "Продвижение бизнеса"},
			{ID: "finance", Label: "Управление финансами"},
			{ID: "accounting", Label: "Разобраться в учёте"},
			{ID: "sales", Label: "Увеличить продажи"},
			{ID: "team", Label: "Найм и команда"},
			{ID: "export", Label: "Выход на зарубежные рынки"},
		},
		Stages: []Option{
			{ID: "idea", Label: "Только идея"},
			{ID: "starting", Label: "Планирую открыть"},
			{ID: "running", Label: "Бизнес уже работает"},
			{ID: "growing", Label: "Развиваю и расту"},
		},
		Industries: []Option{
			{ID: "services", Label: "Услуги"},
			{ID: "retail", Label: "Торговля"},
			{ID: "it", Label: "IT и онлайн"},
			{ID: "food", Label: "Общепит"},
			{ID: "production", Label: "Производство"},
		},
		Experience: []Option{
			{ID: "beginner", Label: "Начальный"},
			{ID: "intermediate", Label: "Средний"},
			{ID: "advanced", Label: "Продвинутый"},
		},
		HelpModes: []Option{
			{ID: "quick_answer", Label: "Быстрый ответ"},
			{ID: "tool", Label: "Практический инструмент"},
			{ID: "deep_learning", Label: "Подробное обучение"},
		},
	}
}

func DemoProfiles() []DemoProfile {
	return []DemoProfile{
		{
			ID:          "P01",
			Title:       "P01 — начинающий",
			Description: "Открывает бизнес в сфере услуг. Нужен первый шаг.",
			Profile: Profile{
				Stage:      "starting",
				Industry:   "services",
				Experience: "beginner",
				TaskID:     "first_steps",
				HelpMode:   "quick_answer",
			}},
		{
			ID:          "P02",
			Title:       "P02 — микробизнес, маркетинг",
			Description: "Ведёт микробизнес в услугах. Хочет привлечь клиентов.",
			Profile: Profile{
				Stage:      "running",
				Industry:   "services",
				Experience: "beginner",
				TaskID:     "attract_clients",
				HelpMode:   "quick_answer",
			}},
		{
			ID:          "P03",
			Title:       "P03 — как P02, но инструмент",
			Description: "Профиль P02; нужен практический инструмент.",
			Profile: Profile{
				Stage:      "running",
				Industry:   "services",
				Experience: "beginner",
				TaskID:     "attract_clients",
				HelpMode:   "tool",
			}},
		{
			ID:          "P04",
			Title:       "P04 — как P02, но обучение",
			Description: "Профиль P02; нужно подробное обучение.",
			Profile: Profile{
				Stage:      "running",
				Industry:   "services",
				Experience: "beginner",
				TaskID:     "attract_clients",
				HelpMode:   "deep_learning",
			}},
		{
			ID:          "P05",
			Title:       "P05 — соло-фаундер, финансы",
			Description: "Развивает бизнес в услугах. Хочет разобраться в финансах.",
			Profile: Profile{
				Stage:      "growing",
				Industry:   "services",
				Experience: "intermediate",
				TaskID:     "finance",
				HelpMode:   "deep_learning",
			}},
		{
			ID:          "P06",
			Title:       "P06 — торговля, учёт",
			Description: "Ведёт торговый бизнес. Нужна помощь с учётом.",
			Profile: Profile{
				Stage:      "running",
				Industry:   "retail",
				Experience: "beginner",
				TaskID:     "accounting",
				HelpMode:   "quick_answer",
			}},
		{
			ID:          "P07",
			Title:       "P07 — пустой профиль",
			Description: "Пустой профиль для проверки холодного старта.",
			Profile:     Profile{},
		},
		{
			ID:          "P08",
			Title:       "P08 — только задача",
			Description: "Известна только задача — продвижение бизнеса.",
			Profile: Profile{
				TaskID: "promotion",
			}},
		{
			ID:          "P09",
			Title:       "P09 — зарубежные рынки",
			Description: "Подбор по задаче выхода на зарубежные рынки.",
			Profile: Profile{
				TaskID: "export",
			}},
		{
			ID:          "P10",
			Title:       "P10 — проверка истории",
			Description: "Профиль P02 для проверки завершения и положительного события на синтетической истории.",
			Profile: Profile{
				Stage:      "running",
				Industry:   "services",
				Experience: "beginner",
				TaskID:     "attract_clients",
				HelpMode:   "quick_answer",
			}},
	}
}
