import { apiClient } from './client'

export type ModelPlazaStatus = 'active' | 'disabled'
export type ModelPlazaNameRule = 'exact' | 'prefix' | 'suffix' | 'contains'
export type ModelPlazaBillingMode = 'token' | 'per_request' | 'image' | 'video' | ''

export interface ModelPlazaPricingInterval {
  min_tokens: number
  max_tokens: number | null
  tier_label?: string
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_read_price: number | null
  per_request_price: number | null
}

export interface ModelPlazaPricing {
  billing_mode: ModelPlazaBillingMode
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_read_price: number | null
  image_input_price: number | null
  image_output_price: number | null
  per_request_price: number | null
  intervals: ModelPlazaPricingInterval[]
}

export interface ModelPlazaGroup {
  id: number
  name: string
  platform: string
  rate_multiplier: number
}

export interface ModelPlazaVendor {
  id: number
  name: string
  description: string
  icon: string
  status: ModelPlazaStatus
  sort_order: number
  created_at: string
  updated_at: string
}

export interface ModelPlazaModel {
  id: number
  model_name: string
  display_name: string
  description: string
  icon: string
  tags: string[]
  vendor_id: number | null
  endpoints: string[]
  status: ModelPlazaStatus
  name_rule: ModelPlazaNameRule
  sort_order: number
  pricing_override: ModelPlazaPricing
  billing_status?: 'applied' | 'display_only' | 'unpriced'
  auto_synced: boolean
  created_at: string
  updated_at: string
}

export interface ModelPlazaModelView {
  id?: number
  model_name: string
  display_name: string
  description?: string
  icon?: string
  tags: string[]
  vendor_id?: number | null
  platform: string
  groups: ModelPlazaGroup[]
  supported_endpoints: string[]
  pricing?: ModelPlazaPricing | null
  pricing_source: string
  sort_order: number
}

export interface ModelPlazaSnapshot {
  models: ModelPlazaModelView[]
  vendors: ModelPlazaVendor[]
  supported_endpoints: string[]
  pricing_version: string
  generated_at: string
}

export async function getSnapshot(options?: { signal?: AbortSignal }): Promise<ModelPlazaSnapshot> {
  const { data } = await apiClient.get<ModelPlazaSnapshot>('/model-plaza', { signal: options?.signal })
  return data
}

export default { getSnapshot }
