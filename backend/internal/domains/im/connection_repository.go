package im

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

type ConnectionListFilters struct {
	Platform string
	Status   string
	Keyword  string
}

type ConnectionRepository struct{}

var defaultConnectionRepository = &ConnectionRepository{}

func (r *ConnectionRepository) Get(ctx context.Context, enterpriseID, id string) (*connectionRecord, error) {
	var row connectionRecord
	if err := g.DB().Model("im_connections").Ctx(ctx).Where("enterprise_id = ? AND id = ?", enterpriseID, id).Scan(&row); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, nil
	}
	return &row, nil
}

func (r *ConnectionRepository) GetByID(ctx context.Context, id string) (*connectionRecord, error) {
	var row connectionRecord
	if err := g.DB().Model("im_connections").Ctx(ctx).Where("id = ?", id).Scan(&row); err != nil {
		return nil, err
	}
	if row.ID == "" {
		return nil, nil
	}
	return &row, nil
}

func (r *ConnectionRepository) Exists(ctx context.Context, enterpriseID, id string) (bool, error) {
	count, err := g.DB().Model("im_connections").Ctx(ctx).Where("enterprise_id = ? AND id = ?", enterpriseID, id).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ConnectionRepository) List(ctx context.Context, enterpriseID string, filters ConnectionListFilters) ([]connectionRecord, error) {
	query := g.DB().Model("im_connections").Ctx(ctx).Where("enterprise_id = ?", enterpriseID)
	if platform := strings.TrimSpace(filters.Platform); platform != "" {
		query = query.Where("platform = ?", platform)
	}
	if status := strings.TrimSpace(filters.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		query = query.Where("name ILIKE ?", "%"+keyword+"%")
	}

	var rows []connectionRecord
	if err := query.Order("created_at DESC").Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ConnectionRepository) Create(ctx context.Context, data g.Map) error {
	_, err := g.DB().Model("im_connections").Ctx(ctx).Data(data).Insert()
	return err
}

func (r *ConnectionRepository) Update(ctx context.Context, enterpriseID, id string, data g.Map) error {
	_, err := g.DB().Model("im_connections").Ctx(ctx).Data(data).Where("enterprise_id = ? AND id = ?", enterpriseID, id).Update()
	return err
}

func (r *ConnectionRepository) ListEvents(ctx context.Context, enterpriseID, connectionID, direction, status string, limit int) ([]eventRecord, error) {
	query := g.DB().Model("external_message_events").Ctx(ctx).Where("enterprise_id = ? AND connection_id = ?", enterpriseID, connectionID)
	if direction = strings.TrimSpace(direction); direction != "" {
		query = query.Where("direction = ?", direction)
	}
	if status = strings.TrimSpace(status); status != "" {
		query = query.Where("status = ?", status)
	}

	var rows []eventRecord
	if err := query.Order("created_at DESC").Limit(limit).Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ConnectionRepository) ListDeliveries(ctx context.Context, enterpriseID, connectionID, status string, limit int) ([]deliveryRecord, error) {
	query := g.DB().Model("channel_delivery_logs").Ctx(ctx).Where("enterprise_id = ? AND connection_id = ?", enterpriseID, connectionID)
	if status = strings.TrimSpace(status); status != "" {
		query = query.Where("status = ?", status)
	}

	var rows []deliveryRecord
	if err := query.Order("created_at DESC").Limit(limit).Scan(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ConnectionRepository) UpdateRuntimeFieldsByID(ctx context.Context, connectionID string, data g.Map) error {
	_, err := g.DB().Model("im_connections").Ctx(ctx).Data(data).Where("id = ?", connectionID).Update()
	return err
}
