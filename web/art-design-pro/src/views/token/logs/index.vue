<template>
  <div class="logs-page flex flex-col gap-4 pb-5">
    <ElCard class="art-table-card">
      <template #header>
        <div class="flex-cb">
          <div class="title-group">
            <h4 class="m-0 font-semibold text-lg">{{ t('logsManage.title') }}</h4>
            <p class="text-gray-400 text-xs mt-1">{{ t('logsManage.subtitle') }}</p>
          </div>
          <div class="flex gap-2">
            <ElTag type="success">
              {{ t('logsManage.dataCount', { count: pagination.total }) }}
            </ElTag>
          </div>
        </div>
      </template>

      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadLogs" />

      <!-- Table of Access Logs -->
      <ArtTable
        v-loading="loading"
        :data="logs"
        :columns="columns"
        :pagination="pagination"
        :height="computedTableHeight"
        empty-height="360px"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
        <!-- Custom slot: token -->
        <template #token="{ row }">
          <span class="font-mono text-xs text-gray-600 dark:text-gray-400">{{ row.token }}</span>
        </template>

        <!-- Custom slot: platform -->
        <template #platform="{ row }">
          <ElTooltip :content="row.platform" placement="top">
            <ElTag
              :type="getPlatformTagType(row.platform)"
              size="small"
              class="capitalize font-medium cursor-pointer"
            >
              <ArtSvgIcon :icon="getPlatformIcon(row.platform)" class="text-xs" />
            </ElTag>
          </ElTooltip>
        </template>

        <!-- Custom slot: version -->
        <template #version="{ row }">
          <ElTag size="small" type="info">{{ row.version }}</ElTag>
        </template>

        <!-- Custom slot: user_uuid -->
        <template #user_uuid="{ row }">
          <span class="font-mono text-xs text-gray-700 dark:text-gray-300">{{
            row.user_uuid || '-'
          }}</span>
        </template>

        <!-- Custom slot: ip -->
        <template #ip="{ row }">
          <span class="font-mono text-xs text-gray-500">{{ row.ip || '-' }}</span>
        </template>

        <!-- Custom slot: createdAt -->
        <template #createdAt="{ row }">
          {{ formatTime(row.created_at) }}
        </template>

        <!-- Custom slot: operation -->
        <template #operation="{ row }">
          <div class="flex items-center justify-center">
            <ElTooltip
              v-if="!isBlacklisted(row)"
              :content="t('logsManage.blacklistBtn')"
              placement="top"
            >
              <ArtButtonTable
                icon="ri:forbid-line"
                icon-class="bg-error/12 text-error"
                @click="oneClickBlacklist(row)"
              />
            </ElTooltip>
            <ElTag v-else type="info">{{ t('logsManage.blocked') }}</ElTag>
          </div>
        </template>
      </ArtTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted, computed } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { fetchGetLogs, fetchAddLogBlacklist } from '@/api/token'

  defineOptions({ name: 'Logs' })

  const { t } = useI18n()

  const computedTableHeight = computed(() => {
    return ''
  })

  const loading = ref(false)
  const logs = ref<any[]>([])
  const blacklistedKeys = ref<Record<string, boolean>>({})

  // Pagination state
  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const { columns, columnChecks } = useTableColumns(() => [
    { type: 'globalIndex', label: t('logsManage.index'), width: 70, align: 'center' },
    { prop: 'app_name', label: t('logsManage.appName') || '应用名称', minWidth: 120 },
    {
      prop: 'token',
      label: t('logsManage.token'),
      minWidth: 140,
      showOverflowTooltip: true,
      useSlot: true,
      slotName: 'token'
    },
    {
      prop: 'version',
      label: t('logsManage.version'),
      width: 90,
      useSlot: true,
      slotName: 'version'
    },
    {
      prop: 'platform',
      label: t('logsManage.platform'),
      width: 90,
      align: 'center',
      useSlot: true,
      slotName: 'platform'
    },
    {
      prop: 'user_uuid',
      label: t('logsManage.userUuid'),
      minWidth: 140,
      showOverflowTooltip: true,
      useSlot: true,
      slotName: 'user_uuid'
    },
    {
      prop: 'ip',
      label: t('logsManage.ip'),
      width: 120,
      useSlot: true,
      slotName: 'ip'
    },
    {
      prop: 'ip_location',
      label: t('logsManage.ipLocation'),
      minWidth: 130,
      showOverflowTooltip: true
    },
    { prop: 'api_path', label: t('logsManage.apiPath'), minWidth: 160 },
    {
      prop: 'created_at',
      label: t('logsManage.createdAt'),
      width: 160,
      useSlot: true,
      slotName: 'createdAt'
    },
    {
      prop: 'operation',
      label: t('logsManage.operations'),
      width: 120,
      fixed: 'right',
      useSlot: true,
      slotName: 'operation'
    }
  ])

  onMounted(() => {
    loadLogs()
  })

  const loadLogs = async () => {
    loading.value = true
    try {
      const res = await fetchGetLogs({
        current: pagination.current,
        size: pagination.size
      })
      logs.value = res.list || []
      pagination.total = res.total || 0
      blacklistedKeys.value = res.blacklistedKeys || {}
    } catch (e: any) {
      ElMessage.error(e.message || t('logsManage.errorLoad'))
    } finally {
      loading.value = false
    }
  }

  const handleSizeChange = (val: number) => {
    pagination.size = val
    pagination.current = 1
    loadLogs()
  }

  const handleCurrentChange = (val: number) => {
    pagination.current = val
    loadLogs()
  }

  const isBlacklisted = (row: any) => {
    const key = `${row.token}:${row.user_uuid}`
    return !!blacklistedKeys.value[key]
  }

  const oneClickBlacklist = (row: any) => {
    ElMessageBox.confirm(t('logsManage.blacklistConfirm'), t('logsManage.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'error'
    }).then(async () => {
      try {
        await fetchAddLogBlacklist({
          token_id: row.token_id,
          token: row.token,
          user_uuid: row.user_uuid
        })
        ElMessage.success(t('logsManage.successBlock'))
        loadLogs()
      } catch (e: any) {
        ElMessage.error(e.message || t('logsManage.errorBlock'))
      }
    })
  }

  const formatTime = (timeStr: string) => {
    if (!timeStr) return '-'
    const d = new Date(timeStr)
    return d.toLocaleString()
  }

  const getPlatformIcon = (platform: string) => {
    const p = (platform || '').toLowerCase()
    if (p.includes('android')) return 'ri:android-fill'
    if (p.includes('ios') || p.includes('mac') || p.includes('apple')) return 'ri:apple-fill'
    if (p.includes('win')) return 'ri:windows-fill'
    if (p.includes('linux') || p.includes('ubuntu')) return 'ri:ubuntu-fill'
    return 'ri:computer-line'
  }

  const getPlatformTagType = (platform: string) => {
    const p = (platform || '').toLowerCase()
    if (p.includes('android')) return 'success'
    if (p.includes('ios') || p.includes('mac') || p.includes('apple')) return 'primary'
    if (p.includes('win')) return 'info'
    if (p.includes('linux')) return 'warning'
    return 'info'
  }
</script>

<style scoped></style>
