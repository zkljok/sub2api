package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type modelPlazaRepository struct{ db *sql.DB }

func NewModelPlazaRepository(db *sql.DB) service.ModelPlazaRepository {
	return &modelPlazaRepository{db: db}
}

func (r *modelPlazaRepository) ListVendors(ctx context.Context, includeDisabled bool) ([]service.ModelPlazaVendor, error) {
	query := `SELECT id, name, description, icon, status, sort_order, created_at, updated_at FROM model_plaza_vendors`
	if !includeDisabled {
		query += ` WHERE status = 'active'`
	}
	query += ` ORDER BY sort_order, name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list model plaza vendors: %w", err)
	}
	defer rows.Close()
	result := []service.ModelPlazaVendor{}
	for rows.Next() {
		var row service.ModelPlazaVendor
		if err := rows.Scan(&row.ID, &row.Name, &row.Description, &row.Icon, &row.Status, &row.SortOrder, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *modelPlazaRepository) CreateVendor(ctx context.Context, vendor *service.ModelPlazaVendor) error {
	return r.db.QueryRowContext(ctx, `INSERT INTO model_plaza_vendors (name, description, icon, status, sort_order) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at`,
		vendor.Name, vendor.Description, vendor.Icon, vendor.Status, vendor.SortOrder).Scan(&vendor.ID, &vendor.CreatedAt, &vendor.UpdatedAt)
}

func (r *modelPlazaRepository) UpdateVendor(ctx context.Context, vendor *service.ModelPlazaVendor) error {
	result, err := r.db.ExecContext(ctx, `UPDATE model_plaza_vendors SET name=$1, description=$2, icon=$3, status=$4, sort_order=$5, updated_at=NOW() WHERE id=$6`,
		vendor.Name, vendor.Description, vendor.Icon, vendor.Status, vendor.SortOrder, vendor.ID)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *modelPlazaRepository) DeleteVendor(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM model_plaza_vendors WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *modelPlazaRepository) ListModels(ctx context.Context, includeDisabled bool) ([]service.ModelPlazaModel, error) {
	query := `SELECT id, model_name, display_name, description, icon, tags, vendor_id, endpoints, status, name_rule, sort_order, pricing_override, auto_synced, created_at, updated_at FROM model_plaza_models`
	if !includeDisabled {
		query += ` WHERE status = 'active'`
	}
	query += ` ORDER BY sort_order, model_name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list model plaza models: %w", err)
	}
	defer rows.Close()
	result := []service.ModelPlazaModel{}
	for rows.Next() {
		var row service.ModelPlazaModel
		var tags, endpoints, override []byte
		if err := rows.Scan(&row.ID, &row.ModelName, &row.DisplayName, &row.Description, &row.Icon, &tags, &row.VendorID, &endpoints, &row.Status, &row.NameRule, &row.SortOrder, &override, &row.AutoSynced, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(tags, &row.Tags)
		_ = json.Unmarshal(endpoints, &row.Endpoints)
		_ = json.Unmarshal(override, &row.PricingOverride)
		if row.Tags == nil {
			row.Tags = []string{}
		}
		if row.Endpoints == nil {
			row.Endpoints = []string{}
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *modelPlazaRepository) CreateModel(ctx context.Context, model *service.ModelPlazaModel) error {
	tags, endpoints, override, err := marshalModelPlazaModelJSON(model)
	if err != nil {
		return err
	}
	return r.db.QueryRowContext(ctx, `INSERT INTO model_plaza_models (model_name, display_name, description, icon, tags, vendor_id, endpoints, status, name_rule, sort_order, pricing_override, auto_synced)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING id, created_at, updated_at`,
		model.ModelName, model.DisplayName, model.Description, model.Icon, tags, model.VendorID, endpoints, model.Status, model.NameRule, model.SortOrder, override, model.AutoSynced,
	).Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt)
}

func (r *modelPlazaRepository) UpdateModel(ctx context.Context, model *service.ModelPlazaModel) error {
	tags, endpoints, override, err := marshalModelPlazaModelJSON(model)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `UPDATE model_plaza_models SET model_name=$1, display_name=$2, description=$3, icon=$4, tags=$5, vendor_id=$6, endpoints=$7, status=$8, name_rule=$9, sort_order=$10, pricing_override=$11, auto_synced=$12, updated_at=NOW() WHERE id=$13`,
		model.ModelName, model.DisplayName, model.Description, model.Icon, tags, model.VendorID, endpoints, model.Status, model.NameRule, model.SortOrder, override, model.AutoSynced, model.ID)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *modelPlazaRepository) DeleteModel(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM model_plaza_models WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if changed, _ := result.RowsAffected(); changed == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *modelPlazaRepository) InsertMissingModels(ctx context.Context, models []string) (int, error) {
	inserted := 0
	for _, name := range models {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		result, err := r.db.ExecContext(ctx, `INSERT INTO model_plaza_models (model_name, display_name, auto_synced) VALUES ($1, $1, TRUE) ON CONFLICT (model_name, name_rule) DO NOTHING`, name)
		if err != nil {
			return inserted, err
		}
		if changed, _ := result.RowsAffected(); changed > 0 {
			inserted++
		}
	}
	return inserted, nil
}

func marshalModelPlazaModelJSON(model *service.ModelPlazaModel) ([]byte, []byte, []byte, error) {
	tags, err := json.Marshal(model.Tags)
	if err != nil {
		return nil, nil, nil, err
	}
	endpoints, err := json.Marshal(model.Endpoints)
	if err != nil {
		return nil, nil, nil, err
	}
	override, err := json.Marshal(model.PricingOverride)
	if err != nil {
		return nil, nil, nil, err
	}
	return tags, endpoints, override, nil
}
