<template>
  <div class="model-plaza min-h-screen bg-slate-50 text-slate-950 dark:bg-dark-950 dark:text-white">
    <header class="sticky top-0 z-20 border-b border-slate-200 bg-white/90 backdrop-blur-xl dark:border-dark-800 dark:bg-dark-900/90">
      <div class="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-8">
        <router-link to="/home" class="text-lg font-semibold">{{ appStore.siteName || 'Sub2API' }}</router-link>
        <div class="flex items-center gap-3">
          <router-link class="btn btn-secondary" :to="authStore.isAuthenticated ? dashboardPath : '/login'">
            {{ authStore.isAuthenticated ? '控制台' : '登录' }}
          </router-link>
        </div>
      </div>
    </header>

    <main class="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
      <section class="plaza-hero mb-5 overflow-hidden rounded-lg border border-slate-200 bg-white p-5 shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div class="max-w-3xl">
            <p class="text-xs font-semibold uppercase tracking-wide text-primary-600 dark:text-primary-300">Model Square</p>
            <h1 class="mt-2 text-3xl font-semibold tracking-normal text-slate-950 dark:text-white">模型广场</h1>
            <p class="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-300">
              汇总当前中转站公开可用模型。未登录用户只看到公开分组和基础价格信息，不展示用户专属倍率。
            </p>
          </div>
          <div class="grid grid-cols-3 gap-2 sm:min-w-[420px]">
            <div class="stat-tile">
              <div class="stat-value">{{ models.length }}</div>
              <div class="stat-label">公开模型</div>
            </div>
            <div class="stat-tile">
              <div class="stat-value">{{ activeVendorCount }}</div>
              <div class="stat-label">厂商</div>
            </div>
            <div class="stat-tile">
              <div class="stat-value">{{ endpoints.length }}</div>
              <div class="stat-label">端点</div>
            </div>
          </div>
        </div>
      </section>

      <section class="mb-5 rounded-lg border border-slate-200 bg-white p-4 shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_180px_180px_180px]">
          <div class="relative">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
            <input v-model="query" class="input h-11 pl-10" placeholder="搜索模型、厂商、标签、端点" />
          </div>
          <select v-model="vendorFilter" class="input h-11">
            <option value="">全部厂商</option>
            <option v-for="vendor in vendorOptions" :key="vendor.id" :value="String(vendor.id)">
              {{ vendor.name }} ({{ vendor.count }})
            </option>
          </select>
          <select v-model="endpointFilter" class="input h-11">
            <option value="">全部端点</option>
            <option v-for="endpoint in endpointOptions" :key="endpoint.value" :value="endpoint.value">
              {{ endpoint.label }} ({{ endpoint.count }})
            </option>
          </select>
          <button class="btn btn-secondary h-11" :disabled="loading" @click="load">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            刷新
          </button>
        </div>

        <div class="mt-4 grid gap-4 xl:grid-cols-3">
          <FilterBlock title="厂商" :items="vendorChipOptions" :active="vendorFilter" @select="vendorFilter = $event" />
          <FilterBlock title="标签" :items="tagOptions" :active="tagFilter" @select="tagFilter = $event" />
          <FilterBlock title="计费" :items="billingOptions" :active="billingFilter" @select="billingFilter = $event" />
        </div>
      </section>

      <section class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div class="text-sm text-slate-500 dark:text-slate-400">
          当前显示 <span class="font-semibold text-slate-900 dark:text-white">{{ filteredModels.length }}</span> 个模型
          <span v-if="snapshotVersion" class="ml-2 font-mono text-xs">#{{ snapshotVersion.slice(0, 8) }}</span>
        </div>
        <button v-if="hasActiveFilter" type="button" class="btn btn-secondary btn-sm" @click="resetFilters">清空筛选</button>
      </section>

      <section v-if="loading" class="rounded-lg border border-slate-200 bg-white p-8 text-center text-sm text-slate-500 dark:border-dark-700 dark:bg-dark-900 dark:text-slate-400">
        加载中...
      </section>
      <section v-else-if="filteredModels.length === 0" class="rounded-lg border border-slate-200 bg-white p-8 text-center text-sm text-slate-500 dark:border-dark-700 dark:bg-dark-900 dark:text-slate-400">
        暂无匹配模型
      </section>
      <section v-else class="grid gap-4 lg:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="model in filteredModels"
          :key="model.model_name"
          class="model-card group relative flex min-h-[250px] flex-col overflow-hidden rounded-lg border border-slate-200 bg-white p-4 shadow-sm transition dark:border-dark-700 dark:bg-dark-900"
        >
          <div class="pointer-events-none absolute inset-x-0 top-0 h-1 bg-gradient-to-r from-cyan-500 via-primary-500 to-emerald-500 opacity-0 transition group-hover:opacity-100" />
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <img v-if="model.icon" :src="model.icon" alt="" class="h-8 w-8 rounded object-cover" />
                <div v-else class="flex h-8 w-8 items-center justify-center rounded bg-slate-100 text-xs font-semibold text-slate-600 dark:bg-dark-700 dark:text-slate-300">
                  {{ modelInitial(model) }}
                </div>
                <h2 class="truncate text-base font-semibold text-slate-950 dark:text-white">{{ model.display_name || model.model_name }}</h2>
                <button type="button" class="copy-button" title="复制模型名称" @click="copyModelName(model.model_name)">
                  <Icon name="copy" size="sm" />
                </button>
              </div>
              <p class="mt-1 truncate font-mono text-xs text-slate-500 dark:text-slate-400">{{ model.model_name }}</p>
            </div>
            <span class="rounded bg-slate-100 px-2 py-1 text-xs font-medium text-slate-600 dark:bg-dark-700 dark:text-slate-300">
              {{ platformLabel(model.platform) }}
            </span>
          </div>

          <p v-if="model.description" class="mt-3 line-clamp-2 text-sm leading-5 text-slate-600 dark:text-slate-300">
            {{ model.description }}
          </p>

          <div class="mt-4 grid grid-cols-2 gap-2">
            <PriceTile label="Input" :value="model.pricing?.input_price" unit="/1M" />
            <PriceTile label="Output" :value="model.pricing?.output_price" unit="/1M" tone="strong" />
            <PriceTile label="Cache Write" :value="model.pricing?.cache_write_price" unit="/1M" />
            <PriceTile label="Cached" :value="model.pricing?.cache_read_price" unit="/1M" tone="accent" />
          </div>

          <div class="mt-4 flex flex-wrap gap-1.5">
            <span v-if="vendorName(model.vendor_id)" class="chip chip-emerald">{{ vendorName(model.vendor_id) }}</span>
            <span v-for="tag in model.tags" :key="tag" class="chip chip-primary">{{ tag }}</span>
            <span v-if="pricingSourceLabel(model.pricing_source)" class="chip chip-slate">{{ pricingSourceLabel(model.pricing_source) }}</span>
          </div>

          <div class="mt-3 flex flex-wrap gap-1.5">
            <span v-for="endpoint in model.supported_endpoints" :key="endpoint" class="endpoint-chip">{{ endpoint }}</span>
            <span v-if="model.supported_endpoints.length === 0" class="text-xs text-slate-400">未指定端点</span>
          </div>

          <div class="mt-auto pt-4 text-xs text-slate-500 dark:text-slate-400">
            计费：{{ billingLabel(model.pricing?.billing_mode || '') }}
          </div>
        </article>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, ref, type PropType } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import modelPlazaAPI, { type ModelPlazaModelView, type ModelPlazaVendor } from '@/api/modelPlaza'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

interface FilterItem {
  label: string
  value: string
  count: number
}

const FilterBlock = defineComponent({
  name: 'FilterBlock',
  props: {
    title: { type: String, required: true },
    active: { type: String, required: true },
    items: { type: Array as PropType<FilterItem[]>, required: true },
  },
  emits: ['select'],
  setup(props, { emit }) {
    return () =>
      h('div', { class: 'filter-block' }, [
        h('div', { class: 'filter-title' }, props.title),
        h('div', { class: 'filter-chips' }, [
          h(
            'button',
            {
              type: 'button',
              class: ['filter-chip', props.active === '' ? 'is-active' : ''],
              onClick: () => emit('select', ''),
            },
            ['全部'],
          ),
          ...props.items.slice(0, 12).map((item) =>
            h(
              'button',
              {
                type: 'button',
                class: ['filter-chip', props.active === item.value ? 'is-active' : ''],
                onClick: () => emit('select', item.value),
              },
              [h('span', item.label), h('span', { class: 'filter-count' }, String(item.count))],
            ),
          ),
        ]),
      ])
  },
})

const PriceTile = defineComponent({
  name: 'PriceTile',
  props: {
    label: { type: String, required: true },
    value: { type: Number as PropType<number | null | undefined>, default: null },
    unit: { type: String, required: true },
    tone: { type: String, default: '' },
  },
  setup(props) {
    return () =>
      h('div', { class: ['price-tile', props.tone ? `price-tile-${props.tone}` : ''] }, [
        h('div', { class: 'price-label' }, props.label),
        h('div', { class: 'price-value' }, price(props.value, props.unit)),
      ])
  },
})

const appStore = useAppStore()
const authStore = useAuthStore()

const models = ref<ModelPlazaModelView[]>([])
const vendors = ref<ModelPlazaVendor[]>([])
const endpoints = ref<string[]>([])
const query = ref('')
const vendorFilter = ref('')
const endpointFilter = ref('')
const tagFilter = ref('')
const billingFilter = ref('')
const loading = ref(false)
const snapshotVersion = ref('')

const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const activeVendorCount = computed(() => vendors.value.filter((vendor) => vendor.status === 'active').length)

const filteredModels = computed(() => {
  const q = query.value.trim().toLowerCase()
  return models.value.filter((model) => {
    if (vendorFilter.value && String(model.vendor_id || '') !== vendorFilter.value) return false
    if (endpointFilter.value && !model.supported_endpoints.includes(endpointFilter.value)) return false
    if (tagFilter.value && !model.tags.includes(tagFilter.value)) return false
    if (billingFilter.value && (model.pricing?.billing_mode || '') !== billingFilter.value) return false
    if (!q) return true
    const haystack = [
      model.model_name,
      model.display_name,
      model.description || '',
      vendorName(model.vendor_id),
      platformLabel(model.platform),
      pricingSourceLabel(model.pricing_source),
      ...model.tags,
      ...model.supported_endpoints,
    ].join(' ').toLowerCase()
    return haystack.includes(q)
  })
})

const hasActiveFilter = computed(() =>
  Boolean(query.value || vendorFilter.value || endpointFilter.value || tagFilter.value || billingFilter.value),
)

const vendorOptions = computed(() =>
  vendors.value.map((vendor) => ({
    ...vendor,
    count: models.value.filter((model) => model.vendor_id === vendor.id).length,
  })),
)

const vendorChipOptions = computed<FilterItem[]>(() =>
  vendorOptions.value
    .filter((vendor) => vendor.count > 0)
    .map((vendor) => ({ label: vendor.name, value: String(vendor.id), count: vendor.count })),
)

const endpointOptions = computed(() => countItems(models.value.flatMap((model) => model.supported_endpoints)))
const tagOptions = computed(() => countItems(models.value.flatMap((model) => model.tags)))
const billingOptions = computed(() =>
  countItems(models.value.map((model) => model.pricing?.billing_mode || '').filter(Boolean)).map((item) => ({
    ...item,
    label: billingLabel(item.value),
  })),
)

async function load() {
  loading.value = true
  try {
    const snapshot = await modelPlazaAPI.getSnapshot()
    models.value = snapshot.models || []
    vendors.value = snapshot.vendors || []
    endpoints.value = snapshot.supported_endpoints || []
    snapshotVersion.value = snapshot.pricing_version || ''
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  query.value = ''
  vendorFilter.value = ''
  endpointFilter.value = ''
  tagFilter.value = ''
  billingFilter.value = ''
}

function countItems(values: string[]): FilterItem[] {
  const counter = new Map<string, number>()
  for (const value of values) {
    const trimmed = String(value || '').trim()
    if (!trimmed) continue
    counter.set(trimmed, (counter.get(trimmed) || 0) + 1)
  }
  return Array.from(counter.entries())
    .map(([value, count]) => ({ label: value, value, count }))
    .sort((a, b) => b.count - a.count || a.label.localeCompare(b.label))
}

async function copyModelName(modelName: string) {
  try {
    await navigator.clipboard.writeText(modelName)
    appStore.showSuccess('模型名称已复制')
  } catch {
    appStore.showError('复制失败')
  }
}

function vendorName(id?: number | null): string {
  if (!id) return ''
  return vendors.value.find((vendor) => vendor.id === id)?.name || ''
}

function price(value: number | null | undefined, unit: string): string {
  if (value === null || value === undefined) return '-'
  return `$${Number(value).toLocaleString(undefined, { maximumFractionDigits: 6 })}${unit}`
}

function modelInitial(model: ModelPlazaModelView): string {
  return (model.display_name || model.model_name || 'M').slice(0, 1).toUpperCase()
}

function platformLabel(platform: string): string {
  const labels: Record<string, string> = {
    openai: 'OpenAI',
    anthropic: 'Anthropic',
    gemini: 'Gemini',
    grok: 'Grok',
    antigravity: 'Antigravity',
  }
  return labels[platform] || platform || '-'
}

function billingLabel(mode: string): string {
  const labels: Record<string, string> = {
    token: 'Token',
    per_request: '按次',
    image: '图片',
    video: '视频',
  }
  return labels[mode] || '未配置'
}

function pricingSourceLabel(source: string): string {
  const labels: Record<string, string> = {
    account: '账号同步',
    channel: '渠道价格',
    override: '管理员覆盖',
    official_preset: '官方预设',
  }
  return labels[source] || ''
}

onMounted(load)
</script>

<style scoped>
.plaza-hero {
  background:
    linear-gradient(135deg, rgba(14, 165, 233, 0.08), rgba(16, 185, 129, 0.08)),
    var(--tw-bg-opacity, #fff);
}

.stat-tile {
  border: 1px solid rgb(226 232 240);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.76);
  padding: 14px;
}

.dark .stat-tile {
  border-color: rgb(55 65 81);
  background: rgba(17, 24, 39, 0.72);
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  line-height: 1;
}

.stat-label {
  margin-top: 6px;
  font-size: 12px;
  color: rgb(100 116 139);
}

.filter-block {
  min-width: 0;
}

.filter-title {
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 600;
  color: rgb(71 85 105);
}

.dark .filter-title {
  color: rgb(203 213 225);
}

.filter-chips {
  display: flex;
  max-height: 74px;
  flex-wrap: wrap;
  gap: 6px;
  overflow: hidden;
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 1px solid rgb(226 232 240);
  border-radius: 8px;
  background: rgb(248 250 252);
  padding: 5px 8px;
  font-size: 12px;
  color: rgb(71 85 105);
  transition: all 160ms ease;
}

.filter-chip:hover,
.filter-chip.is-active {
  border-color: rgb(14 165 233);
  background: rgb(240 249 255);
  color: rgb(3 105 161);
}

.dark .filter-chip {
  border-color: rgb(55 65 81);
  background: rgb(31 41 55);
  color: rgb(203 213 225);
}

.dark .filter-chip:hover,
.dark .filter-chip.is-active {
  border-color: rgb(56 189 248);
  background: rgba(14, 165, 233, 0.14);
  color: rgb(186 230 253);
}

.filter-count {
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.08);
  padding: 0 6px;
}

.model-card:hover {
  border-color: rgba(14, 165, 233, 0.72);
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.12);
  transform: translateY(-3px);
}

.price-tile {
  border: 1px solid rgb(226 232 240);
  border-radius: 8px;
  background: rgb(248 250 252);
  padding: 8px;
  transition: all 160ms ease;
}

.model-card:hover .price-tile {
  border-color: rgba(14, 165, 233, 0.32);
}

.price-tile-strong {
  background: rgba(14, 165, 233, 0.06);
}

.price-tile-accent {
  background: rgba(16, 185, 129, 0.08);
}

.dark .price-tile {
  border-color: rgb(55 65 81);
  background: rgba(31, 41, 55, 0.76);
}

.price-label {
  font-size: 11px;
  color: rgb(100 116 139);
}

.price-value {
  margin-top: 3px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  font-weight: 650;
  color: rgb(15 23 42);
}

.dark .price-value {
  color: rgb(248 250 252);
}

.chip,
.endpoint-chip {
  border-radius: 8px;
  padding: 3px 8px;
  font-size: 12px;
  line-height: 1.4;
}

.chip-primary {
  background: rgb(239 246 255);
  color: rgb(29 78 216);
}

.chip-emerald {
  background: rgb(236 253 245);
  color: rgb(4 120 87);
}

.chip-slate,
.endpoint-chip {
  background: rgb(241 245 249);
  color: rgb(71 85 105);
}

.dark .chip-primary {
  background: rgba(59, 130, 246, 0.18);
  color: rgb(147 197 253);
}

.dark .chip-emerald {
  background: rgba(16, 185, 129, 0.18);
  color: rgb(110 231 183);
}

.dark .chip-slate,
.dark .endpoint-chip {
  background: rgb(31 41 55);
  color: rgb(203 213 225);
}

.copy-button {
  display: inline-flex;
  height: 26px;
  width: 26px;
  flex: 0 0 26px;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border: 1px solid rgb(226 232 240);
  color: rgb(100 116 139);
  opacity: 0.72;
  transition: all 160ms ease;
}

.copy-button:hover {
  border-color: rgb(14 165 233);
  background: rgb(240 249 255);
  color: rgb(3 105 161);
  opacity: 1;
}

.dark .copy-button {
  border-color: rgb(55 65 81);
  color: rgb(203 213 225);
}
</style>
