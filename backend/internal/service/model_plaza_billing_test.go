package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

type modelPlazaBillingRepoStub struct {
	models []ModelPlazaModel
}

func (r *modelPlazaBillingRepoStub) ListVendors(context.Context, bool) ([]ModelPlazaVendor, error) {
	return nil, nil
}
func (r *modelPlazaBillingRepoStub) CreateVendor(context.Context, *ModelPlazaVendor) error {
	return nil
}
func (r *modelPlazaBillingRepoStub) UpdateVendor(context.Context, *ModelPlazaVendor) error {
	return nil
}
func (r *modelPlazaBillingRepoStub) DeleteVendor(context.Context, int64) error { return nil }
func (r *modelPlazaBillingRepoStub) ListModels(context.Context, bool) ([]ModelPlazaModel, error) {
	return r.models, nil
}
func (r *modelPlazaBillingRepoStub) CreateModel(context.Context, *ModelPlazaModel) error { return nil }
func (r *modelPlazaBillingRepoStub) UpdateModel(context.Context, *ModelPlazaModel) error { return nil }
func (r *modelPlazaBillingRepoStub) DeleteModel(context.Context, int64) error            { return nil }
func (r *modelPlazaBillingRepoStub) InsertMissingModels(context.Context, []string) (int, error) {
	return 0, nil
}

type modelPlazaBillingSettingRepoStub struct {
	values map[string]string
}

func (r *modelPlazaBillingSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	return nil, ErrSettingNotFound
}
func (r *modelPlazaBillingSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if r.values == nil {
		return "", ErrSettingNotFound
	}
	value, ok := r.values[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return value, nil
}
func (r *modelPlazaBillingSettingRepoStub) Set(_ context.Context, key, value string) error {
	if r.values == nil {
		r.values = map[string]string{}
	}
	r.values[key] = value
	return nil
}
func (r *modelPlazaBillingSettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	return map[string]string{}, nil
}
func (r *modelPlazaBillingSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	return nil
}
func (r *modelPlazaBillingSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}
func (r *modelPlazaBillingSettingRepoStub) Delete(context.Context, string) error { return nil }

func TestModelPlazaBillingSettingsDefaultDisabled(t *testing.T) {
	svc := NewModelPlazaService(&modelPlazaBillingRepoStub{}, nil, nil, &modelPlazaBillingSettingRepoStub{})
	settings, err := svc.GetBillingSettings(context.Background())
	if err != nil {
		t.Fatalf("GetBillingSettings: %v", err)
	}
	if settings.Enabled {
		t.Fatalf("default model plaza billing should be disabled")
	}
}

func TestModelPlazaBillingResolveConvertsPerMillionPricing(t *testing.T) {
	input, output := 2.5, 10.0
	svc := NewModelPlazaService(
		&modelPlazaBillingRepoStub{models: []ModelPlazaModel{
			{
				ModelName: "gpt-4o",
				Status:    ModelPlazaStatusActive,
				NameRule:  ModelPlazaNameRuleExact,
				PricingOverride: ModelPlazaPricingOverride{
					BillingMode: BillingModeToken,
					InputPrice:  &input,
					OutputPrice: &output,
				},
			},
		}},
		nil,
		nil,
		&modelPlazaBillingSettingRepoStub{values: map[string]string{SettingKeyModelPlazaBillingSettings: `{"enabled":true}`}},
	)

	resolved, ok, err := svc.ResolveBillingPricing(context.Background(), "gpt-4o")
	if err != nil {
		t.Fatalf("ResolveBillingPricing: %v", err)
	}
	if !ok || resolved == nil {
		t.Fatalf("expected resolved model plaza pricing")
	}
	if resolved.Source != PricingSourceModelPlazaOverride {
		t.Fatalf("source = %q", resolved.Source)
	}
	if got, want := resolved.BasePricing.InputPricePerToken, 0.0000025; got != want {
		t.Fatalf("input price = %.10f, want %.10f", got, want)
	}
	if got, want := resolved.BasePricing.OutputPricePerToken, 0.00001; got != want {
		t.Fatalf("output price = %.10f, want %.10f", got, want)
	}
}

func TestModelPricingResolverSkipsModelPlazaWhenDisabled(t *testing.T) {
	input := 2.5
	plaza := NewModelPlazaService(
		&modelPlazaBillingRepoStub{models: []ModelPlazaModel{
			{
				ModelName: "gpt-4o",
				Status:    ModelPlazaStatusActive,
				NameRule:  ModelPlazaNameRuleExact,
				PricingOverride: ModelPlazaPricingOverride{
					BillingMode: BillingModeToken,
					InputPrice:  &input,
				},
			},
		}},
		nil,
		nil,
		&modelPlazaBillingSettingRepoStub{values: map[string]string{SettingKeyModelPlazaBillingSettings: `{"enabled":false}`}},
	)
	resolver := NewModelPricingResolver(nil, NewBillingService(&config.Config{}, nil))
	resolver.modelPlazaService = plaza

	resolved := resolver.Resolve(context.Background(), PricingInput{Model: "gpt-4o"})
	if resolved == nil {
		t.Fatalf("expected fallback resolver result")
	}
	if resolved.Source == PricingSourceModelPlazaOverride {
		t.Fatalf("disabled model plaza pricing should not be used")
	}
}

func TestGatewayTokenCostUsesModelPlazaPricingWithGroupMultiplier(t *testing.T) {
	input, output := 2.0, 4.0
	plaza := NewModelPlazaService(
		&modelPlazaBillingRepoStub{models: []ModelPlazaModel{
			{
				ModelName: "custom-model",
				Status:    ModelPlazaStatusActive,
				NameRule:  ModelPlazaNameRuleExact,
				PricingOverride: ModelPlazaPricingOverride{
					BillingMode: BillingModeToken,
					InputPrice:  &input,
					OutputPrice: &output,
				},
			},
		}},
		nil,
		nil,
		&modelPlazaBillingSettingRepoStub{values: map[string]string{SettingKeyModelPlazaBillingSettings: `{"enabled":true}`}},
	)
	billing := NewBillingService(&config.Config{}, nil)
	resolver := NewModelPricingResolver(nil, billing)
	resolver.modelPlazaService = plaza
	svc := &GatewayService{billingService: billing, resolver: resolver}

	resolved := resolver.Resolve(context.Background(), PricingInput{Model: "custom-model"})
	if got, want := resolved.Source, PricingSourceModelPlazaOverride; got != want {
		t.Fatalf("pricing source = %q, want %q", got, want)
	}

	cost := svc.calculateTokenCost(
		context.Background(),
		&ForwardResult{Usage: ClaudeUsage{InputTokens: 1_000_000, OutputTokens: 500_000}},
		&APIKey{Group: &Group{ID: 7}},
		"custom-model",
		1.5,
		&recordUsageOpts{},
	)

	if got, want := cost.TotalCost, 4.0; got != want {
		t.Fatalf("total cost = %.6f, want %.6f", got, want)
	}
	if got, want := cost.ActualCost, 6.0; got != want {
		t.Fatalf("actual cost = %.6f, want %.6f", got, want)
	}
}

func TestModelPlazaAdminBillingStatus(t *testing.T) {
	input := 2.5
	svc := NewModelPlazaService(
		&modelPlazaBillingRepoStub{models: []ModelPlazaModel{
			{ID: 1, ModelName: "priced", Status: ModelPlazaStatusActive, NameRule: ModelPlazaNameRuleExact, PricingOverride: ModelPlazaPricingOverride{InputPrice: &input}},
			{ID: 2, ModelName: "empty", Status: ModelPlazaStatusActive, NameRule: ModelPlazaNameRuleExact, CreatedAt: time.Now()},
		}},
		nil,
		nil,
		&modelPlazaBillingSettingRepoStub{values: map[string]string{SettingKeyModelPlazaBillingSettings: `{"enabled":true}`}},
	)
	rows, err := svc.ListModelsForAdmin(context.Background(), true)
	if err != nil {
		t.Fatalf("ListModelsForAdmin: %v", err)
	}
	if rows[0].BillingStatus != ModelPlazaBillingStatusApplied {
		t.Fatalf("priced status = %q", rows[0].BillingStatus)
	}
	if rows[1].BillingStatus != ModelPlazaBillingStatusUnpriced {
		t.Fatalf("empty status = %q", rows[1].BillingStatus)
	}
}
