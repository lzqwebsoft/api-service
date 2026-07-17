<template>
  <ElRow :gutter="20" class="flex" v-loading="props.loading">
    <ElCol v-for="(item, index) in computedCards" :key="index" :sm="24" :md="8" :lg="8">
      <div class="art-card relative flex flex-col justify-center h-35 px-5 mb-5 max-sm:mb-4">
        <span class="text-g-700 text-sm">{{ item.des }}</span>
        <ArtCountTo class="text-[26px] font-medium mt-2" :target="item.num" :duration="1300" />
        <div
          class="absolute top-0 bottom-0 right-5 m-auto size-12.5 rounded-xl flex-cc bg-theme/10"
        >
          <ArtSvgIcon :icon="item.icon" class="text-xl text-theme" />
        </div>
      </div>
    </ElCol>
  </ElRow>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import type { DashboardStats } from '@/api/dashboard'

  defineOptions({ name: 'CardList' })

  const props = defineProps<{
    stats: DashboardStats
    loading: boolean
  }>()

  const computedCards = computed(() => [
    {
      des: '应用版本总数',
      icon: 'ri:archive-line',
      num: props.stats.totalApps
    },
    {
      des: '启用中版本',
      icon: 'ri:play-circle-line',
      num: props.stats.activeApps
    },
    {
      des: '全局已发放 Token',
      icon: 'ri:key-2-line',
      num: props.stats.totalTokens
    }
  ])
</script>
