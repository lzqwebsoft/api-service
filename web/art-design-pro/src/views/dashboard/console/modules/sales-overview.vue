<template>
  <div class="art-card h-105 p-5 mb-5 max-sm:mb-4" v-loading="props.loading">
    <div class="art-card-header">
      <div class="title">
        <h4>每日 Token 访问趋势 (最近 7 天)</h4>
      </div>
    </div>
    <ArtLineChart
      height="calc(100% - 56px)"
      :data="chartData"
      :xAxisData="xAxisData"
      :showAreaColor="true"
      :showAxisLine="false"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import type { TrendItem } from '@/api/dashboard'

  defineOptions({ name: 'SalesOverview' })

  const props = defineProps<{
    trend: TrendItem[]
    loading: boolean
  }>()

  const chartData = computed(() => props.trend.map((item) => item.count))

  const xAxisData = computed(() => {
    return props.trend.map((item) => {
      // Format to MM-DD
      const parts = item.date.split('-')
      return parts.length >= 3 ? `${parts[1]}-${parts[2]}` : item.date
    })
  })
</script>
