package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	ModelPlazaStatusActive   = "active"
	ModelPlazaStatusDisabled = "disabled"
)

type ModelPlazaNameRule string

const (
	ModelPlazaNameRuleExact    ModelPlazaNameRule = "exact"
	ModelPlazaNameRulePrefix   ModelPlazaNameRule = "prefix"
	ModelPlazaNameRuleSuffix   ModelPlazaNameRule = "suffix"
	ModelPlazaNameRuleContains ModelPlazaNameRule = "contains"
)

func (r ModelPlazaNameRule) IsValid() bool {
	switch r {
	case ModelPlazaNameRuleExact, ModelPlazaNameRulePrefix, ModelPlazaNameRuleSuffix, ModelPlazaNameRuleContains:
		return true
	default:
		return false
	}
}

type ModelPlazaVendor struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Status      string    `json:"status"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ModelPlazaPricingOverride struct {
	BillingMode      BillingMode                     `json:"billing_mode"`
	InputPrice       *float64                        `json:"input_price"`
	OutputPrice      *float64                        `json:"output_price"`
	CacheWritePrice  *float64                        `json:"cache_write_price"`
	CacheReadPrice   *float64                        `json:"cache_read_price"`
	ImageInputPrice  *float64                        `json:"image_input_price"`
	ImageOutputPrice *float64                        `json:"image_output_price"`
	PerRequestPrice  *float64                        `json:"per_request_price"`
	Intervals        []ModelPlazaPricingIntervalView `json:"intervals"`
}

func (p ModelPlazaPricingOverride) IsConfigured() bool {
	return p.BillingMode != "" || p.InputPrice != nil || p.OutputPrice != nil ||
		p.CacheWritePrice != nil || p.CacheReadPrice != nil || p.ImageInputPrice != nil ||
		p.ImageOutputPrice != nil || p.PerRequestPrice != nil || len(p.Intervals) > 0
}

type ModelPlazaModel struct {
	ID              int64                     `json:"id"`
	ModelName       string                    `json:"model_name"`
	DisplayName     string                    `json:"display_name"`
	Description     string                    `json:"description"`
	Icon            string                    `json:"icon"`
	Tags            []string                  `json:"tags"`
	VendorID        *int64                    `json:"vendor_id"`
	Endpoints       []string                  `json:"endpoints"`
	Status          string                    `json:"status"`
	NameRule        ModelPlazaNameRule        `json:"name_rule"`
	SortOrder       int                       `json:"sort_order"`
	PricingOverride ModelPlazaPricingOverride `json:"pricing_override"`
	AutoSynced      bool                      `json:"auto_synced"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

type ModelPlazaRepository interface {
	ListVendors(ctx context.Context, includeDisabled bool) ([]ModelPlazaVendor, error)
	CreateVendor(ctx context.Context, vendor *ModelPlazaVendor) error
	UpdateVendor(ctx context.Context, vendor *ModelPlazaVendor) error
	DeleteVendor(ctx context.Context, id int64) error
	ListModels(ctx context.Context, includeDisabled bool) ([]ModelPlazaModel, error)
	CreateModel(ctx context.Context, model *ModelPlazaModel) error
	UpdateModel(ctx context.Context, model *ModelPlazaModel) error
	DeleteModel(ctx context.Context, id int64) error
	InsertMissingModels(ctx context.Context, models []string) (int, error)
}

type ModelPlazaGroup struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Platform       string  `json:"platform"`
	RateMultiplier float64 `json:"rate_multiplier"`
}

type ModelPlazaPricingIntervalView struct {
	MinTokens       int      `json:"min_tokens"`
	MaxTokens       *int     `json:"max_tokens"`
	TierLabel       string   `json:"tier_label,omitempty"`
	InputPrice      *float64 `json:"input_price"`
	OutputPrice     *float64 `json:"output_price"`
	CacheWritePrice *float64 `json:"cache_write_price"`
	CacheReadPrice  *float64 `json:"cache_read_price"`
	PerRequestPrice *float64 `json:"per_request_price"`
}

type ModelPlazaPricingView struct {
	BillingMode      BillingMode                     `json:"billing_mode"`
	InputPrice       *float64                        `json:"input_price"`
	OutputPrice      *float64                        `json:"output_price"`
	CacheWritePrice  *float64                        `json:"cache_write_price"`
	CacheReadPrice   *float64                        `json:"cache_read_price"`
	ImageInputPrice  *float64                        `json:"image_input_price"`
	ImageOutputPrice *float64                        `json:"image_output_price"`
	PerRequestPrice  *float64                        `json:"per_request_price"`
	Intervals        []ModelPlazaPricingIntervalView `json:"intervals"`
}

type ModelPlazaModelView struct {
	ID                 int64                  `json:"id,omitempty"`
	ModelName          string                 `json:"model_name"`
	DisplayName        string                 `json:"display_name"`
	Description        string                 `json:"description,omitempty"`
	Icon               string                 `json:"icon,omitempty"`
	Tags               []string               `json:"tags"`
	VendorID           *int64                 `json:"vendor_id,omitempty"`
	Platform           string                 `json:"platform"`
	Groups             []ModelPlazaGroup      `json:"groups"`
	SupportedEndpoints []string               `json:"supported_endpoints"`
	Pricing            *ModelPlazaPricingView `json:"pricing,omitempty"`
	PricingSource      string                 `json:"pricing_source"`
	SortOrder          int                    `json:"sort_order"`
}

type ModelPlazaSnapshot struct {
	Models             []ModelPlazaModelView `json:"models"`
	Vendors            []ModelPlazaVendor    `json:"vendors"`
	SupportedEndpoints []string              `json:"supported_endpoints"`
	PricingVersion     string                `json:"pricing_version"`
	GeneratedAt        time.Time             `json:"generated_at"`
}

type ModelPlazaService struct {
	repo           ModelPlazaRepository
	channelService *ChannelService
}

func NewModelPlazaService(repo ModelPlazaRepository, channelService *ChannelService) *ModelPlazaService {
	return &ModelPlazaService{repo: repo, channelService: channelService}
}

func (s *ModelPlazaService) ListVendors(ctx context.Context, includeDisabled bool) ([]ModelPlazaVendor, error) {
	return s.repo.ListVendors(ctx, includeDisabled)
}

func (s *ModelPlazaService) CreateVendor(ctx context.Context, vendor *ModelPlazaVendor) error {
	vendor.Name = strings.TrimSpace(vendor.Name)
	if vendor.Name == "" {
		return fmt.Errorf("vendor name is required")
	}
	if vendor.Status == "" {
		vendor.Status = ModelPlazaStatusActive
	}
	return s.repo.CreateVendor(ctx, vendor)
}

func (s *ModelPlazaService) UpdateVendor(ctx context.Context, vendor *ModelPlazaVendor) error {
	vendor.Name = strings.TrimSpace(vendor.Name)
	if vendor.Name == "" {
		return fmt.Errorf("vendor name is required")
	}
	if vendor.Status == "" {
		vendor.Status = ModelPlazaStatusActive
	}
	return s.repo.UpdateVendor(ctx, vendor)
}

func (s *ModelPlazaService) DeleteVendor(ctx context.Context, id int64) error {
	return s.repo.DeleteVendor(ctx, id)
}

func (s *ModelPlazaService) ListModels(ctx context.Context, includeDisabled bool) ([]ModelPlazaModel, error) {
	return s.repo.ListModels(ctx, includeDisabled)
}

func (s *ModelPlazaService) CreateModel(ctx context.Context, model *ModelPlazaModel) error {
	if err := normalizeModelPlazaModel(model); err != nil {
		return err
	}
	return s.repo.CreateModel(ctx, model)
}

func (s *ModelPlazaService) UpdateModel(ctx context.Context, model *ModelPlazaModel) error {
	if err := normalizeModelPlazaModel(model); err != nil {
		return err
	}
	return s.repo.UpdateModel(ctx, model)
}

func (s *ModelPlazaService) DeleteModel(ctx context.Context, id int64) error {
	return s.repo.DeleteModel(ctx, id)
}

func (s *ModelPlazaService) SyncFromChannels(ctx context.Context) (int, error) {
	channels, err := s.channelService.ListAvailable(ctx)
	if err != nil {
		return 0, err
	}
	seen := make(map[string]struct{})
	models := make([]string, 0)
	for _, channel := range channels {
		if channel.Status != StatusActive {
			continue
		}
		for _, model := range channel.SupportedModels {
			name := strings.TrimSpace(model.Name)
			if name == "" {
				continue
			}
			key := strings.ToLower(name)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			models = append(models, name)
		}
	}
	sort.Strings(models)
	return s.repo.InsertMissingModels(ctx, models)
}

// Snapshot returns public pricing data. Only non-exclusive groups are included:
// anonymous visitors never see user-specific availability or multipliers.
func (s *ModelPlazaService) Snapshot(ctx context.Context) (*ModelPlazaSnapshot, error) {
	metadata, err := s.repo.ListModels(ctx, false)
	if err != nil {
		return nil, err
	}
	vendors, err := s.repo.ListVendors(ctx, false)
	if err != nil {
		return nil, err
	}
	channels, err := s.channelService.ListAvailable(ctx)
	if err != nil {
		return nil, err
	}

	entries := make(map[string]*ModelPlazaModelView)
	for _, channel := range channels {
		if channel.Status != StatusActive {
			continue
		}
		publicGroups := make([]ModelPlazaGroup, 0, len(channel.Groups))
		for _, group := range channel.Groups {
			if group.IsExclusive {
				continue
			}
			publicGroups = append(publicGroups, ModelPlazaGroup{
				ID: group.ID, Name: group.Name, Platform: group.Platform, RateMultiplier: group.RateMultiplier,
			})
		}
		if len(publicGroups) == 0 {
			continue
		}
		for _, supported := range channel.SupportedModels {
			key := strings.ToLower(strings.TrimSpace(supported.Name))
			if key == "" {
				continue
			}
			entry, ok := entries[key]
			if !ok {
				entry = &ModelPlazaModelView{
					ModelName: supported.Name, DisplayName: supported.Name, Platform: supported.Platform,
					Tags: []string{}, Groups: []ModelPlazaGroup{}, SupportedEndpoints: defaultEndpointsForPlatform(supported.Platform),
					Pricing: pricingViewFromChannel(supported.Pricing), PricingSource: "channel",
				}
				entries[key] = entry
			}
			entry.Groups = mergeModelPlazaGroups(entry.Groups, publicGroups)
			entry.SupportedEndpoints = mergeStrings(entry.SupportedEndpoints, defaultEndpointsForPlatform(supported.Platform))
			if entry.Pricing == nil && supported.Pricing != nil {
				entry.Pricing = pricingViewFromChannel(supported.Pricing)
			}
		}
	}

	for _, meta := range metadata {
		if meta.Status != ModelPlazaStatusActive {
			continue
		}
		for _, entry := range entries {
			if !matchesModelPlazaRule(meta, entry.ModelName) {
				continue
			}
			entry.ID = meta.ID
			entry.DisplayName = firstNonEmptyModelPlazaValue(meta.DisplayName, entry.ModelName)
			entry.Description = meta.Description
			entry.Icon = meta.Icon
			entry.Tags = append([]string{}, meta.Tags...)
			entry.VendorID = meta.VendorID
			entry.SortOrder = meta.SortOrder
			entry.SupportedEndpoints = mergeStrings(entry.SupportedEndpoints, meta.Endpoints)
			if meta.PricingOverride.IsConfigured() {
				entry.Pricing = pricingViewFromChannel(pricingFromOverride(meta.PricingOverride))
				entry.PricingSource = "override"
			}
		}
	}

	rows := make([]ModelPlazaModelView, 0, len(entries))
	endpoints := make([]string, 0)
	for _, entry := range entries {
		sort.SliceStable(entry.Groups, func(i, j int) bool { return entry.Groups[i].Name < entry.Groups[j].Name })
		sort.Strings(entry.SupportedEndpoints)
		endpoints = mergeStrings(endpoints, entry.SupportedEndpoints)
		rows = append(rows, *entry)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].SortOrder != rows[j].SortOrder {
			return rows[i].SortOrder < rows[j].SortOrder
		}
		return strings.ToLower(rows[i].DisplayName) < strings.ToLower(rows[j].DisplayName)
	})
	sort.Strings(endpoints)
	now := time.Now().UTC()
	return &ModelPlazaSnapshot{
		Models: rows, Vendors: vendors, SupportedEndpoints: endpoints,
		PricingVersion: modelPlazaVersion(rows, vendors), GeneratedAt: now,
	}, nil
}

func normalizeModelPlazaModel(model *ModelPlazaModel) error {
	model.ModelName = strings.TrimSpace(model.ModelName)
	model.DisplayName = strings.TrimSpace(model.DisplayName)
	if model.ModelName == "" {
		return fmt.Errorf("model name is required")
	}
	if model.NameRule == "" {
		model.NameRule = ModelPlazaNameRuleExact
	}
	if !model.NameRule.IsValid() {
		return fmt.Errorf("invalid model name rule")
	}
	if model.Status == "" {
		model.Status = ModelPlazaStatusActive
	}
	if model.Status != ModelPlazaStatusActive && model.Status != ModelPlazaStatusDisabled {
		return fmt.Errorf("invalid model status")
	}
	model.Tags = normalizeStringList(model.Tags)
	model.Endpoints = normalizeStringList(model.Endpoints)
	return nil
}

func matchesModelPlazaRule(meta ModelPlazaModel, name string) bool {
	left, right := strings.ToLower(strings.TrimSpace(name)), strings.ToLower(strings.TrimSpace(meta.ModelName))
	switch meta.NameRule {
	case ModelPlazaNameRulePrefix:
		return strings.HasPrefix(left, right)
	case ModelPlazaNameRuleSuffix:
		return strings.HasSuffix(left, right)
	case ModelPlazaNameRuleContains:
		return strings.Contains(left, right)
	default:
		return left == right
	}
}

func pricingViewFromChannel(pricing *ChannelModelPricing) *ModelPlazaPricingView {
	if pricing == nil {
		return nil
	}
	mode := pricing.BillingMode
	if mode == "" {
		mode = BillingModeToken
	}
	view := &ModelPlazaPricingView{
		BillingMode:      mode,
		InputPrice:       pricing.InputPrice,
		OutputPrice:      pricing.OutputPrice,
		CacheWritePrice:  pricing.CacheWritePrice,
		CacheReadPrice:   pricing.CacheReadPrice,
		ImageInputPrice:  pricing.ImageInputPrice,
		ImageOutputPrice: pricing.ImageOutputPrice,
		PerRequestPrice:  pricing.PerRequestPrice,
		Intervals:        make([]ModelPlazaPricingIntervalView, 0, len(pricing.Intervals)),
	}
	for _, iv := range pricing.Intervals {
		view.Intervals = append(view.Intervals, ModelPlazaPricingIntervalView{
			MinTokens:       iv.MinTokens,
			MaxTokens:       iv.MaxTokens,
			TierLabel:       iv.TierLabel,
			InputPrice:      iv.InputPrice,
			OutputPrice:     iv.OutputPrice,
			CacheWritePrice: iv.CacheWritePrice,
			CacheReadPrice:  iv.CacheReadPrice,
			PerRequestPrice: iv.PerRequestPrice,
		})
	}
	return view
}

func pricingFromOverride(override ModelPlazaPricingOverride) *ChannelModelPricing {
	var result ChannelModelPricing
	if override.BillingMode != "" {
		result.BillingMode = override.BillingMode
	}
	if override.InputPrice != nil {
		result.InputPrice = override.InputPrice
	}
	if override.OutputPrice != nil {
		result.OutputPrice = override.OutputPrice
	}
	if override.CacheWritePrice != nil {
		result.CacheWritePrice = override.CacheWritePrice
	}
	if override.CacheReadPrice != nil {
		result.CacheReadPrice = override.CacheReadPrice
	}
	if override.ImageInputPrice != nil {
		result.ImageInputPrice = override.ImageInputPrice
	}
	if override.ImageOutputPrice != nil {
		result.ImageOutputPrice = override.ImageOutputPrice
	}
	if override.PerRequestPrice != nil {
		result.PerRequestPrice = override.PerRequestPrice
	}
	if len(override.Intervals) > 0 {
		result.Intervals = make([]PricingInterval, 0, len(override.Intervals))
		for _, iv := range override.Intervals {
			result.Intervals = append(result.Intervals, PricingInterval{
				MinTokens:       iv.MinTokens,
				MaxTokens:       iv.MaxTokens,
				TierLabel:       iv.TierLabel,
				InputPrice:      iv.InputPrice,
				OutputPrice:     iv.OutputPrice,
				CacheWritePrice: iv.CacheWritePrice,
				CacheReadPrice:  iv.CacheReadPrice,
				PerRequestPrice: iv.PerRequestPrice,
			})
		}
	}
	return &result
}

func mergeModelPlazaGroups(base, more []ModelPlazaGroup) []ModelPlazaGroup {
	seen := make(map[int64]struct{}, len(base))
	for _, group := range base {
		seen[group.ID] = struct{}{}
	}
	for _, group := range more {
		if _, ok := seen[group.ID]; ok {
			continue
		}
		seen[group.ID] = struct{}{}
		base = append(base, group)
	}
	return base
}

func mergeStrings(base, more []string) []string {
	seen := make(map[string]struct{}, len(base)+len(more))
	result := make([]string, 0, len(base)+len(more))
	for _, value := range append(base, more...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeStringList(values []string) []string { return mergeStrings(nil, values) }
func firstNonEmptyModelPlazaValue(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func defaultEndpointsForPlatform(platform string) []string {
	switch platform {
	case PlatformOpenAI:
		return []string{"chat-completions", "responses"}
	case PlatformAnthropic:
		return []string{"messages"}
	case PlatformGemini:
		return []string{"generate-content"}
	case PlatformAntigravity:
		return []string{"generate-content"}
	case PlatformGrok:
		return []string{"chat-completions", "responses"}
	default:
		return []string{}
	}
}

func modelPlazaVersion(models []ModelPlazaModelView, vendors []ModelPlazaVendor) string {
	raw := fmt.Sprintf("%v|%v", models, vendors)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
