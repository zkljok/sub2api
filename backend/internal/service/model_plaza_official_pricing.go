package service

import "strings"

var modelPlazaFallbackBillingService = NewBillingService(nil, nil)

type modelPlazaOfficialPricingPreset struct {
	Names       []string
	Prefixes    []string
	BillingMode BillingMode
	Input       *float64
	Output      *float64
	CacheWrite  *float64
	CacheRead   *float64
}

func officialModelPlazaPricing(modelName, platform string) (ModelPlazaPricingOverride, bool) {
	name := normalizeOfficialPricingModelName(modelName)
	if name == "" {
		return ModelPlazaPricingOverride{}, false
	}
	if pricing := modelPlazaFallbackBillingService.getFallbackPricing(name); pricing != nil {
		return modelPricingToModelPlazaOverride(pricing), true
	}
	for _, preset := range modelPlazaOfficialPricingPresets() {
		for _, exact := range preset.Names {
			if name == exact {
				return preset.toOverride(), true
			}
		}
		for _, prefix := range preset.Prefixes {
			if strings.HasPrefix(name, prefix) {
				return preset.toOverride(), true
			}
		}
	}
	return ModelPlazaPricingOverride{}, false
}

func modelPricingToModelPlazaOverride(pricing *ModelPricing) ModelPlazaPricingOverride {
	return ModelPlazaPricingOverride{
		BillingMode:      BillingModeToken,
		InputPrice:       pricePerMillionPtr(pricing.InputPricePerToken),
		OutputPrice:      pricePerMillionPtr(pricing.OutputPricePerToken),
		CacheWritePrice:  pricePerMillionPtr(pricing.CacheCreationPricePerToken),
		CacheReadPrice:   pricePerMillionPtr(pricing.CacheReadPricePerToken),
		ImageInputPrice:  pricePerMillionPtr(pricing.ImageInputPricePerToken),
		ImageOutputPrice: pricePerMillionPtr(pricing.ImageOutputPricePerToken),
		Intervals:        []ModelPlazaPricingIntervalView{},
	}
}

func pricePerMillionPtr(value float64) *float64 {
	if value == 0 {
		return nil
	}
	converted := value * 1000000
	return &converted
}

func (p modelPlazaOfficialPricingPreset) toOverride() ModelPlazaPricingOverride {
	mode := p.BillingMode
	if mode == "" {
		mode = BillingModeToken
	}
	return ModelPlazaPricingOverride{
		BillingMode:     mode,
		InputPrice:      cloneFloat64(p.Input),
		OutputPrice:     cloneFloat64(p.Output),
		CacheWritePrice: cloneFloat64(p.CacheWrite),
		CacheReadPrice:  cloneFloat64(p.CacheRead),
		Intervals:       []ModelPlazaPricingIntervalView{},
	}
}

func normalizeOfficialPricingModelName(modelName string) string {
	name := strings.ToLower(strings.TrimSpace(modelName))
	if slash := strings.LastIndex(name, "/"); slash >= 0 && slash < len(name)-1 {
		name = name[slash+1:]
	}
	return strings.TrimSpace(name)
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func modelPlazaFloat64Ptr(value float64) *float64 { return &value }

func modelPlazaOfficialPricingPresets() []modelPlazaOfficialPricingPreset {
	return []modelPlazaOfficialPricingPreset{
		tokenPreset([]string{"gpt-4o", "chatgpt-4o-latest"}, nil, 2.5, 10, 1.25, 1.25),
		tokenPreset([]string{"gpt-4o-mini"}, nil, 0.15, 0.6, 0.075, 0.075),
		tokenPreset([]string{"gpt-4.1"}, nil, 2, 8, 0.5, 0.5),
		tokenPreset([]string{"gpt-4.1-mini"}, nil, 0.4, 1.6, 0.1, 0.1),
		tokenPreset([]string{"gpt-4.1-nano"}, nil, 0.1, 0.4, 0.025, 0.025),
		tokenPreset([]string{"o3"}, []string{"o3-"}, 2, 8, 0.5, 0.5),
		tokenPreset([]string{"o4-mini"}, []string{"o4-mini-"}, 1.1, 4.4, 0.275, 0.275),
		tokenPreset([]string{"gpt-3.5-turbo"}, nil, 0.5, 1.5, 0, 0),
		tokenPreset([]string{"text-embedding-3-small"}, nil, 0.02, 0, 0, 0),
		tokenPreset([]string{"text-embedding-3-large"}, nil, 0.13, 0, 0, 0),

		tokenPreset([]string{"claude-opus-4.1", "claude-opus-4-1"}, []string{"claude-opus-4"}, 15, 75, 18.75, 1.5),
		tokenPreset([]string{"claude-sonnet-4.5", "claude-sonnet-4-5"}, []string{"claude-sonnet-4"}, 3, 15, 3.75, 0.3),
		tokenPreset([]string{"claude-haiku-3.5", "claude-3-5-haiku"}, nil, 0.8, 4, 1, 0.08),
		tokenPreset([]string{"claude-3-opus"}, nil, 15, 75, 18.75, 1.5),
		tokenPreset([]string{"claude-3-sonnet"}, nil, 3, 15, 3.75, 0.3),
		tokenPreset([]string{"claude-3-haiku"}, nil, 0.25, 1.25, 0.3, 0.03),

		tokenPreset([]string{"gemini-2.5-pro"}, nil, 1.25, 10, 0.31, 0.31),
		tokenPreset([]string{"gemini-2.5-flash"}, nil, 0.3, 2.5, 0.075, 0.075),
		tokenPreset([]string{"gemini-2.5-flash-lite"}, nil, 0.1, 0.4, 0.025, 0.025),
		tokenPreset([]string{"gemini-1.5-pro"}, nil, 1.25, 5, 0.3125, 0.3125),
		tokenPreset([]string{"gemini-1.5-flash"}, nil, 0.075, 0.3, 0.01875, 0.01875),

		tokenPreset([]string{"grok-4"}, []string{"grok-4-"}, 3, 15, 0.75, 0.75),
		tokenPreset([]string{"grok-3"}, []string{"grok-3-"}, 3, 15, 0.75, 0.75),
		tokenPreset([]string{"grok-3-mini"}, []string{"grok-3-mini-"}, 0.3, 0.5, 0.075, 0.075),
	}
}

func tokenPreset(names []string, prefixes []string, input, output, cacheWrite, cacheRead float64) modelPlazaOfficialPricingPreset {
	return modelPlazaOfficialPricingPreset{
		Names:       names,
		Prefixes:    prefixes,
		BillingMode: BillingModeToken,
		Input:       modelPlazaFloat64Ptr(input),
		Output:      modelPlazaFloat64Ptr(output),
		CacheWrite:  modelPlazaFloat64Ptr(cacheWrite),
		CacheRead:   modelPlazaFloat64Ptr(cacheRead),
	}
}
