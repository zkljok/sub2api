<template>
  <AppLayout>
    <div class="donation-page">
      <div v-if="loading" class="flex h-full items-center justify-center">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <div v-else-if="!enabled || !safeDonationUrl" class="flex h-full items-center justify-center p-8 text-center">
        <div class="max-w-md">
          <div class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-primary-50 text-primary-500 dark:bg-primary-900/20 dark:text-primary-300">
            <Icon name="gift" size="lg" />
          </div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('donation.notConfiguredTitle') }}
          </h2>
          <p class="mt-2 text-sm text-gray-500 dark:text-dark-300">
            {{ t('donation.notConfiguredDesc') }}
          </p>
        </div>
      </div>

      <div v-else class="donation-embed-shell">
        <div class="donation-toolbar">
          <div class="min-w-0">
            <h1 class="truncate text-base font-semibold text-gray-900 dark:text-white">
              {{ t('donation.title') }}
            </h1>
            <p class="truncate text-xs text-gray-500 dark:text-dark-300">
              {{ t('donation.description') }}
            </p>
          </div>
          <a
            :href="safeDonationUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="btn btn-secondary btn-sm shrink-0"
          >
            <Icon name="externalLink" size="sm" class="mr-1.5" />
            {{ t('donation.openInNewTab') }}
          </a>
        </div>
        <iframe
          :src="safeDonationUrl"
          class="donation-embed-frame"
          sandbox="allow-forms allow-popups allow-popups-to-escape-sandbox allow-same-origin allow-scripts"
          referrerpolicy="strict-origin-when-cross-origin"
          allow="clipboard-write; payment"
        ></iframe>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)

const enabled = computed(() => appStore.cachedPublicSettings?.donation_enabled !== false)
const safeDonationUrl = computed(() =>
  sanitizeUrl(appStore.cachedPublicSettings?.donation_url || 'https://www.kufaka.com/shop/YJLink')
)

onMounted(async () => {
  if (appStore.publicSettingsLoaded) return
  loading.value = true
  try {
    await appStore.fetchPublicSettings()
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.donation-page {
  height: calc(100vh - 64px - 4rem);
  min-height: 520px;
}

.donation-embed-shell {
  display: flex;
  height: 100%;
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgb(229 231 235);
  border-radius: 0.75rem;
  background: white;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
}

.dark .donation-embed-shell {
  border-color: rgb(55 65 81);
  background: rgb(17 24 39);
}

.donation-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  border-bottom: 1px solid rgb(229 231 235);
  padding: 0.75rem 1rem;
  background: rgba(255, 255, 255, 0.92);
}

.dark .donation-toolbar {
  border-color: rgb(55 65 81);
  background: rgba(17, 24, 39, 0.92);
}

.donation-embed-frame {
  display: block;
  width: 100%;
  height: 100%;
  min-height: 0;
  flex: 1 1 auto;
  border: 0;
  background: white;
}
</style>
