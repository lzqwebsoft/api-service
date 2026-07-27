<template>
  <div class="page-content">
    <ElCard shadow="never" class="mb-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h2 class="text-lg font-bold text-gray-800 dark:text-gray-100 flex items-center gap-2">
            <ArtSvgIcon icon="ri:file-download-line" class="text-primary text-xl" />
            {{ t('fileManage.title') }}
          </h2>
          <p class="text-xs text-gray-400 mt-1">
            {{ t('fileManage.subtitle') }}
          </p>
        </div>
        <div class="flex items-center gap-3">
          <!-- Filter App -->
          <ElSelect
            v-model="filterAppRecordID"
            :placeholder="t('fileManage.filterApp')"
            clearable
            style="width: 180px"
            @change="handleSearch"
          >
            <ElOption
              v-for="app in appList"
              :key="app.id"
              :label="`${app.name} (${app.app_id})`"
              :value="app.id"
            />
          </ElSelect>

          <!-- Create File Button -->
          <ElButton type="primary" class="font-medium" @click="openCreateDialog">
            <ArtSvgIcon icon="ri:add-line" class="mr-1 text-sm" />
            {{ t('fileManage.createBtn') }}
          </ElButton>
        </div>
      </div>
    </ElCard>

    <!-- Table Card -->
    <ElCard shadow="never">
      <ArtTable
        v-loading="loading"
        :data="files"
        :columns="columns"
        :pagination="{
          current: pagination.current,
          size: pagination.size,
          total: pagination.total
        }"
        @page-change="handlePageChange"
        @size-change="handleSizeChange"
      >
        <!-- App Name -->
        <template #appName="{ row }">
          <div class="flex flex-col">
            <span class="font-medium text-gray-800 dark:text-gray-200 text-xs">{{
              row.app_name
            }}</span>
            <span class="text-[10px] text-gray-400 font-mono">{{ row.app_id }}</span>
          </div>
        </template>

        <!-- Version -->
        <template #version="{ row }">
          <ElTag size="small" type="primary" effect="light" class="font-mono">
            v{{ row.version }}
          </ElTag>
        </template>

        <!-- Download URL -->
        <template #downloadUrl="{ row }">
          <div class="flex items-center gap-1 group truncate max-w-[260px]">
            <a
              :href="getRealDownloadUrl(row)"
              target="_blank"
              class="font-mono text-xs text-primary hover:underline truncate"
              :title="getRealDownloadUrl(row)"
            >
              {{ getRealDownloadUrl(row) }}
            </a>
            <ElButton
              circle
              size="small"
              type="primary"
              link
              class="opacity-60 group-hover:opacity-100 transition-opacity flex-shrink-0"
              @click="copyText(getRealDownloadUrl(row))"
            >
              <ArtSvgIcon icon="ri:file-copy-line" class="text-xs" />
            </ElButton>
          </div>
        </template>

        <!-- Total Downloads -->
        <template #downloadCount="{ row }">
          <ElTag size="small" type="success" effect="plain" class="font-mono font-bold">
            <ArtSvgIcon icon="ri:download-2-line" class="mr-1 text-xs" />
            {{ row.download_count }}
          </ElTag>
        </template>

        <!-- Status -->
        <template #status="{ row }">
          <ElSwitch v-model="row.is_active" size="small" @change="handleToggleStatus(row)" />
        </template>

        <!-- Created At -->
        <template #createdAt="{ row }">
          <span class="text-xs text-gray-500 font-mono">
            {{ formatTime(row.created_at) }}
          </span>
        </template>

        <!-- Operations -->
        <template #operations="{ row }">
          <div class="flex items-center justify-center gap-1.5">
            <ElTooltip :content="t('fileManage.copyLink')" placement="top">
              <ArtButtonTable
                icon="ri:code-s-slash-line"
                iconClass="bg-primary/12 text-primary"
                class="!mr-0"
                @click="openLinkGeneratorDialog(row)"
              />
            </ElTooltip>
            <ElTooltip :content="t('fileManage.viewLogs')" placement="top">
              <ArtButtonTable
                icon="ri:bar-chart-2-line"
                iconClass="bg-info/12 text-info"
                class="!mr-0"
                @click="openLogsDrawer(row)"
              />
            </ElTooltip>
            <ElTooltip :content="t('fileManage.edit')" placement="top">
              <ArtButtonTable type="edit" class="!mr-0" @click="openEditDialog(row)" />
            </ElTooltip>
            <ElTooltip :content="t('fileManage.delete')" placement="top">
              <ArtButtonTable type="delete" class="!mr-0" @click="deleteItem(row)" />
            </ElTooltip>
          </div>
        </template>
      </ArtTable>
    </ElCard>

    <!-- Dialog: Create / Edit File -->
    <ElDialog
      v-model="fileDialogVisible"
      :title="isEdit ? t('fileManage.editTitle') : t('fileManage.createTitle')"
      width="560px"
    >
      <ElForm :model="fileForm" label-width="120px" class="pr-4">
        <ElFormItem :label="t('fileManage.appName')" required>
          <ElSelect
            v-model="fileForm.app_record_id"
            :placeholder="t('fileManage.filterApp')"
            style="width: 100%"
          >
            <ElOption
              v-for="app in appList"
              :key="app.id"
              :label="`${app.name} (${app.app_id})`"
              :value="app.id"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem :label="t('fileManage.version')" required>
          <ElInput v-model="fileForm.version" :placeholder="t('fileManage.placeholderVersion')" />
        </ElFormItem>

        <ElFormItem :label="t('fileManage.fileName')" required>
          <ElInput
            v-model="fileForm.file_name"
            :placeholder="t('fileManage.placeholderFileName')"
          />
        </ElFormItem>

        <ElFormItem :label="t('fileManage.storageType')" required>
          <ElRadioGroup v-model="storageType" class="flex flex-col sm:flex-row gap-2">
            <ElRadio value="external">
              <span class="font-medium text-xs">{{ t('fileManage.storageExternal') }}</span>
            </ElRadio>
            <ElRadio value="local">
              <span class="font-medium text-xs">{{ t('fileManage.storageLocal') }}</span>
            </ElRadio>
          </ElRadioGroup>
        </ElFormItem>

        <ElFormItem v-if="storageType === 'external'" :label="t('fileManage.downloadUrl')" required>
          <ElInput
            v-model="fileForm.download_url"
            type="textarea"
            :rows="3"
            :placeholder="t('fileManage.placeholderDownloadUrl')"
          />
        </ElFormItem>

        <ElFormItem v-else :label="t('fileManage.uploadLocalFile')" required>
          <div class="flex flex-col gap-2 w-full">
            <ElUpload
              action="#"
              :auto-upload="false"
              :show-file-list="false"
              :on-change="handleCustomFileUpload"
            >
              <ElButton type="primary" :loading="uploading">
                <ArtSvgIcon icon="ri:upload-cloud-2-line" class="mr-1 text-sm" />
                {{ t('fileManage.selectAndUpload') }}
              </ElButton>
            </ElUpload>
            <div
              v-if="fileForm.download_url"
              class="p-2 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800 rounded text-xs text-emerald-700 dark:text-emerald-300 font-mono break-all flex items-center justify-between"
            >
              <span>{{ t('fileManage.uploadedToServer') }}: {{ fileForm.download_url }}</span>
              <ElTag v-if="fileForm.file_size" size="small" type="success">
                {{ formatFileSize(fileForm.file_size) }}
              </ElTag>
            </div>
          </div>
        </ElFormItem>

        <ElFormItem :label="t('fileManage.description')">
          <ElInput
            v-model="fileForm.description"
            type="textarea"
            :rows="3"
            :placeholder="t('fileManage.placeholderDescription')"
          />
        </ElFormItem>

        <ElFormItem :label="t('fileManage.status')">
          <ElSwitch v-model="fileForm.is_active" />
        </ElFormItem>
      </ElForm>

      <template #footer>
        <div class="dialog-footer">
          <ElButton @click="fileDialogVisible = false">{{ t('common.cancel') }}</ElButton>
          <ElButton type="primary" :loading="submitLoading" @click="submitFileForm">
            {{ t('common.confirm') }}
          </ElButton>
        </div>
      </template>
    </ElDialog>

    <!-- Dialog: Link & HTML Code Generator -->
    <ElDialog
      v-model="linkGeneratorVisible"
      :title="t('fileManage.linkGeneratorTitle')"
      width="640px"
    >
      <div v-if="currentFile" class="flex flex-col gap-4">
        <!-- Exact ID API Download Link -->
        <div
          class="p-3 bg-gray-50 dark:bg-zinc-800/60 rounded border border-gray-200 dark:border-zinc-700 flex flex-col gap-2"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">
              {{ t('fileManage.exactDownloadUrl') }}：
            </span>
            <ElButton
              size="small"
              type="primary"
              link
              @click="copyText(getApiDownloadUrl(currentFile.id))"
            >
              <ArtSvgIcon icon="ri:file-copy-line" class="mr-1 text-xs" />
              {{ t('common.copy') }}
            </ElButton>
          </div>
          <div
            class="font-mono text-xs text-primary bg-white dark:bg-zinc-900 p-2 rounded border border-gray-200 dark:border-zinc-800 break-all select-all"
          >
            {{ getApiDownloadUrl(currentFile.id) }}
          </div>
        </div>

        <!-- App Latest Download Link -->
        <div
          class="p-3 bg-gray-50 dark:bg-zinc-800/60 rounded border border-gray-200 dark:border-zinc-700 flex flex-col gap-2"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">
              {{ t('fileManage.latestDownloadUrl') }}：
            </span>
            <ElButton
              size="small"
              type="primary"
              link
              @click="copyText(getLatestDownloadUrl(currentFile.app_id))"
            >
              <ArtSvgIcon icon="ri:file-copy-line" class="mr-1 text-xs" />
              {{ t('common.copy') }}
            </ElButton>
          </div>
          <div
            class="font-mono text-xs text-emerald-600 dark:text-emerald-400 bg-white dark:bg-zinc-900 p-2 rounded border border-gray-200 dark:border-zinc-800 break-all select-all"
          >
            {{ getLatestDownloadUrl(currentFile.app_id) }}
          </div>
        </div>

        <!-- HTML Snippet -->
        <div
          class="p-3 bg-gray-50 dark:bg-zinc-800/60 rounded border border-gray-200 dark:border-zinc-700 flex flex-col gap-2"
        >
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold text-gray-700 dark:text-gray-300">
              {{ t('fileManage.htmlEmbedCode') }}：
            </span>
            <ElButton
              size="small"
              type="primary"
              link
              @click="copyText(getHtmlSnippet(currentFile))"
            >
              <ArtSvgIcon icon="ri:file-copy-line" class="mr-1 text-xs" />
              {{ t('common.copy') }}
            </ElButton>
          </div>
          <div
            class="font-mono text-xs text-amber-600 dark:text-amber-400 bg-white dark:bg-zinc-900 p-2 rounded border border-gray-200 dark:border-zinc-800 break-all select-all whitespace-pre-wrap"
          >
            {{ getHtmlSnippet(currentFile) }}
          </div>
        </div>
      </div>
    </ElDialog>

    <!-- Drawer: Download Logs -->
    <ElDrawer
      v-model="logsDrawerVisible"
      :title="`${t('fileManage.logsTitle')}: ${currentFile?.file_name || ''}`"
      size="65%"
    >
      <div v-loading="logsLoading" class="p-4 flex flex-col gap-4">
        <ElTable :data="logs" stripe border style="width: 100%" class="rounded-lg overflow-hidden">
          <ElTableColumn prop="ip" :label="t('fileManage.ip')" width="140" align="center">
            <template #default="scope">
              <span class="font-mono text-xs">{{ scope.row.ip }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn
            prop="ip_location"
            :label="t('fileManage.ipLocation')"
            width="120"
            align="center"
          >
            <template #default="scope">
              <span class="text-xs text-gray-500">{{ scope.row.ip_location || '-' }}</span>
            </template>
          </ElTableColumn>

          <ElTableColumn prop="channel" :label="t('fileManage.channel')" width="100" align="center">
            <template #default="scope">
              <ElTag v-if="scope.row.channel" size="small" type="info">{{
                scope.row.channel
              }}</ElTag>
              <span v-else class="text-xs text-gray-400">-</span>
            </template>
          </ElTableColumn>

          <ElTableColumn prop="user_agent" :label="t('fileManage.userAgent')" min-width="200">
            <template #default="scope">
              <span
                class="text-xs text-gray-600 dark:text-gray-300 font-mono truncate block"
                :title="scope.row.user_agent"
              >
                {{ scope.row.user_agent || '-' }}
              </span>
            </template>
          </ElTableColumn>

          <ElTableColumn prop="referer" :label="t('fileManage.referer')" min-width="180">
            <template #default="scope">
              <span
                class="text-xs text-gray-600 dark:text-gray-300 font-mono truncate block"
                :title="scope.row.referer"
              >
                {{ scope.row.referer || '-' }}
              </span>
            </template>
          </ElTableColumn>

          <ElTableColumn
            prop="created_at"
            :label="t('fileManage.downloadTime')"
            width="160"
            align="center"
          >
            <template #default="scope">
              <span class="text-xs text-gray-500 font-mono">{{
                formatTime(scope.row.created_at)
              }}</span>
            </template>
          </ElTableColumn>
        </ElTable>

        <div class="flex justify-end mt-2">
          <ElPagination
            background
            layout="prev, pager, next, total"
            :current-page="logsPagination.current"
            :page-size="logsPagination.size"
            :total="logsPagination.total"
            @current-change="handleLogsPageChange"
          />
        </div>
      </div>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useClipboard } from '@vueuse/core'
  import { fetchGetApps } from '@/api/token'
  import {
    AppFileItem,
    FileDownloadLogItem,
    fetchGetFiles,
    fetchCreateFile,
    fetchUpdateFile,
    fetchDeleteFile,
    fetchGetDownloadLogs,
    fetchUploadFile
  } from '@/api/file'

  defineOptions({ name: 'Files' })

  const { t } = useI18n()
  const { copy } = useClipboard()

  const loading = ref(false)
  const submitLoading = ref(false)
  const logsLoading = ref(false)
  const uploading = ref(false)

  const storageType = ref<'external' | 'local'>('external')

  const files = ref<AppFileItem[]>([])
  const appList = ref<any[]>([])
  const filterAppRecordID = ref<number | undefined>(undefined)

  const pagination = reactive({
    current: 1,
    size: 15,
    total: 0
  })

  const fileDialogVisible = ref(false)
  const isEdit = ref(false)

  const fileForm = reactive<Partial<AppFileItem>>({
    id: 0,
    app_record_id: undefined,
    version: '',
    file_name: '',
    download_url: '',
    description: '',
    is_active: true
  })

  const linkGeneratorVisible = ref(false)
  const currentFile = ref<AppFileItem | null>(null)

  const logsDrawerVisible = ref(false)
  const logs = ref<FileDownloadLogItem[]>([])
  const logsPagination = reactive({
    current: 1,
    size: 15,
    total: 0
  })

  const { columns } = useTableColumns(() => [
    { type: 'globalIndex', label: t('fileManage.index'), width: 65, align: 'center' },
    { prop: 'appName', label: t('fileManage.appName'), width: 140, useSlot: true },
    { prop: 'version', label: t('fileManage.version'), width: 90, align: 'center', useSlot: true },
    { prop: 'file_name', label: t('fileManage.fileName'), minWidth: 160 },
    { prop: 'downloadUrl', label: t('fileManage.downloadUrl'), minWidth: 200, useSlot: true },
    {
      prop: 'downloadCount',
      label: t('fileManage.downloadCount'),
      width: 110,
      align: 'center',
      useSlot: true
    },
    { prop: 'status', label: t('fileManage.status'), width: 90, align: 'center', useSlot: true },
    {
      prop: 'createdAt',
      label: t('fileManage.createdAt'),
      width: 150,
      align: 'center',
      useSlot: true
    },
    {
      prop: 'operations',
      label: t('fileManage.operations'),
      width: 175,
      fixed: 'right',
      align: 'center',
      useSlot: true
    }
  ])

  onMounted(() => {
    loadApps()
    loadFiles()
  })

  const loadApps = async () => {
    try {
      const res = await fetchGetApps()
      appList.value = res || []
    } catch (e: any) {
      ElMessage.error(e.message || '加载应用列表失败')
    }
  }

  const loadFiles = async () => {
    loading.value = true
    try {
      const res = await fetchGetFiles({
        app_record_id: filterAppRecordID.value,
        current: pagination.current,
        size: pagination.size
      })
      files.value = res.list || []
      pagination.total = res.total || 0
    } catch (e: any) {
      ElMessage.error(e.message || '加载文件列表失败')
    } finally {
      loading.value = false
    }
  }

  const handleSearch = () => {
    pagination.current = 1
    loadFiles()
  }

  const handlePageChange = (page: number) => {
    pagination.current = page
    loadFiles()
  }

  const handleSizeChange = (size: number) => {
    pagination.size = size
    pagination.current = 1
    loadFiles()
  }

  const openCreateDialog = () => {
    isEdit.value = false
    storageType.value = 'external'
    fileForm.id = 0
    fileForm.app_record_id = appList.value[0]?.id
    fileForm.version = ''
    fileForm.file_name = ''
    fileForm.download_url = ''
    fileForm.file_size = 0
    fileForm.description = ''
    fileForm.is_active = true
    fileDialogVisible.value = true
  }

  const openEditDialog = (row: AppFileItem) => {
    isEdit.value = true
    storageType.value =
      row.download_url.startsWith('local://') || row.download_url.startsWith('runtimes/uploads/')
        ? 'local'
        : 'external'
    fileForm.id = row.id
    fileForm.app_record_id = row.app_record_id
    fileForm.version = row.version
    fileForm.file_name = row.file_name
    fileForm.download_url = row.download_url
    fileForm.file_size = row.file_size
    fileForm.description = row.description
    fileForm.is_active = row.is_active
    fileDialogVisible.value = true
  }

  const submitFileForm = async () => {
    if (!fileForm.app_record_id) {
      return ElMessage.warning(t('fileManage.ruleAppName'))
    }
    if (!fileForm.version) {
      return ElMessage.warning(t('fileManage.ruleVersion'))
    }
    if (!fileForm.file_name) {
      return ElMessage.warning(t('fileManage.ruleFileName'))
    }
    if (!fileForm.download_url) {
      return ElMessage.warning(t('fileManage.ruleDownloadUrl'))
    }

    submitLoading.value = true
    try {
      if (isEdit.value) {
        await fetchUpdateFile(fileForm)
        ElMessage.success(t('fileManage.successUpdate'))
      } else {
        await fetchCreateFile(fileForm)
        ElMessage.success(t('fileManage.successCreate'))
      }
      fileDialogVisible.value = false
      loadFiles()
    } catch (e: any) {
      ElMessage.error(e.message || '操作失败')
    } finally {
      submitLoading.value = false
    }
  }

  const handleToggleStatus = async (row: AppFileItem) => {
    try {
      await fetchUpdateFile({
        id: row.id,
        app_record_id: row.app_record_id,
        version: row.version,
        file_name: row.file_name,
        download_url: row.download_url,
        description: row.description,
        is_active: row.is_active
      })
      ElMessage.success(t('fileManage.statusUpdated'))
    } catch (e: any) {
      row.is_active = !row.is_active
      ElMessage.error(e.message || t('fileManage.statusUpdateFailed'))
    }
  }

  const deleteItem = (row: AppFileItem) => {
    ElMessageBox.confirm(t('fileManage.deleteConfirm', { name: row.file_name }), t('common.tips'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      try {
        await fetchDeleteFile({ id: row.id })
        ElMessage.success(t('fileManage.successDelete'))
        loadFiles()
      } catch (e: any) {
        ElMessage.error(e.message || '删除文件失败')
      }
    })
  }

  const openLinkGeneratorDialog = (row: AppFileItem) => {
    currentFile.value = row
    linkGeneratorVisible.value = true
  }

  const getApiDownloadUrl = (fileId: number) => {
    return `${window.location.origin}/api/v1/download?file_id=${fileId}`
  }

  const getRealDownloadUrl = (row?: AppFileItem) => {
    if (!row || !row.download_url) return ''
    if (
      row.download_url.startsWith('local://') ||
      row.download_url.startsWith('runtimes/uploads/')
    ) {
      return getApiDownloadUrl(row.id)
    }
    return row.download_url
  }

  const getLatestDownloadUrl = (appId?: string) => {
    return `${window.location.origin}/api/v1/download?app_id=${appId || ''}`
  }

  const getHtmlSnippet = (file: AppFileItem) => {
    const url = getApiDownloadUrl(file.id)
    return `<a href="${url}" target="_blank">${t('fileManage.downloadFileName', { name: file.file_name })}</a>`
  }

  const openLogsDrawer = (row: AppFileItem) => {
    currentFile.value = row
    logsPagination.current = 1
    logsDrawerVisible.value = true
    loadLogs(row.id)
  }

  const loadLogs = async (fileId: number) => {
    logsLoading.value = true
    try {
      const res = await fetchGetDownloadLogs({
        file_id: fileId,
        current: logsPagination.current,
        size: logsPagination.size
      })
      logs.value = res.list || []
      logsPagination.total = res.total || 0
    } catch (e: any) {
      ElMessage.error(e.message || '加载下载日志失败')
    } finally {
      logsLoading.value = false
    }
  }

  const handleLogsPageChange = (page: number) => {
    logsPagination.current = page
    if (currentFile.value) {
      loadLogs(currentFile.value.id)
    }
  }

  const formatTime = (timeStr: string) => {
    if (!timeStr) return '-'
    const d = new Date(timeStr)
    return d.toLocaleString()
  }

  const copyText = async (str: string) => {
    try {
      await copy(str)
      ElMessage.success(t('common.copySuccess'))
    } catch {
      ElMessage.error(t('common.copyFailed'))
    }
  }

  const handleCustomFileUpload = async (uploadFile: any) => {
    const rawFile = uploadFile.raw
    if (!rawFile) return

    const formData = new FormData()
    formData.append('file', rawFile)

    uploading.value = true
    try {
      const res = await fetchUploadFile(formData)
      fileForm.download_url = res.download_url
      if (!fileForm.file_name) {
        fileForm.file_name = res.file_name
      }
      fileForm.file_size = res.file_size
      ElMessage.success(t('fileManage.uploadSuccess'))
    } catch (e: any) {
      ElMessage.error(e.message || t('fileManage.uploadFailed'))
    } finally {
      uploading.value = false
    }
  }

  const formatFileSize = (bytes?: number) => {
    if (!bytes || bytes <= 0) return '0 B'
    const units = ['B', 'KB', 'MB', 'GB']
    let size = bytes
    let unitIndex = 0
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024
      unitIndex++
    }
    return `${size.toFixed(2)} ${units[unitIndex]}`
  }
</script>

<style scoped></style>
