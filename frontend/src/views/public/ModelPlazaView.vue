<template>
  <div class="min-h-screen bg-gray-50 text-gray-900 dark:bg-dark-950 dark:text-white">
    <header class="border-b border-gray-200 bg-white/90 backdrop-blur dark:border-dark-800 dark:bg-dark-900/90">
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
      <section class="mb-5 flex flex-col justify-between gap-4 lg:flex-row lg:items-end">
        <div>
          <h1 class="text-2xl font-semibold tracking-normal text-gray-950 dark:text-white">模型广场</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            共 {{ filteredModels.length }} 个公开可用模型，价格单位为美元。
          </p>
        </div>
        <button class="btn btn-secondary" :disabled="loading" @click="load">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </section>

      <section class="mb-5 grid gap-3 md:grid-cols-4">
        <div class="relative md:col-span-2">
          <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input v-model="query" class="input pl-10" placeholder="搜索模型、厂商、标签、分组" />
        </div>
        <select v-model="vendorFilter" class="input">
          <option value="">全部厂商</option>
          <option v-for="vendor in vendors" :key="vendor.id" :value="String(vendor.id)">{{ vendor.name }}</option>
        </select>
        <select v-model="endpointFilter" class="input">
          <option value="">全部端点</option>
          <option v-for="endpoint in endpoints" :key="endpoint" :value="endpoint">{{ endpoint }}</option>
        </select>
      </section>

      <section v-if="loading" class="card p-8 text-center text-sm text-gray-500 dark:text-gray-400">加载中...</section>
      <section v-else-if="filteredModels.length === 0" class="card p-8 text-center text-sm text-gray-500 dark:text-gray-400">暂无匹配模型</section>
      <section v-else class="grid gap-4 lg:grid-cols-2 xl:grid-cols-3">
        <article v-for="model in filteredModels" :key="model.model_name" class="card flex min-h-[260px] flex-col p-5">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <img v-if="model.icon" :src="model.icon" alt="" class="h-6 w-6 rounded object-cover" />
                <h2 class="truncate text-base font-semibold text-gray-950 dark:text-white">{{ model.display_name || model.model_name }}</h2>
              </div>
              <p class="mt-1 truncate font-mono text-xs text-gray-500 dark:text-gray-400">{{ model.model_name }}</p>
            </div>
            <span class="rounded bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-300">{{ model.platform || '-' }}</span>
          </div>

          <p class="mt-3 line-clamp-2 min-h-[2.5rem] text-sm text-gray-600 dark:text-gray-300">{{ model.description || '暂无模型说明' }}</p>

          <div class="mt-3 flex flex-wrap gap-1.5">
            <span v-for="tag in model.tags" :key="tag" class="rounded bg-primary-50 px-2 py-0.5 text-xs text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">{{ tag }}</span>
            <span v-if="vendorName(model.vendor_id)" class="rounded bg-emerald-50 px-2 py-0.5 text-xs text-emerald-700 dark:bg-emerald-900/25 dark:text-emerald-300">{{ vendorName(model.vendor_id) }}</span>
          </div>

          <div class="mt-4 grid grid-cols-2 gap-2 text-xs">
            <div class="rounded border border-gray-100 p-2 dark:border-dark-700">
              <div class="text-gray-400">输入</div>
              <div class="mt-1 font-semibold">{{ price(model.pricing?.input_price, 'MTok') }}</div>
            </div>
            <div class="rounded border border-gray-100 p-2 dark:border-dark-700">
              <div class="text-gray-400">输出</div>
              <div class="mt-1 font-semibold">{{ price(model.pricing?.output_price, 'MTok') }}</div>
            </div>
            <div class="rounded border border-gray-100 p-2 dark:border-dark-700">
              <div class="text-gray-400">缓存写入</div>
              <div class="mt-1 font-semibold">{{ price(model.pricing?.cache_write_price, 'MTok') }}</div>
            </div>
            <div class="rounded border border-gray-100 p-2 dark:border-dark-700">
              <div class="text-gray-400">缓存读取</div>
              <div class="mt-1 font-semibold">{{ price(model.pricing?.cache_read_price, 'MTok') }}</div>
            </div>
          </div>

          <div class="mt-3 flex flex-wrap gap-1.5">
            <span v-for="endpoint in model.supported_endpoints" :key="endpoint" class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-300">{{ endpoint }}</span>
          </div>
          <div class="mt-auto pt-4 text-xs text-gray-500 dark:text-gray-400">
            公开分组：{{ model.groups.map((g) => g.name).join('、') || '-' }}
          </div>
        </article>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import modelPlazaAPI, { type ModelPlazaModelView, type ModelPlazaVendor } from '@/api/modelPlaza'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const appStore = useAppStore()
const authStore = useAuthStore()

const models = ref<ModelPlazaModelView[]>([])
const vendors = ref<ModelPlazaVendor[]>([])
const endpoints = ref<string[]>([])
const query = ref('')
const vendorFilter = ref('')
const endpointFilter = ref('')
const loading = ref(false)

const dashboardPath = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))

const filteredModels = computed(() => {
  const q = query.value.trim().toLowerCase()
  return models.value.filter((model) => {
    if (vendorFilter.value && String(model.vendor_id || '') !== vendorFilter.value) return false
    if (endpointFilter.value && !model.supported_endpoints.includes(endpointFilter.value)) return false
    if (!q) return true
    const haystack = [
      model.model_name,
      model.display_name,
      model.description || '',
      vendorName(model.vendor_id),
      ...model.tags,
      ...model.groups.map((g) => g.name),
      ...model.supported_endpoints,
    ].join(' ').toLowerCase()
    return haystack.includes(q)
  })
})

async function load() {
  loading.value = true
  try {
    const snapshot = await modelPlazaAPI.getSnapshot()
    models.value = snapshot.models || []
    vendors.value = snapshot.vendors || []
    endpoints.value = snapshot.supported_endpoints || []
  } finally {
    loading.value = false
  }
}

function vendorName(id?: number | null): string {
  if (!id) return ''
  return vendors.value.find((vendor) => vendor.id === id)?.name || ''
}

function price(value: number | null | undefined, unit: string): string {
  if (value === null || value === undefined) return '-'
  return `$${Number(value).toLocaleString(undefined, { maximumFractionDigits: 6 })}/${unit}`
}

onMounted(load)
</script>
