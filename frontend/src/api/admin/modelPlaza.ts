import { apiClient } from '../client'
import type { ModelPlazaModel, ModelPlazaStatus, ModelPlazaNameRule, ModelPlazaPricing, ModelPlazaVendor } from '../modelPlaza'

export type { ModelPlazaStatus, ModelPlazaNameRule, ModelPlazaPricing, ModelPlazaVendor, ModelPlazaModel }

export interface ModelPlazaVendorRequest {
  name: string
  description?: string
  icon?: string
  status?: ModelPlazaStatus
  sort_order?: number
}

export interface ModelPlazaModelRequest {
  model_name: string
  display_name?: string
  description?: string
  icon?: string
  tags?: string[]
  vendor_id?: number | null
  endpoints?: string[]
  status?: ModelPlazaStatus
  name_rule?: ModelPlazaNameRule
  sort_order?: number
  pricing_override?: ModelPlazaPricing
  auto_synced?: boolean
}

export type ModelPlazaBatchAction =
  | 'enable'
  | 'disable'
  | 'delete'
  | 'set_vendor'
  | 'clear_vendor'
  | 'set_tags'
  | 'add_tags'
  | 'remove_tags'
  | 'set_endpoints'
  | 'clear_pricing'

export interface ModelPlazaBatchRequest {
  ids: number[]
  action: ModelPlazaBatchAction
  vendor_id?: number | null
  tags?: string[]
  endpoints?: string[]
}

export async function listVendors(): Promise<ModelPlazaVendor[]> {
  const { data } = await apiClient.get<ModelPlazaVendor[]>('/admin/model-plaza/vendors')
  return data
}

export async function createVendor(req: ModelPlazaVendorRequest): Promise<ModelPlazaVendor> {
  const { data } = await apiClient.post<ModelPlazaVendor>('/admin/model-plaza/vendors', req)
  return data
}

export async function updateVendor(id: number, req: ModelPlazaVendorRequest): Promise<ModelPlazaVendor> {
  const { data } = await apiClient.put<ModelPlazaVendor>(`/admin/model-plaza/vendors/${id}`, req)
  return data
}

export async function removeVendor(id: number): Promise<void> {
  await apiClient.delete(`/admin/model-plaza/vendors/${id}`)
}

export async function listModels(): Promise<ModelPlazaModel[]> {
  const { data } = await apiClient.get<ModelPlazaModel[]>('/admin/model-plaza/models')
  return data
}

export async function createModel(req: ModelPlazaModelRequest): Promise<ModelPlazaModel> {
  const { data } = await apiClient.post<ModelPlazaModel>('/admin/model-plaza/models', req)
  return data
}

export async function updateModel(id: number, req: ModelPlazaModelRequest): Promise<ModelPlazaModel> {
  const { data } = await apiClient.put<ModelPlazaModel>(`/admin/model-plaza/models/${id}`, req)
  return data
}

export async function removeModel(id: number): Promise<void> {
  await apiClient.delete(`/admin/model-plaza/models/${id}`)
}

export async function batchModels(req: ModelPlazaBatchRequest): Promise<{ updated: number }> {
  const { data } = await apiClient.post<{ updated: number }>('/admin/model-plaza/batch/models', req)
  return data
}

export async function syncFromChannels(): Promise<{ inserted: number }> {
  const { data } = await apiClient.post<{ inserted: number }>('/admin/model-plaza/sync')
  return data
}

export default {
  listVendors,
  createVendor,
  updateVendor,
  removeVendor,
  listModels,
  createModel,
  updateModel,
  removeModel,
  batchModels,
  syncFromChannels,
}
