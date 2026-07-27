<template>
  <div class="model-square min-h-screen bg-[#f7fafc] text-zinc-950">
    <header class="sticky top-0 z-30 border-b border-transparent bg-[#f7fafc]/90 backdrop-blur-xl">
      <div class="mx-auto flex h-16 max-w-[1080px] items-center justify-between rounded-b-[28px] border border-zinc-200/70 bg-white/90 px-4 shadow-sm">
        <router-link to="/home" class="flex items-center gap-3 font-semibold">
          <span class="brand-mark">{{ siteInitial }}</span>
          <span>{{ appStore.siteName || 'Sub2API' }}</span>
        </router-link>
        <nav class="hidden items-center gap-8 text-sm font-medium text-zinc-500 md:flex">
          <router-link class="nav-link" to="/home">主页</router-link>
          <router-link class="nav-link" :to="authStore.isAuthenticated ? dashboardPath : '/login'">控制台</router-link>
          <span class="text-zinc-950">模型广场</span>
        </nav>
        <router-link class="login-pill" :to="authStore.isAuthenticated ? dashboardPath : '/login'">
          {{ authStore.isAuthenticated ? userInitial : '登录' }}
        </router-link>
      </div>
    </header>

    <main class="mx-auto max-w-[1880px] px-5 py-6">
      <section class="mx-auto mb-10 max-w-[840px]">
        <div class="relative">
          <Icon name="search" size="md" class="absolute left-5 top-1/2 -translate-y-1/2 text-zinc-400" />
          <input
            v-model="query"
            class="h-14 w-full rounded-[20px] border border-zinc-200 bg-white px-14 text-base text-zinc-800 shadow-sm outline-none transition placeholder:text-zinc-400 focus:border-sky-300 focus:ring-4 focus:ring-sky-100"
            placeholder="搜索模型名称、供应商、端点或标签..."
          />
          <span class="absolute right-4 top-1/2 hidden -translate-y-1/2 rounded-lg border border-zinc-200 bg-zinc-50 px-2 py-1 text-xs text-zinc-400 sm:inline-flex">⌘K</span>
        </div>
      </section>

      <div class="model-plaza-layout" :class="{ 'filters-collapsed': filterCollapsed }">
        <aside class="filter-panel" :class="{ collapsed: filterCollapsed }">
          <div class="mb-6 flex items-start justify-between gap-3">
            <div v-if="!filterCollapsed">
              <h2 class="text-lg font-bold">筛选</h2>
              <p class="mt-1 text-sm text-zinc-500">按模型供应商细化模型。</p>
            </div>
            <button
              type="button"
              class="collapse-button"
              :title="filterCollapsed ? '展开筛选' : '收起筛选'"
              @click="filterCollapsed = !filterCollapsed"
            >
              <Icon :name="filterCollapsed ? 'chevronRight' : 'chevronLeft'" size="sm" />
            </button>
            <button v-if="!filterCollapsed" type="button" class="reset-button" :disabled="!hasActiveFilter" @click="resetFilters">
              <Icon name="refresh" size="sm" />
              重置
            </button>
          </div>

          <div v-if="filterCollapsed" class="collapsed-filter-tools">
            <button type="button" class="mini-filter-button" :class="{ active: !vendorFilter }" title="所有供应商" @click="vendorFilter = ''">
              <span>•</span>
              <b>{{ models.length }}</b>
            </button>
            <button
              v-for="item in vendorChipOptions.slice(0, 8)"
              :key="item.value"
              type="button"
              class="mini-filter-button"
              :class="{ active: vendorFilter === item.value }"
              :style="vendorStyle(item.label)"
              :title="item.label"
              @click="vendorFilter = item.value"
            >
              <span>{{ vendorIcon(item.label) }}</span>
              <b>{{ item.count }}</b>
            </button>
          </div>
          <FilterSection v-else title="所有供应商" :items="vendorChipOptions" :active="vendorFilter" @select="vendorFilter = $event" />
        </aside>

        <section class="min-w-0">
          <div class="toolbar">
            <div class="text-base font-semibold">
              {{ filteredModels.length }} <span class="font-normal text-zinc-500">个模型</span>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <div class="segmented">
                <button type="button" :class="{ active: unitMode === 'million' }" @click="unitMode = 'million'">/1M</button>
                <button type="button" :class="{ active: unitMode === 'thousand' }" @click="unitMode = 'thousand'">/1K</button>
              </div>
              <button type="button" class="sort-pill" @click="sortAsc = !sortAsc">
                <span>{{ sortAsc ? '↑' : '↓' }}</span>
                名称
              </button>
              <button type="button" class="icon-pill active" title="卡片视图">
                <Icon name="grid" size="sm" />
              </button>
            </div>
          </div>

          <div v-if="loading" class="empty-state">加载中...</div>
          <div v-else-if="filteredModels.length === 0" class="empty-state">暂无匹配模型</div>
          <div v-else class="grid gap-5 xl:grid-cols-2 2xl:grid-cols-3">
            <article v-for="model in sortedModels" :key="model.model_name" class="model-card">
              <div class="flex items-start gap-4">
                <div class="vendor-mark" :style="vendorStyle(vendorName(model.vendor_id) || model.platform || model.model_name)">
                  {{ vendorIcon(vendorName(model.vendor_id) || model.platform || model.model_name) }}
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <h3 class="truncate font-mono text-lg font-bold tracking-normal">{{ model.display_name || model.model_name }}</h3>
                      <div class="mt-2 space-y-0.5 text-sm">
                        <div>
                          <span class="text-zinc-500">输入</span>
                          <span class="ml-1 font-bold">{{ price(model.pricing?.input_price) }}</span>
                        </div>
                        <div>
                          <span class="text-zinc-500">输出</span>
                          <span class="ml-1 font-bold">{{ price(model.pricing?.output_price) }}</span>
                        </div>
                        <div v-if="model.pricing?.cache_write_price != null" class="text-zinc-400">
                          缓存写入 {{ price(model.pricing?.cache_write_price) }}
                        </div>
                        <div v-if="model.pricing?.cache_read_price != null" class="text-zinc-400">
                          缓存读取 {{ price(model.pricing?.cache_read_price) }}
                        </div>
                        <div v-if="model.pricing?.image_input_price != null" class="text-zinc-400">
                          图片输入 {{ price(model.pricing?.image_input_price) }}
                        </div>
                        <div v-if="model.pricing?.per_request_price != null" class="text-zinc-400">
                          按次 {{ requestPrice(model.pricing?.per_request_price) }}
                        </div>
                      </div>
                    </div>
                    <button type="button" class="copy-icon" title="复制模型名称" @click="copyModelName(model.model_name)">
                      <Icon name="copy" size="sm" />
                    </button>
                  </div>
                </div>
              </div>

              <p v-if="model.description" class="mt-7 line-clamp-2 min-h-[3rem] text-base leading-6 text-zinc-600">
                {{ model.description }}
              </p>

              <div class="mt-auto flex flex-wrap items-center gap-x-5 gap-y-2 pt-8 text-sm text-zinc-500">
                <span>{{ endpointLabel(model) }}</span>
                <span v-if="primaryTag(model)">{{ primaryTag(model) }}</span>
                <span>{{ pricingSourceLabel(model.pricing_source) }}</span>
              </div>
            </article>
          </div>
        </section>
      </div>
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

const FilterSection = defineComponent({
  name: 'FilterSection',
  props: {
    title: { type: String, required: true },
    active: { type: String, required: true },
    items: { type: Array as PropType<FilterItem[]>, required: true },
  },
  emits: ['select'],
  setup(props, { emit }) {
    return () =>
      h('section', { class: 'filter-section' }, [
        h('div', { class: 'filter-section-title' }, [
          h('span', props.title),
          h('span', { class: 'chevron' }, '⌃'),
        ]),
        h('div', { class: 'filter-chip-wrap' }, [
          h(
            'button',
            {
              type: 'button',
              class: ['filter-chip', props.active === '' ? 'active' : ''],
              onClick: () => emit('select', ''),
            },
            [h('i', { class: 'vendor-chip-icon all' }, '•'), h('span', '所有供应商'), h('b', totalCount(props.items))],
          ),
          ...props.items.slice(0, 14).map((item) =>
            h(
              'button',
              {
                type: 'button',
                class: ['filter-chip', props.active === item.value ? 'active' : ''],
                style: vendorStyle(item.label),
                onClick: () => emit('select', item.value),
              },
              [h('i', { class: 'vendor-chip-icon' }, vendorIcon(item.label)), h('span', item.label), h('b', String(item.count))],
            ),
          ),
        ]),
      ])
  },
})

const appStore = useAppStore()
const authStore = useAuthStore()

const models = ref<ModelPlazaModelView[]>([])
const vendors = ref<ModelPlazaVendor[]>([])
const query = ref('')
const vendorFilter = ref('')
const loading = ref(false)
const sortAsc = ref(true)
const filterCollapsed = ref(false)
const unitMode = ref<'million' | 'thousand'>('million')

const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
const siteInitial = computed(() => (appStore.siteName || 'S').slice(0, 1).toUpperCase())
const userInitial = computed(() => 'J')

const filteredModels = computed(() => {
  const q = query.value.trim().toLowerCase()
  return models.value.filter((model) => {
    if (vendorFilter.value && String(model.vendor_id || '') !== vendorFilter.value) return false
    if (!q) return true
    const haystack = [
      model.model_name,
      model.display_name,
      model.description || '',
      vendorName(model.vendor_id),
      ...model.tags,
      ...model.groups.map((group) => group.name),
      ...model.supported_endpoints,
    ].join(' ').toLowerCase()
    return haystack.includes(q)
  })
})

const sortedModels = computed(() =>
  [...filteredModels.value].sort((a, b) => {
    const result = (a.display_name || a.model_name).localeCompare(b.display_name || b.model_name)
    return sortAsc.value ? result : -result
  }),
)

const hasActiveFilter = computed(() => Boolean(query.value || vendorFilter.value))

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
async function load() {
  loading.value = true
  try {
    const snapshot = await modelPlazaAPI.getSnapshot()
    models.value = snapshot.models || []
    vendors.value = snapshot.vendors || []
  } catch (error) {
    appStore.showError('加载模型广场失败')
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  query.value = ''
  vendorFilter.value = ''
}

function totalCount(items: FilterItem[]): string {
  return String(models.value.length || items.reduce((sum, item) => sum + item.count, 0))
}

function vendorName(id?: number | null): string {
  if (!id) return ''
  return vendors.value.find((vendor) => vendor.id === id)?.name || ''
}

interface VendorVisual {
  icon: string
  color: string
  bg: string
}

const vendorVisuals: Record<string, VendorVisual> = {
  openai: { icon: '◎', color: '#111827', bg: '#f3f4f6' },
  deepseek: { icon: 'D', color: '#2563eb', bg: '#eff6ff' },
  anthropic: { icon: '✣', color: '#d97706', bg: '#fff7ed' },
  google: { icon: 'G', color: '#16a34a', bg: '#f0fdf4' },
  xai: { icon: 'X', color: '#4b5563', bg: '#f8fafc' },
  阿里巴巴: { icon: 'A', color: '#f97316', bg: '#fff7ed' },
  智谱: { icon: 'Z', color: '#6366f1', bg: '#eef2ff' },
  moonshot: { icon: 'M', color: '#64748b', bg: '#f8fafc' },
  字节豆包: { icon: '豆', color: '#7c3aed', bg: '#f5f3ff' },
  meta: { icon: '∞', color: '#2563eb', bg: '#eff6ff' },
  mistral: { icon: 'M', color: '#dc2626', bg: '#fef2f2' },
  cohere: { icon: 'C', color: '#0891b2', bg: '#ecfeff' },
}

function vendorVisual(name: string): VendorVisual {
  return vendorVisuals[name.toLowerCase()] || vendorVisuals[name] || { icon: name.slice(0, 1).toUpperCase(), color: '#df714b', bg: '#fff3ec' }
}

function vendorIcon(name: string): string {
  return vendorVisual(name).icon
}

function vendorStyle(name: string): Record<string, string> {
  const visual = vendorVisual(name)
  return {
    '--vendor-color': visual.color,
    '--vendor-bg': visual.bg,
  }
}

function price(value: number | null | undefined): string {
  if (value == null) return '-'
  const normalized = unitMode.value === 'thousand' ? value / 1000 : value
  const suffix = unitMode.value === 'thousand' ? '/1K' : '/1M'
  return `$${Number(normalized).toLocaleString(undefined, { maximumFractionDigits: 6 })}${suffix}`
}

function requestPrice(value: number | null | undefined): string {
  if (value == null) return '-'
  return `$${Number(value).toLocaleString(undefined, { maximumFractionDigits: 6 })}/次`
}

function primaryTag(model: ModelPlazaModelView): string {
  return model.tags[0] || ''
}

function endpointLabel(model: ModelPlazaModelView): string {
  return model.supported_endpoints[0] || model.platform || 'openai'
}

function pricingSourceLabel(source: string): string {
  const labels: Record<string, string> = {
    account: '按量计费',
    channel: '渠道价格',
    override: '管理员价格',
    official_preset: '官方预设',
  }
  return labels[source] || '按量计费'
}

async function copyModelName(modelName: string) {
  const copied = await copyText(modelName)
  if (copied) {
    appStore.showSuccess('模型名称已复制')
  } else {
    appStore.showError('复制失败，请手动选择模型名称')
  }
}

async function copyText(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
      return true
    }
  } catch {
    // Fall back below.
  }
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.left = '-9999px'
  document.body.appendChild(textarea)
  textarea.select()
  let success = false
  try {
    success = document.execCommand('copy')
  } finally {
    document.body.removeChild(textarea)
  }
  return success
}

onMounted(load)
</script>

<style scoped>
.brand-mark {
  display: inline-flex;
  height: 38px;
  width: 38px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #111827;
  color: #fff;
  font-weight: 800;
}

.nav-link {
  transition: color 150ms ease;
}

.nav-link:hover {
  color: #111827;
}

.login-pill {
  display: inline-flex;
  min-width: 36px;
  height: 36px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #a3c83b;
  padding: 0 12px;
  color: white;
  font-weight: 700;
}

.model-plaza-layout {
  display: grid;
  gap: 20px;
}

@media (min-width: 1024px) {
  .model-plaza-layout {
    grid-template-columns: 340px minmax(0, 1fr);
    transition: grid-template-columns 180ms ease;
  }

  .model-plaza-layout.filters-collapsed {
    grid-template-columns: 78px minmax(0, 1fr);
  }
}

.filter-panel {
  position: sticky;
  top: 88px;
  max-height: calc(100vh - 110px);
  overflow: hidden auto;
  border: 1px solid #e5e7eb;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.92);
  padding: 22px;
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.06);
  scrollbar-width: thin;
  scrollbar-color: #d4d4d8 transparent;
  transition:
    padding 180ms ease,
    border-radius 180ms ease,
    box-shadow 180ms ease;
}

.filter-panel.collapsed {
  padding: 14px 10px;
}

.filter-panel::-webkit-scrollbar {
  width: 8px;
}

.filter-panel::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: #d4d4d8;
}

.reset-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #71717a;
  font-size: 14px;
  transition: color 150ms ease;
}

.reset-button:not(:disabled):hover {
  color: #111827;
}

.reset-button:disabled {
  opacity: 0.45;
}

.collapse-button {
  display: inline-flex;
  height: 36px;
  width: 36px;
  flex: 0 0 36px;
  align-items: center;
  justify-content: center;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  background: #fff;
  color: #71717a;
  transition: all 150ms ease;
}

.collapse-button:hover {
  border-color: #bae6fd;
  background: #f0f9ff;
  color: #0284c7;
}

.collapsed-filter-tools {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: center;
}

.mini-filter-button {
  display: inline-flex;
  min-height: 46px;
  width: 52px;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  border: 1px solid #e5e7eb;
  border-radius: 18px;
  background: var(--vendor-bg, #fff);
  color: var(--vendor-color, #52525b);
  font-size: 12px;
  font-weight: 800;
  transition: all 150ms ease;
}

.mini-filter-button span {
  line-height: 1;
}

.mini-filter-button b {
  color: #71717a;
  font-size: 11px;
  line-height: 1;
}

.mini-filter-button:hover,
.mini-filter-button.active {
  border-color: color-mix(in srgb, var(--vendor-color, #71717a) 34%, #cdd3da);
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.08);
  transform: translateY(-1px);
}

:deep(.filter-section) {
  border-top: 1px solid #eceff3;
  padding: 20px 0;
}

:deep(.filter-section:first-of-type) {
  border-top: 0;
  padding-top: 0;
}

:deep(.filter-section-title) {
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 800;
}

:deep(.chevron) {
  color: #71717a;
  font-size: 16px;
}

:deep(.filter-chip-wrap) {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

:deep(.filter-chip) {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  background: var(--vendor-bg, #fff);
  padding: 8px 11px;
  color: #666;
  font-size: 15px;
  font-weight: 600;
  transition: all 150ms ease;
}

:deep(.vendor-chip-icon) {
  display: inline-flex;
  height: 22px;
  width: 22px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #fff;
  color: var(--vendor-color, #52525b);
  font-size: 12px;
  font-style: normal;
  font-weight: 900;
  box-shadow: 0 4px 10px rgba(15, 23, 42, 0.08);
}

:deep(.vendor-chip-icon.all) {
  color: #71717a;
}

:deep(.filter-chip b) {
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.78);
  padding: 1px 8px;
  color: #777;
}

:deep(.filter-chip:hover),
:deep(.filter-chip.active) {
  border-color: color-mix(in srgb, var(--vendor-color, #71717a) 34%, #cdd3da);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--vendor-color, #71717a) 18%, #d8dde3), 0 2px 8px rgba(15, 23, 42, 0.08);
  color: #222;
  transform: translateY(-1px);
}

.toolbar {
  margin-bottom: 20px;
  display: flex;
  min-height: 74px;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border: 1px solid #e5e7eb;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.86);
  padding: 16px;
  box-shadow: 0 16px 38px rgba(15, 23, 42, 0.05);
}

.segmented {
  display: inline-flex;
  gap: 3px;
  border-radius: 18px;
  background: #f4f4f5;
  padding: 3px;
}

.segmented button,
.sort-pill,
.icon-pill {
  display: inline-flex;
  min-height: 40px;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border-radius: 15px;
  padding: 0 14px;
  color: #52525b;
  font-weight: 700;
  transition: all 150ms ease;
}

.segmented button.active,
.icon-pill.active {
  background: #37a8eb;
  color: #fff;
  box-shadow: 0 8px 18px rgba(55, 168, 235, 0.28);
}

.sort-pill,
.icon-pill {
  border: 1px solid #e5e7eb;
  background: #fff;
}

.model-card {
  display: flex;
  min-height: 288px;
  flex-direction: column;
  border: 1px solid #e1e5ea;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.94);
  padding: 26px;
  box-shadow: 0 16px 34px rgba(15, 23, 42, 0.045);
  transition:
    transform 180ms ease,
    border-color 180ms ease,
    box-shadow 180ms ease;
  animation: card-in 360ms ease both;
}

.model-card:hover {
  transform: translateY(-5px);
  border-color: #bad3e8;
  box-shadow: 0 28px 60px rgba(15, 23, 42, 0.12);
}

.vendor-mark {
  display: inline-flex;
  height: 54px;
  width: 54px;
  flex: 0 0 54px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: var(--vendor-bg, #fff3ec);
  color: var(--vendor-color, #df714b);
  font-size: 28px;
  font-weight: 900;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.9), 0 14px 24px color-mix(in srgb, var(--vendor-color, #df714b) 12%, transparent);
}

.copy-icon {
  display: inline-flex;
  height: 40px;
  width: 40px;
  flex: 0 0 40px;
  align-items: center;
  justify-content: center;
  border: 1px solid #e5e7eb;
  border-radius: 15px;
  color: #71717a;
  transition: all 150ms ease;
}

.copy-icon:hover {
  border-color: #37a8eb;
  background: #eff8ff;
  color: #1288d6;
  transform: translateY(-1px);
}

.empty-state {
  border: 1px solid #e5e7eb;
  border-radius: 28px;
  background: #fff;
  padding: 60px 24px;
  text-align: center;
  color: #71717a;
}

@keyframes card-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
