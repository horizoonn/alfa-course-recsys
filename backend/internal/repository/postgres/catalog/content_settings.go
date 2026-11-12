package catalog

import (
	"context"
	"fmt"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/horizoonn/alfa-course-recsys/backend/internal/repository/postgres/model"
)

type contentRelations struct {
	topicsBySettings     map[uuid.UUID][]model.ContentTopic
	tagsBySettings       map[uuid.UUID][]model.ContentTag
	industriesBySettings map[uuid.UUID][]model.ContentIndustry
}

func (r *Repository) loadContentRelations(
	ctx context.Context,
	settingsIDs []uuid.UUID,
) (contentRelations, error) {
	if len(settingsIDs) == 0 {
		return contentRelations{}, nil
	}

	topics, err := r.loadTopics(ctx, settingsIDs)
	if err != nil {
		return contentRelations{}, err
	}
	tags, err := r.loadTags(ctx, settingsIDs)
	if err != nil {
		return contentRelations{}, err
	}
	industries, err := r.loadIndustries(ctx, settingsIDs)
	if err != nil {
		return contentRelations{}, err
	}

	relations := contentRelations{
		topicsBySettings:     make(map[uuid.UUID][]model.ContentTopic, len(settingsIDs)),
		tagsBySettings:       make(map[uuid.UUID][]model.ContentTag, len(settingsIDs)),
		industriesBySettings: make(map[uuid.UUID][]model.ContentIndustry, len(settingsIDs)),
	}
	for _, topic := range topics {
		relations.topicsBySettings[topic.ContentSettingsID] = append(relations.topicsBySettings[topic.ContentSettingsID], topic)
	}
	for _, tag := range tags {
		relations.tagsBySettings[tag.ContentSettingsID] = append(relations.tagsBySettings[tag.ContentSettingsID], tag)
	}
	for _, industry := range industries {
		relations.industriesBySettings[industry.ContentSettingsID] = append(relations.industriesBySettings[industry.ContentSettingsID], industry)
	}
	return relations, nil
}

func (r *Repository) loadTopics(
	ctx context.Context,
	settingsIDs []uuid.UUID,
) ([]model.ContentTopic, error) {
	if len(settingsIDs) == 0 {
		return []model.ContentTopic{}, nil
	}

	const query = `
		SELECT
			binding."ContentSettingsId",
			binding."TopicId",
			topic."Title",
			topic."IsDeleted",
			topic."IsPublic",
			binding."Order"
		FROM public."ContentSettingsTopics" AS binding
		JOIN public."Topics" AS topic ON topic."Id" = binding."TopicId"
		WHERE binding."ContentSettingsId" = ANY($1::uuid[])
		ORDER BY binding."ContentSettingsId", binding."Order", binding."TopicId"
	`

	rows, err := r.executor(ctx).Query(ctx, query, settingsIDs)
	if err != nil {
		return nil, fmt.Errorf("query content topics: %w", err)
	}

	topics, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.ContentTopic, error) {
		var topic model.ContentTopic
		err := topic.Scan(row)
		return topic, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan content topics: %w", err)
	}

	return topics, nil
}

func (r *Repository) loadTags(
	ctx context.Context,
	settingsIDs []uuid.UUID,
) ([]model.ContentTag, error) {
	if len(settingsIDs) == 0 {
		return []model.ContentTag{}, nil
	}

	const query = `
		SELECT
			binding."ContentSettingsId",
			binding."TagId",
			tag."Title",
			tag."IsDeleted",
			binding."Order"
		FROM public."ContentSettingsTags" AS binding
		JOIN public."Tags" AS tag ON tag."Id" = binding."TagId"
		WHERE binding."ContentSettingsId" = ANY($1::uuid[])
		ORDER BY binding."ContentSettingsId", binding."Order", binding."TagId"
	`

	rows, err := r.executor(ctx).Query(ctx, query, settingsIDs)
	if err != nil {
		return nil, fmt.Errorf("query content tags: %w", err)
	}

	tags, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.ContentTag, error) {
		var tag model.ContentTag
		err := tag.Scan(row)
		return tag, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan content tags: %w", err)
	}

	return tags, nil
}

func (r *Repository) loadIndustries(
	ctx context.Context,
	settingsIDs []uuid.UUID,
) ([]model.ContentIndustry, error) {
	if len(settingsIDs) == 0 {
		return []model.ContentIndustry{}, nil
	}

	const query = `
		SELECT
			binding."ContentSettingsId",
			binding."IndustryId",
			industry."Title",
			industry."IsDeleted",
			binding."Order"
		FROM public."ContentSettingsIndustries" AS binding
		JOIN public."Industries" AS industry ON industry."Id" = binding."IndustryId"
		WHERE binding."ContentSettingsId" = ANY($1::uuid[])
		ORDER BY binding."ContentSettingsId", binding."Order", binding."IndustryId"
	`

	rows, err := r.executor(ctx).Query(ctx, query, settingsIDs)
	if err != nil {
		return nil, fmt.Errorf("query content industries: %w", err)
	}

	industries, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.ContentIndustry, error) {
		var industry model.ContentIndustry
		err := industry.Scan(row)
		return industry, err
	})
	if err != nil {
		return nil, fmt.Errorf("scan content industries: %w", err)
	}

	return industries, nil
}
