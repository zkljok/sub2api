<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 class="text-xl font-semibold text-gray-900 dark:text-white">模型广场管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            维护公开模型广场的厂商信息、展示文案、端点标签和价格覆盖。
          </p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <button type="button" class="btn btn-secondary" :disabled="loading" @click="loadAll">
            <Icon name="refresh" size="sm" class="mr-1.5" :class="{ 'animate-spin': loading }" />
            刷新
          </button>
          <button type="button" class="btn btn-secondary" :disabled="syncing" @click="syncModels">
            <Icon name="download" size="sm" class="mr-1.5" :class="{ 'animate-pulse': syncing }" />
            从账号/渠道同步模型
          </button>
          <RouterLink to="/models" class="btn btn-secondary">
            <Icon name="eye" size="sm" class="mr-1.5" />
            查看公开页
          </RouterLink>
        </div>
      </div>

      <div class="grid gap-6 xl:grid-cols-[380px_minmax(0,1fr)]">
        <section class="card p-5">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">厂商配置</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">用于模型卡片的归属和筛选。</p>
            </div>
            <button type="button" class="btn btn-primary btn-sm" @click="openVendorDialog()">
              <Icon name="plus" size="sm" class="mr-1" />
              新增
            </button>
          </div>

          <div v-if="loading" class="py-8 text-center text-sm text-gray-500">加载中...</div>
          <div v-else-if="vendors.length === 0" class="py-8 text-center text-sm text-gray-500">暂无厂商</div>
          <div v-else class="space-y-3">
            <div
              v-for="vendor in vendors"
              :key="vendor.id"
              class="rounded-lg border border-gray-200 p-3 dark:border-dark-600"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <div class="flex items-center gap-2">
                    <span class="truncate font-medium text-gray-900 dark:text-white">{{ vendor.name }}</span>
                    <span :class="statusClass(vendor.status)">{{ statusLabel(vendor.status) }}</span>
                  </div>
                  <p class="mt-1 line-clamp-2 text-xs text-gray-500 dark:text-gray-400">
                    {{ vendor.description || '未填写描述' }}
                  </p>
                  <div class="mt-2 text-xs text-gray-400">排序：{{ vendor.sort_order }}</div>
                </div>
                <div class="flex shrink-0 items-center gap-1">
                  <button type="button" class="btn btn-secondary btn-sm" @click="openVendorDialog(vendor)">编辑</button>
                  <button type="button" class="btn btn-danger btn-sm" @click="deleteVendor(vendor)">删除</button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="card p-5">
          <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">模型配置</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                优先从账号管理的模型映射同步，缺失时再从渠道补充；已有展示和价格覆盖不会被覆盖。
              </p>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <div class="relative">
                <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                <input v-model.trim="search" type="text" class="input h-9 w-56 pl-9" placeholder="搜索模型名称" />
              </div>
              <button type="button" class="btn btn-primary btn-sm" @click="openModelDialog()">
                <Icon name="plus" size="sm" class="mr-1" />
                新增模型
              </button>
            </div>
          </div>

          <div class="mb-4 rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-800/50">
            <div class="flex flex-wrap items-center gap-2">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-200">
                已选择 {{ selectedIds.length }} / {{ filteredModels.length }}
              </span>
              <button type="button" class="btn btn-secondary btn-sm" :disabled="filteredModels.length === 0" @click="toggleAllFiltered">
                {{ allFilteredSelected ? '取消本页选择' : '选择当前结果' }}
              </button>
              <button type="button" class="btn btn-secondary btn-sm" :disabled="selectedIds.length === 0" @click="clearSelection">
                清空选择
              </button>
              <div class="mx-1 hidden h-6 w-px bg-gray-200 dark:bg-dark-600 sm:block" />
              <select v-model="batchAction" class="input h-9 w-36">
                <option value="enable">批量启用</option>
                <option value="disable">批量禁用</option>
                <option value="delete">批量删除</option>
                <option value="set_vendor">设置厂商</option>
                <option value="clear_vendor">清空厂商</option>
                <option value="set_tags">替换标签</option>
                <option value="add_tags">追加标签</option>
                <option value="remove_tags">移除标签</option>
                <option value="set_endpoints">替换端点</option>
                <option value="clear_pricing">清空价格覆盖</option>
              </select>
              <select v-if="batchAction === 'set_vendor'" v-model="batchVendorId" class="input h-9 w-40">
                <option :value="null">不指定</option>
                <option v-for="vendor in vendors" :key="vendor.id" :value="vendor.id">{{ vendor.name }}</option>
              </select>
              <input
                v-if="['set_tags', 'add_tags', 'remove_tags'].includes(batchAction)"
                v-model.trim="batchTagsText"
                class="input h-9 w-56"
                placeholder="标签，逗号分隔"
              />
              <input
                v-if="batchAction === 'set_endpoints'"
                v-model.trim="batchEndpointsText"
                class="input h-9 w-64"
                placeholder="端点，逗号分隔"
              />
              <button type="button" class="btn btn-primary btn-sm" :disabled="saving || selectedIds.length === 0" @click="runBatchAction">
                执行批量操作
              </button>
            </div>
          </div>

          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-600">
              <thead>
                <tr class="text-left text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                  <th class="w-10 px-3 py-2">
                    <input type="checkbox" class="h-4 w-4 rounded border-gray-300" :checked="allFilteredSelected" @change="toggleAllFiltered" />
                  </th>
                  <th class="px-3 py-2">模型</th>
                  <th class="px-3 py-2">厂商</th>
                  <th class="px-3 py-2">端点</th>
                  <th class="px-3 py-2">价格覆盖</th>
                  <th class="px-3 py-2">状态</th>
                  <th class="px-3 py-2 text-right">操作</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-if="loading">
                  <td colspan="7" class="px-3 py-10 text-center text-sm text-gray-500">加载中...</td>
                </tr>
                <tr v-else-if="filteredModels.length === 0">
                  <td colspan="7" class="px-3 py-10 text-center text-sm text-gray-500">暂无模型</td>
                </tr>
                <tr v-for="model in filteredModels" v-else :key="model.id" class="align-top">
                  <td class="px-3 py-3">
                    <input v-model="selectedIds" type="checkbox" class="h-4 w-4 rounded border-gray-300" :value="model.id" />
                  </td>
                  <td class="px-3 py-3">
                    <div class="font-medium text-gray-900 dark:text-white">{{ model.display_name || model.model_name }}</div>
                    <div class="mt-0.5 font-mono text-xs text-gray-500">{{ ruleLabel(model.name_rule) }}：{{ model.model_name }}</div>
                    <div v-if="model.tags.length" class="mt-2 flex flex-wrap gap-1">
                      <span v-for="tag in model.tags" :key="tag" class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600 dark:bg-dark-600 dark:text-gray-300">
                        {{ tag }}
                      </span>
                    </div>
                  </td>
                  <td class="px-3 py-3 text-sm text-gray-600 dark:text-gray-300">{{ vendorName(model.vendor_id) }}</td>
                  <td class="px-3 py-3">
                    <div class="flex max-w-xs flex-wrap gap-1">
                      <span v-for="endpoint in model.endpoints" :key="endpoint" class="rounded bg-primary-50 px-1.5 py-0.5 text-xs text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                        {{ endpoint }}
                      </span>
                      <span v-if="model.endpoints.length === 0" class="text-sm text-gray-400">跟随渠道</span>
                    </div>
                  </td>
                  <td class="px-3 py-3 text-sm text-gray-600 dark:text-gray-300">
                    <span v-if="hasPricingOverride(model.pricing_override)">已配置</span>
                    <span v-else class="text-gray-400">跟随渠道</span>
                  </td>
                  <td class="px-3 py-3">
                    <span :class="statusClass(model.status)">{{ statusLabel(model.status) }}</span>
                    <div v-if="model.auto_synced" class="mt-1 text-xs text-gray-400">自动同步</div>
                  </td>
                  <td class="px-3 py-3 text-right">
                    <div class="flex justify-end gap-2">
                      <button type="button" class="btn btn-secondary btn-sm" @click="openModelDialog(model)">编辑</button>
                      <button type="button" class="btn btn-danger btn-sm" @click="deleteModel(model)">删除</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </div>

    <BaseDialog :show="vendorDialogOpen" :title="editingVendor ? '编辑厂商' : '新增厂商'" @close="vendorDialogOpen = false">
      <form class="space-y-4" @submit.prevent="saveVendor">
        <div>
          <label class="input-label">厂商名称</label>
          <input v-model.trim="vendorForm.name" class="input" required />
        </div>
        <div>
          <label class="input-label">描述</label>
          <textarea v-model.trim="vendorForm.description" class="input min-h-[84px]" />
        </div>
        <div>
          <label class="input-label">图标 URL</label>
          <input v-model.trim="vendorForm.icon" class="input" placeholder="https://..." />
        </div>
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">状态</label>
            <select v-model="vendorForm.status" class="input">
              <option value="active">启用</option>
              <option value="disabled">禁用</option>
            </select>
          </div>
          <div>
            <label class="input-label">排序</label>
            <input v-model.number="vendorForm.sort_order" type="number" class="input" />
          </div>
        </div>
        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn btn-secondary" @click="vendorDialogOpen = false">取消</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">保存</button>
        </div>
      </form>
    </BaseDialog>

    <BaseDialog :show="modelDialogOpen" :title="editingModel ? '编辑模型' : '新增模型'" width="wide" @close="modelDialogOpen = false">
      <form class="space-y-5" @submit.prevent="saveModel">
        <div class="grid gap-4 lg:grid-cols-2">
          <div>
            <label class="input-label">匹配模型名</label>
            <input v-model.trim="modelForm.model_name" class="input font-mono" required placeholder="gpt-4o" />
          </div>
          <div>
            <label class="input-label">公开显示名称</label>
            <input v-model.trim="modelForm.display_name" class="input" placeholder="留空则使用模型名" />
          </div>
          <div>
            <label class="input-label">匹配规则</label>
            <select v-model="modelForm.name_rule" class="input">
              <option value="exact">完全匹配</option>
              <option value="prefix">前缀匹配</option>
              <option value="suffix">后缀匹配</option>
              <option value="contains">包含匹配</option>
            </select>
          </div>
          <div>
            <label class="input-label">厂商</label>
            <select v-model="modelForm.vendor_id" class="input">
              <option :value="null">不指定</option>
              <option v-for="vendor in vendors" :key="vendor.id" :value="vendor.id">{{ vendor.name }}</option>
            </select>
          </div>
          <div>
            <label class="input-label">标签（逗号分隔）</label>
            <input v-model.trim="tagsText" class="input" placeholder="文本, 视觉, 低延迟" />
          </div>
          <div>
            <label class="input-label">端点（逗号分隔）</label>
            <input v-model.trim="endpointsText" class="input" placeholder="chat-completions, responses" />
          </div>
          <div>
            <label class="input-label">状态</label>
            <select v-model="modelForm.status" class="input">
              <option value="active">启用</option>
              <option value="disabled">禁用</option>
            </select>
          </div>
          <div>
            <label class="input-label">排序</label>
            <input v-model.number="modelForm.sort_order" type="number" class="input" />
          </div>
        </div>

        <div>
          <label class="input-label">公开描述</label>
          <textarea v-model.trim="modelForm.description" class="input min-h-[92px]" />
        </div>

        <div class="rounded-lg border border-gray-200 p-4 dark:border-dark-600">
          <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">价格覆盖</h3>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                留空时公开页自动展示渠道定价；填写后按这里的 USD 价格展示。
              </p>
            </div>
            <button type="button" class="btn btn-secondary btn-sm" @click="clearPricing">清空覆盖</button>
          </div>
          <div class="grid gap-4 lg:grid-cols-4">
            <div>
              <label class="input-label">计费模式</label>
              <select v-model="modelForm.pricing_override.billing_mode" class="input">
                <option value="">不覆盖</option>
                <option value="token">Token</option>
                <option value="per_request">按次</option>
                <option value="image">图片</option>
                <option value="video">视频</option>
              </select>
            </div>
            <div>
              <label class="input-label">输入 $/MTok</label>
              <input v-model="pricingText.input_price" type="number" step="0.000001" class="input" />
            </div>
            <div>
              <label class="input-label">输出 $/MTok</label>
              <input v-model="pricingText.output_price" type="number" step="0.000001" class="input" />
            </div>
            <div>
              <label class="input-label">按次 $/次</label>
              <input v-model="pricingText.per_request_price" type="number" step="0.000001" class="input" />
            </div>
            <div>
              <label class="input-label">缓存写入 $/MTok</label>
              <input v-model="pricingText.cache_write_price" type="number" step="0.000001" class="input" />
            </div>
            <div>
              <label class="input-label">缓存读取 $/MTok</label>
              <input v-model="pricingText.cache_read_price" type="number" step="0.000001" class="input" />
            </div>
            <div>
              <label class="input-label">图片输入 $/图</label>
              <input v-model="pricingText.image_input_price" type="number" step="0.000001" class="input" />
            </div>
            <div>
              <label class="input-label">图片输出 $/图</label>
              <input v-model="pricingText.image_output_price" type="number" step="0.000001" class="input" />
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn btn-secondary" @click="modelDialogOpen = false">取消</button>
          <button type="submit" class="btn btn-primary" :disabled="saving">保存</button>
        </div>
      </form>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { modelPlazaAPI } from '@/api/admin'
import type {
  ModelPlazaBatchAction,
  ModelPlazaModel,
  ModelPlazaModelRequest,
  ModelPlazaPricing,
  ModelPlazaStatus,
  ModelPlazaVendor,
  ModelPlazaVendorRequest,
} from '@/api/admin/modelPlaza'

const appStore = useAppStore()

const vendors = ref<ModelPlazaVendor[]>([])
const models = ref<ModelPlazaModel[]>([])
const loading = ref(false)
const saving = ref(false)
const syncing = ref(false)
const search = ref('')
const selectedIds = ref<number[]>([])
const batchAction = ref<ModelPlazaBatchAction>('enable')
const batchVendorId = ref<number | null>(null)
const batchTagsText = ref('')
const batchEndpointsText = ref('')

const vendorDialogOpen = ref(false)
const modelDialogOpen = ref(false)
const editingVendor = ref<ModelPlazaVendor | null>(null)
const editingModel = ref<ModelPlazaModel | null>(null)
const tagsText = ref('')
const endpointsText = ref('')

const emptyPricing = (): ModelPlazaPricing => ({
  billing_mode: '',
  input_price: null,
  output_price: null,
  cache_write_price: null,
  cache_read_price: null,
  image_input_price: null,
  image_output_price: null,
  per_request_price: null,
  intervals: [],
})

const vendorForm = reactive<ModelPlazaVendorRequest>({
  name: '',
  description: '',
  icon: '',
  status: 'active',
  sort_order: 0,
})

type ModelPlazaModelForm = Omit<ModelPlazaModelRequest, 'pricing_override'> & {
  pricing_override: ModelPlazaPricing
}

const modelForm = reactive<ModelPlazaModelForm>({
  model_name: '',
  display_name: '',
  description: '',
  icon: '',
  tags: [],
  vendor_id: null,
  endpoints: [],
  status: 'active',
  name_rule: 'exact',
  sort_order: 0,
  pricing_override: emptyPricing(),
  auto_synced: false,
})

const pricingText = reactive<Record<keyof Omit<ModelPlazaPricing, 'billing_mode' | 'intervals'>, string>>({
  input_price: '',
  output_price: '',
  cache_write_price: '',
  cache_read_price: '',
  image_input_price: '',
  image_output_price: '',
  per_request_price: '',
})

const filteredModels = computed(() => {
  const q = search.value.toLowerCase()
  if (!q) return models.value
  return models.value.filter((model) =>
    [model.model_name, model.display_name, model.description, ...model.tags].some((value) =>
      String(value || '').toLowerCase().includes(q),
    ),
  )
})

const allFilteredSelected = computed(() => {
  if (filteredModels.value.length === 0) return false
  const selected = new Set(selectedIds.value)
  return filteredModels.value.every((model) => selected.has(model.id))
})

async function loadAll() {
  loading.value = true
  try {
    const [vendorRows, modelRows] = await Promise.all([modelPlazaAPI.listVendors(), modelPlazaAPI.listModels()])
    vendors.value = vendorRows
    models.value = modelRows
    selectedIds.value = selectedIds.value.filter((id) => modelRows.some((model) => model.id === id))
  } catch (error) {
    appStore.showError(errorMessage(error, '加载模型广场配置失败'))
  } finally {
    loading.value = false
  }
}

function toggleAllFiltered() {
  const filteredIds = filteredModels.value.map((model) => model.id)
  if (filteredIds.length === 0) return
  if (allFilteredSelected.value) {
    const filtered = new Set(filteredIds)
    selectedIds.value = selectedIds.value.filter((id) => !filtered.has(id))
    return
  }
  selectedIds.value = Array.from(new Set([...selectedIds.value, ...filteredIds]))
}

function clearSelection() {
  selectedIds.value = []
}

function openVendorDialog(vendor?: ModelPlazaVendor) {
  editingVendor.value = vendor || null
  Object.assign(vendorForm, {
    name: vendor?.name || '',
    description: vendor?.description || '',
    icon: vendor?.icon || '',
    status: vendor?.status || 'active',
    sort_order: vendor?.sort_order || 0,
  })
  vendorDialogOpen.value = true
}

async function saveVendor() {
  saving.value = true
  try {
    if (editingVendor.value) {
      await modelPlazaAPI.updateVendor(editingVendor.value.id, vendorForm)
    } else {
      await modelPlazaAPI.createVendor(vendorForm)
    }
    appStore.showSuccess('厂商已保存')
    vendorDialogOpen.value = false
    await loadAll()
  } catch (error) {
    appStore.showError(errorMessage(error, '保存厂商失败'))
  } finally {
    saving.value = false
  }
}

async function deleteVendor(vendor: ModelPlazaVendor) {
  if (!window.confirm(`确认删除厂商「${vendor.name}」吗？已关联模型会变为不指定厂商。`)) return
  try {
    await modelPlazaAPI.removeVendor(vendor.id)
    appStore.showSuccess('厂商已删除')
    await loadAll()
  } catch (error) {
    appStore.showError(errorMessage(error, '删除厂商失败'))
  }
}

function openModelDialog(model?: ModelPlazaModel) {
  editingModel.value = model || null
  const pricing = model?.pricing_override || emptyPricing()
  Object.assign(modelForm, {
    model_name: model?.model_name || '',
    display_name: model?.display_name || '',
    description: model?.description || '',
    icon: model?.icon || '',
    tags: [...(model?.tags || [])],
    vendor_id: model?.vendor_id ?? null,
    endpoints: [...(model?.endpoints || [])],
    status: model?.status || 'active',
    name_rule: model?.name_rule || 'exact',
    sort_order: model?.sort_order || 0,
    pricing_override: { ...emptyPricing(), ...pricing, intervals: pricing.intervals || [] },
    auto_synced: model?.auto_synced || false,
  })
  tagsText.value = (model?.tags || []).join(', ')
  endpointsText.value = (model?.endpoints || []).join(', ')
  setPricingText(modelForm.pricing_override || emptyPricing())
  modelDialogOpen.value = true
}

async function saveModel() {
  saving.value = true
  try {
    const payload: ModelPlazaModelRequest = {
      ...modelForm,
      tags: splitList(tagsText.value),
      endpoints: splitList(endpointsText.value),
      vendor_id: modelForm.vendor_id === null ? null : Number(modelForm.vendor_id),
      pricing_override: pricingFromForm(),
    }
    if (editingModel.value) {
      await modelPlazaAPI.updateModel(editingModel.value.id, payload)
    } else {
      await modelPlazaAPI.createModel(payload)
    }
    appStore.showSuccess('模型已保存')
    modelDialogOpen.value = false
    await loadAll()
  } catch (error) {
    appStore.showError(errorMessage(error, '保存模型失败'))
  } finally {
    saving.value = false
  }
}

async function deleteModel(model: ModelPlazaModel) {
  if (!window.confirm(`确认删除模型「${model.display_name || model.model_name}」吗？`)) return
  try {
    await modelPlazaAPI.removeModel(model.id)
    appStore.showSuccess('模型已删除')
    await loadAll()
  } catch (error) {
    appStore.showError(errorMessage(error, '删除模型失败'))
  }
}

async function runBatchAction() {
  if (selectedIds.value.length === 0) return
  const actionLabel = batchActionLabel(batchAction.value)
  if (!window.confirm(`确认对 ${selectedIds.value.length} 个模型执行「${actionLabel}」吗？`)) return

  saving.value = true
  try {
    const payload = {
      ids: selectedIds.value,
      action: batchAction.value,
      vendor_id: batchAction.value === 'set_vendor' ? batchVendorId.value : undefined,
      tags: ['set_tags', 'add_tags', 'remove_tags'].includes(batchAction.value) ? splitList(batchTagsText.value) : undefined,
      endpoints: batchAction.value === 'set_endpoints' ? splitList(batchEndpointsText.value) : undefined,
    }
    const result = await modelPlazaAPI.batchModels(payload)
    appStore.showSuccess(`批量操作完成，处理 ${result.updated} 个模型`)
    clearSelection()
    await loadAll()
  } catch (error) {
    appStore.showError(errorMessage(error, '批量操作失败'))
  } finally {
    saving.value = false
  }
}

async function syncModels() {
  syncing.value = true
  try {
    const result = await modelPlazaAPI.syncFromChannels()
    appStore.showSuccess(`同步完成，新增 ${result.inserted} 个模型`)
    await loadAll()
  } catch (error) {
    appStore.showError(errorMessage(error, '同步模型失败'))
  } finally {
    syncing.value = false
  }
}

function splitList(value: string): string[] {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

function numberOrNull(value: string): number | null {
  if (value === '') return null
  const n = Number(value)
  return Number.isFinite(n) ? n : null
}

function setPricingText(pricing: ModelPlazaPricing) {
  pricingText.input_price = pricing.input_price == null ? '' : String(pricing.input_price)
  pricingText.output_price = pricing.output_price == null ? '' : String(pricing.output_price)
  pricingText.cache_write_price = pricing.cache_write_price == null ? '' : String(pricing.cache_write_price)
  pricingText.cache_read_price = pricing.cache_read_price == null ? '' : String(pricing.cache_read_price)
  pricingText.image_input_price = pricing.image_input_price == null ? '' : String(pricing.image_input_price)
  pricingText.image_output_price = pricing.image_output_price == null ? '' : String(pricing.image_output_price)
  pricingText.per_request_price = pricing.per_request_price == null ? '' : String(pricing.per_request_price)
}

function pricingFromForm(): ModelPlazaPricing {
  return {
    billing_mode: modelForm.pricing_override?.billing_mode || '',
    input_price: numberOrNull(pricingText.input_price),
    output_price: numberOrNull(pricingText.output_price),
    cache_write_price: numberOrNull(pricingText.cache_write_price),
    cache_read_price: numberOrNull(pricingText.cache_read_price),
    image_input_price: numberOrNull(pricingText.image_input_price),
    image_output_price: numberOrNull(pricingText.image_output_price),
    per_request_price: numberOrNull(pricingText.per_request_price),
    intervals: modelForm.pricing_override?.intervals || [],
  }
}

function clearPricing() {
  modelForm.pricing_override = emptyPricing()
  setPricingText(modelForm.pricing_override)
}

function hasPricingOverride(pricing: ModelPlazaPricing): boolean {
  return Boolean(
    pricing?.billing_mode ||
      pricing?.input_price != null ||
      pricing?.output_price != null ||
      pricing?.cache_write_price != null ||
      pricing?.cache_read_price != null ||
      pricing?.image_input_price != null ||
      pricing?.image_output_price != null ||
      pricing?.per_request_price != null ||
      (pricing?.intervals?.length || 0) > 0,
  )
}

function vendorName(id?: number | null): string {
  if (!id) return '未指定'
  return vendors.value.find((vendor) => vendor.id === id)?.name || `#${id}`
}

function statusLabel(status: ModelPlazaStatus): string {
  return status === 'active' ? '启用' : '禁用'
}

function statusClass(status: ModelPlazaStatus): string {
  return status === 'active'
    ? 'inline-flex rounded bg-emerald-50 px-2 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
    : 'inline-flex rounded bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600 dark:bg-dark-600 dark:text-gray-300'
}

function ruleLabel(rule: string): string {
  const labels: Record<string, string> = {
    exact: '完全匹配',
    prefix: '前缀匹配',
    suffix: '后缀匹配',
    contains: '包含匹配',
  }
  return labels[rule] || rule
}

function batchActionLabel(action: ModelPlazaBatchAction): string {
  const labels: Record<ModelPlazaBatchAction, string> = {
    enable: '批量启用',
    disable: '批量禁用',
    delete: '批量删除',
    set_vendor: '设置厂商',
    clear_vendor: '清空厂商',
    set_tags: '替换标签',
    add_tags: '追加标签',
    remove_tags: '移除标签',
    set_endpoints: '替换端点',
    clear_pricing: '清空价格覆盖',
  }
  return labels[action]
}

function errorMessage(error: unknown, fallback: string): string {
  if (typeof error === 'object' && error && 'message' in error) {
    return String((error as { message?: unknown }).message || fallback)
  }
  return fallback
}

onMounted(loadAll)
</script>
