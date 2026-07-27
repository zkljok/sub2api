package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	PricingSourceModelPlazaOverride = "model_plaza_override"
	PricingSourceModelPlazaOfficial = "model_plaza_official"

	ModelPlazaBillingStatusApplied     = "applied"
	ModelPlazaBillingStatusDisplayOnly = "display_only"
	ModelPlazaBillingStatusUnpriced    = "unpriced"
)

type ModelPlazaBillingSettings struct {
	Enabled bool `json:"enabled"`
}

type ModelPlazaModelAdminView struct {
	ModelPlazaModel
	BillingStatus string `json:"billing_status"`
}

func DefaultModelPlazaBillingSettings() ModelPlazaBillingSettings {
	return ModelPlazaBillingSettings{Enabled: false}
}

func (s *ModelPlazaService) GetBillingSettings(ctx context.Context) (ModelPlazaBillingSettings, error) {
	settings := DefaultModelPlazaBillingSettings()
	if s == nil || s.settingRepo == nil {
		return settings, nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyModelPlazaBillingSettings)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return settings, nil
		}
		return settings, err
	}
	if strings.TrimSpace(raw) == "" {
		return settings, nil
	}
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return DefaultModelPlazaBillingSettings(), nil
	}
	return settings, nil
}

func (s *ModelPlazaService) UpdateBillingSettings(ctx context.Context, settings ModelPlazaBillingSettings) (ModelPlazaBillingSettings, error) {
	if s == nil || s.settingRepo == nil {
		return DefaultModelPlazaBillingSettings(), fmt.Errorf("setting repository is not configured")
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return DefaultModelPlazaBillingSettings(), err
	}
	if err := s.settingRepo.Set(ctx, SettingKeyModelPlazaBillingSettings, string(data)); err != nil {
		return DefaultModelPlazaBillingSettings(), err
	}
	return settings, nil
}

func (s *ModelPlazaService) ListModelsForAdmin(ctx context.Context, includeDisabled bool) ([]ModelPlazaModelAdminView, error) {
	models, err := s.ListModels(ctx, includeDisabled)
	if err != nil {
		return nil, err
	}
	settings, err := s.GetBillingSettings(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ModelPlazaModelAdminView, 0, len(models))
	for _, model := range models {
		result = append(result, ModelPlazaModelAdminView{
			ModelPlazaModel: model,
			BillingStatus:   modelPlazaBillingStatus(model, settings.Enabled),
		})
	}
	return result, nil
}

func (s *ModelPlazaService) ResolveBillingPricing(ctx context.Context, modelName string) (*ResolvedPricing, bool, error) {
	settings, err := s.GetBillingSettings(ctx)
	if err != nil || !settings.Enabled {
		return nil, false, err
	}
	models, err := s.repo.ListModels(ctx, false)
	if err != nil {
		return nil, false, err
	}
	for _, meta := range models {
		if !matchesModelPlazaRule(meta, modelName) {
			continue
		}
		if meta.PricingOverride.IsConfigured() {
			return resolvedPricingFromModelPlazaOverride(meta.PricingOverride, PricingSourceModelPlazaOverride), true, nil
		}
		if preset, ok := officialModelPlazaPricing(modelName, ""); ok {
			return resolvedPricingFromModelPlazaOverride(preset, PricingSourceModelPlazaOfficial), true, nil
		}
		return nil, false, nil
	}
	return nil, false, nil
}

func modelPlazaBillingStatus(model ModelPlazaModel, enabled bool) string {
	hasPrice := model.PricingOverride.IsConfigured()
	if !hasPrice {
		_, hasPrice = officialModelPlazaPricing(model.ModelName, "")
	}
	if !hasPrice {
		return ModelPlazaBillingStatusUnpriced
	}
	if enabled && model.Status == ModelPlazaStatusActive {
		return ModelPlazaBillingStatusApplied
	}
	return ModelPlazaBillingStatusDisplayOnly
}

func resolvedPricingFromModelPlazaOverride(override ModelPlazaPricingOverride, source string) *ResolvedPricing {
	chPricing := modelPlazaOverrideToBillingChannelPricing(override)
	mode := chPricing.BillingMode
	if mode == "" {
		mode = BillingModeToken
	}
	resolved := &ResolvedPricing{
		Mode:           mode,
		Source:         source,
		channelPricing: chPricing,
	}
	switch mode {
	case BillingModePerRequest, BillingModeImage:
		resolved.RequestTiers = filterValidIntervals(chPricing.Intervals)
		if chPricing.PerRequestPrice != nil {
			resolved.DefaultPerRequestPrice = *chPricing.PerRequestPrice
		}
	default:
		resolved.Mode = BillingModeToken
		resolved.BasePricing = &ModelPricing{SupportsCacheBreakdown: true}
		(&ModelPricingResolver{}).applyTokenOverrides(chPricing, resolved)
	}
	return resolved
}

func modelPlazaOverrideToBillingChannelPricing(override ModelPlazaPricingOverride) *ChannelModelPricing {
	pricing := pricingFromOverride(override)
	pricing.InputPrice = modelPlazaPerMillionToPerToken(pricing.InputPrice)
	pricing.OutputPrice = modelPlazaPerMillionToPerToken(pricing.OutputPrice)
	pricing.CacheWritePrice = modelPlazaPerMillionToPerToken(pricing.CacheWritePrice)
	pricing.CacheReadPrice = modelPlazaPerMillionToPerToken(pricing.CacheReadPrice)
	pricing.ImageInputPrice = modelPlazaPerMillionToPerToken(pricing.ImageInputPrice)
	pricing.ImageOutputPrice = modelPlazaPerMillionToPerToken(pricing.ImageOutputPrice)
	for i := range pricing.Intervals {
		pricing.Intervals[i].InputPrice = modelPlazaPerMillionToPerToken(pricing.Intervals[i].InputPrice)
		pricing.Intervals[i].OutputPrice = modelPlazaPerMillionToPerToken(pricing.Intervals[i].OutputPrice)
		pricing.Intervals[i].CacheWritePrice = modelPlazaPerMillionToPerToken(pricing.Intervals[i].CacheWritePrice)
		pricing.Intervals[i].CacheReadPrice = modelPlazaPerMillionToPerToken(pricing.Intervals[i].CacheReadPrice)
	}
	return pricing
}

func modelPlazaPerMillionToPerToken(value *float64) *float64 {
	if value == nil {
		return nil
	}
	converted := *value / 1_000_000
	return &converted
}
