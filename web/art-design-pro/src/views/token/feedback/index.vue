<template>
  <div class="feedback-page flex flex-col gap-4 pb-5">
    <ElCard class="art-table-card">
      <template #header>
        <div class="flex-cb">
          <div class="title-group">
            <h4 class="m-0 font-semibold text-lg">{{ t('feedbackManage.title') }}</h4>
            <p class="text-gray-400 text-xs mt-1">{{ t('feedbackManage.subtitle') }}</p>
          </div>
          <div class="flex gap-2">
            <ElTag type="success">
              {{ t('tokenManage.dataCount', { count: pagination.total }) }}
            </ElTag>
          </div>
        </div>
      </template>

      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadFeedback" />

      <!-- Table of Feedback -->
      <ArtTable
        v-loading="loading"
        :data="feedbackList"
        :columns="columns"
        :pagination="pagination"
        :height="computedTableHeight"
        empty-height="360px"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
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

        <!-- Custom slot: status -->
        <template #status="{ row }">
          <ElTag
            :type="row.status === 1 ? 'success' : 'warning'"
            effect="light"
            size="small"
            class="rounded-full px-3 font-medium"
          >
            <i
              class="fas fa-circle text-[6px] mr-1"
              :class="row.status === 1 ? 'text-emerald-500' : 'text-amber-500'"
            ></i>
            {{ row.status === 1 ? t('feedbackManage.processed') : t('feedbackManage.pending') }}
          </ElTag>
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
          <div class="flex items-center justify-center gap-1">
            <ElTooltip :content="t('feedbackManage.detail')" placement="top">
              <ArtButtonTable type="view" @click="openDetail(row)" />
            </ElTooltip>

            <ElTooltip
              :content="
                row.status === 1
                  ? t('feedbackManage.markPending')
                  : t('feedbackManage.markProcessed')
              "
              placement="top"
            >
              <ArtButtonTable
                :icon="row.status === 1 ? 'ri:refresh-line' : 'ri:checkbox-circle-line'"
                :icon-class="
                  row.status === 1 ? 'bg-warning/12 text-warning' : 'bg-success/12 text-success'
                "
                @click="toggleStatus(row)"
              />
            </ElTooltip>

            <ElTooltip :content="t('feedbackManage.delete')" placement="top">
              <ArtButtonTable type="delete" @click="deleteItem(row)" />
            </ElTooltip>
          </div>
        </template>
      </ArtTable>
    </ElCard>

    <!-- Dialog: Feedback Detail -->
    <ElDialog v-model="detailDialogVisible" :title="t('feedbackManage.detailTitle')" width="600px">
      <div v-if="currentDetail" class="flex flex-col gap-4">
        <div
          class="grid grid-cols-2 gap-3 p-3 bg-gray-50 dark:bg-zinc-800/60 rounded border border-gray-200 dark:border-zinc-700 text-xs"
        >
          <div
            ><span class="text-gray-400">{{ t('feedbackManage.appName') }}:</span>
            <span class="font-medium">{{ currentDetail.app_name }}</span> (v{{
              currentDetail.version
            }})</div
          >
          <div
            ><span class="text-gray-400">{{ t('feedbackManage.platform') }}:</span>
            <span class="capitalize font-medium">{{ currentDetail.platform }}</span></div
          >
          <div
            ><span class="text-gray-400">{{ t('feedbackManage.userUuid') }}:</span>
            <span class="font-mono">{{ currentDetail.user_uuid || '-' }}</span></div
          >
          <div
            ><span class="text-gray-400">{{ t('feedbackManage.contact') }}:</span>
            <span class="font-medium">{{ currentDetail.contact || '-' }}</span></div
          >
          <div
            ><span class="text-gray-400">{{ t('feedbackManage.ip') }}:</span>
            <span class="font-mono">{{ currentDetail.ip }}</span> ({{
              currentDetail.ip_location || '未知'
            }})</div
          >
          <div
            ><span class="text-gray-400">{{ t('feedbackManage.createdAt') }}:</span>
            <span>{{ formatTime(currentDetail.created_at) }}</span></div
          >
        </div>

        <div class="flex flex-col gap-1">
          <span class="text-xs font-semibold text-gray-500"
            >{{ t('feedbackManage.content') }}:</span
          >
          <div
            class="p-3 bg-white dark:bg-zinc-900 border border-gray-200 dark:border-zinc-800 rounded text-sm whitespace-pre-wrap break-words leading-relaxed"
          >
            {{ currentDetail.content }}
          </div>
        </div>
      </div>
      <template #footer>
        <div class="flex justify-between items-center">
          <div>
            <ElButton
              v-if="currentDetail"
              :type="currentDetail.status === 1 ? 'warning' : 'success'"
              size="small"
              @click="toggleStatus(currentDetail)"
            >
              {{
                currentDetail.status === 1
                  ? t('feedbackManage.markPending')
                  : t('feedbackManage.markProcessed')
              }}
            </ElButton>
          </div>
          <ElButton type="primary" @click="detailDialogVisible = false">{{
            t('common.confirm')
          }}</ElButton>
        </div>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted, computed } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { fetchGetFeedback, fetchUpdateFeedbackStatus, fetchDeleteFeedback } from '@/api/token'

  defineOptions({ name: 'Feedback' })

  const { t } = useI18n()

  const computedTableHeight = computed(() => {
    return ''
  })

  const loading = ref(false)
  const feedbackList = ref<any[]>([])
  const detailDialogVisible = ref(false)
  const currentDetail = ref<any>(null)

  // Pagination state
  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const { columns, columnChecks } = useTableColumns(() => [
    { type: 'globalIndex', label: t('feedbackManage.index'), width: 70, align: 'center' },
    { prop: 'app_name', label: t('feedbackManage.appName'), minWidth: 120 },
    {
      prop: 'version',
      label: t('feedbackManage.version'),
      width: 90,
      useSlot: true,
      slotName: 'version'
    },
    {
      prop: 'platform',
      label: t('feedbackManage.platform'),
      width: 90,
      align: 'center',
      useSlot: true,
      slotName: 'platform'
    },
    {
      prop: 'user_uuid',
      label: t('feedbackManage.userUuid'),
      minWidth: 140,
      showOverflowTooltip: true,
      useSlot: true,
      slotName: 'user_uuid'
    },
    {
      prop: 'content',
      label: t('feedbackManage.content'),
      minWidth: 220,
      showOverflowTooltip: true
    },
    {
      prop: 'contact',
      label: t('feedbackManage.contact'),
      minWidth: 130,
      showOverflowTooltip: true
    },
    {
      prop: 'ip',
      label: t('feedbackManage.ip'),
      width: 120,
      useSlot: true,
      slotName: 'ip'
    },
    {
      prop: 'ip_location',
      label: t('feedbackManage.ipLocation'),
      minWidth: 130,
      showOverflowTooltip: true
    },
    {
      prop: 'status',
      label: t('feedbackManage.status'),
      width: 100,
      align: 'center',
      useSlot: true,
      slotName: 'status'
    },
    {
      prop: 'created_at',
      label: t('feedbackManage.createdAt'),
      width: 160,
      useSlot: true,
      slotName: 'createdAt'
    },
    {
      prop: 'operation',
      label: t('feedbackManage.operations'),
      width: 150,
      fixed: 'right',
      useSlot: true,
      slotName: 'operation'
    }
  ])

  onMounted(() => {
    loadFeedback()
  })

  const loadFeedback = async () => {
    loading.value = true
    try {
      const res = await fetchGetFeedback({
        current: pagination.current,
        size: pagination.size
      })
      feedbackList.value = res.list || []
      pagination.total = res.total || 0
    } catch (e: any) {
      ElMessage.error(e.message || '加载反馈失败')
    } finally {
      loading.value = false
    }
  }

  const handleSizeChange = (val: number) => {
    pagination.size = val
    loadFeedback()
  }

  const handleCurrentChange = (val: number) => {
    pagination.current = val
    loadFeedback()
  }

  const openDetail = (row: any) => {
    currentDetail.value = row
    detailDialogVisible.value = true
  }

  const toggleStatus = async (row: any) => {
    const targetStatus = row.status === 1 ? 0 : 1
    try {
      await fetchUpdateFeedbackStatus({ id: row.id, status: targetStatus })
      row.status = targetStatus
      ElMessage.success(t('feedbackManage.successUpdateStatus'))
    } catch (e: any) {
      ElMessage.error(e.message || t('feedbackManage.errorUpdateStatus'))
    }
  }

  const deleteItem = (row: any) => {
    ElMessageBox.confirm(t('feedbackManage.deleteConfirm'), t('tokenManage.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      try {
        await fetchDeleteFeedback({ id: row.id })
        ElMessage.success(t('feedbackManage.successDelete'))
        loadFeedback()
      } catch (e: any) {
        ElMessage.error(e.message || t('feedbackManage.errorDelete'))
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
