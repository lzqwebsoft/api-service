<!-- 工作台页面 -->
<template>
  <div>
    <CardList :stats="stats" :loading="loading" />

    <ElRow :gutter="20" class="mt-5">
      <ElCol :span="24">
        <SalesOverview :trend="stats.trend" :loading="loading" />
      </ElCol>
    </ElRow>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import CardList from './modules/card-list.vue'
  import SalesOverview from './modules/sales-overview.vue'
  import { fetchGetDashboardStats, type DashboardStats } from '@/api/dashboard'

  defineOptions({ name: 'Console' })

  const loading = ref(true)
  const stats = ref<DashboardStats>({
    totalApps: 0,
    activeApps: 0,
    totalTokens: 0,
    trend: []
  })

  onMounted(async () => {
    try {
      const res = await fetchGetDashboardStats()
      if (res) {
        stats.value = res
      }
    } catch (error) {
      console.error('Failed to fetch dashboard stats:', error)
    } finally {
      loading.value = false
    }
  })
</script>
