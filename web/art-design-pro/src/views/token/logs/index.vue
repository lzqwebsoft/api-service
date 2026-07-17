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
        <!-- Custom slot: version -->
        <template #version="{ row }">
          <ElTag size="small">{{ row.version }}</ElTag>
        </template>

        <!-- Custom slot: createdAt -->
        <template #createdAt="{ row }">
          {{ formatTime(row.created_at) }}
        </template>

        <!-- Custom slot: operation -->
        <template #operation="{ row }">
          <ElTooltip v-if="!isBlacklisted(row)" :content="t('logsManage.blacklistBtn')" placement="top">
            <ArtButtonTable icon="ri:forbid-line" icon-class="bg-error/12 text-error" @click="oneClickBlacklist(row)" />
          </ElTooltip>
          <ElTag v-else type="info">{{ t('logsManage.blocked') }}</ElTag>
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

  // 计算实际的表格高度
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

  // Use the useTableColumns hook to manage visible/hidden columns, table checks and icons
  const { columns, columnChecks } = useTableColumns(() => [
    { type: 'globalIndex', label: t('logsManage.index'), width: 80, align: 'center' },
    { prop: 'token', label: t('logsManage.token'), minWidth: 150, showOverflowTooltip: true },
    { prop: 'platform', label: t('logsManage.platform'), width: 100 },
    {
      prop: 'version',
      label: t('logsManage.version'),
      width: 100,
      useSlot: true,
      slotName: 'version'
    },
    {
      prop: 'user_uuid',
      label: t('logsManage.userUuid'),
      minWidth: 150,
      showOverflowTooltip: true
    },
    { prop: 'ip', label: t('logsManage.ip'), width: 130 },
    { prop: 'api_path', label: t('logsManage.apiPath'), minWidth: 180 },
    {
      prop: 'created_at',
      label: t('logsManage.createdAt'),
      width: 180,
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
          token: row.token,
          platform: row.platform,
          version: row.version,
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
</script>

<style scoped></style>
