<template>
  <div class="blacklist-page art-full-height">
    <ElCard class="art-table-card">
      <template #header>
        <div class="flex-cb">
          <div class="title-group">
            <h4 class="m-0 font-semibold text-lg">{{ t('blacklistManage.title') }}</h4>
            <p class="text-gray-400 text-xs mt-1">{{ t('blacklistManage.subtitle') }}</p>
          </div>
          <div class="flex gap-2">
            <ElTag type="success">
              {{ t('blacklistManage.dataCount', { count: blacklist.length }) }}
            </ElTag>
          </div>
        </div>
      </template>

      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadBlacklist">
        <template #left>
          <ElSpace wrap>
            <ElButton type="primary" @click="openAddDialog" v-ripple>
              <i class="fas fa-plus mr-1"></i> {{ t('blacklistManage.addBtn') }}
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- Table of Blacklisted entries -->
      <ArtTable
        v-loading="loading"
        :data="blacklist"
        :columns="columns"
        :show-pagination="false"
        :height="computedTableHeight"
        empty-height="360px"
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
          <ElTooltip :content="t('blacklistManage.remove')" placement="top">
            <ArtButtonTable type="delete" @click="deleteBlacklist(row)" />
          </ElTooltip>
        </template>
      </ArtTable>
    </ElCard>

    <!-- Dialog: Add Blacklist Entry -->
    <ElDialog v-model="addDialogVisible" :title="t('blacklistManage.addBtn')" width="500px">
      <ElForm :model="addForm" label-width="120px" :rules="formRules" ref="addFormRef">
        <ElFormItem :label="t('blacklistManage.token')" prop="token">
          <ElInput v-model="addForm.token" :placeholder="t('blacklistManage.placeholderToken')" />
        </ElFormItem>
        <ElFormItem :label="t('blacklistManage.platform')" prop="platform">
          <ElSelect
            v-model="addForm.platform"
            :placeholder="t('blacklistManage.placeholderPlatform')"
            style="width: 100%"
          >
            <ElOption label="android" value="android" />
            <ElOption label="iOS" value="iOS" />
            <ElOption label="windows" value="windows" />
            <ElOption label="Linux" value="Linux" />
            <ElOption label="mac" value="mac" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('blacklistManage.version')" prop="version">
          <ElInput
            v-model="addForm.version"
            :placeholder="t('blacklistManage.placeholderVersion')"
          />
        </ElFormItem>
        <ElFormItem :label="t('blacklistManage.userUuid')" prop="user_uuid">
          <ElInput
            v-model="addForm.user_uuid"
            :placeholder="t('blacklistManage.placeholderUserUuid')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <span class="dialog-footer">
          <ElButton @click="addDialogVisible = false">{{ t('common.cancel') }}</ElButton>
          <ElButton type="primary" @click="submitAdd">{{ t('common.confirm') }}</ElButton>
        </span>
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
  import { useWindowSize } from '@vueuse/core'
  import { fetchGetBlacklist, fetchAddBlacklist, fetchDeleteBlacklist } from '@/api/token'

  defineOptions({ name: 'Blacklist' })

  const { t } = useI18n()
  const { height: windowHeight } = useWindowSize()

  const computedTableHeight = computed(() => {
    const val = windowHeight.value - 360
    return val > 300 ? val : 300
  })

  const loading = ref(false)
  const blacklist = ref<any[]>([])
  const addDialogVisible = ref(false)

  const addForm = reactive({
    token: '',
    platform: 'android',
    version: '',
    user_uuid: ''
  })

  // Reactive translation rules
  const formRules = computed(() => ({
    token: [{ required: true, message: t('blacklistManage.ruleToken'), trigger: 'blur' }],
    platform: [{ required: true, message: t('blacklistManage.rulePlatform'), trigger: 'change' }],
    version: [{ required: true, message: t('blacklistManage.ruleVersion'), trigger: 'blur' }],
    user_uuid: [{ required: true, message: t('blacklistManage.ruleUserUuid'), trigger: 'blur' }]
  }))

  const addFormRef = ref()

  // Use the useTableColumns hook to manage visible/hidden columns, table checks and icons
  const { columns, columnChecks } = useTableColumns(() => [
    { type: 'globalIndex', label: t('blacklistManage.index'), width: 80, align: 'center' },
    { prop: 'token', label: t('blacklistManage.token'), minWidth: 200, showOverflowTooltip: true },
    { prop: 'platform', label: t('blacklistManage.platform'), width: 120 },
    {
      prop: 'version',
      label: t('blacklistManage.version'),
      width: 120,
      useSlot: true,
      slotName: 'version'
    },
    { prop: 'user_uuid', label: t('blacklistManage.userUuid'), minWidth: 180 },
    {
      prop: 'created_at',
      label: t('blacklistManage.interceptTime'),
      width: 180,
      useSlot: true,
      slotName: 'createdAt'
    },
    {
      prop: 'operation',
      label: t('blacklistManage.operations'),
      width: 100,
      fixed: 'right',
      useSlot: true,
      slotName: 'operation'
    }
  ])

  onMounted(() => {
    loadBlacklist()
  })

  const loadBlacklist = async () => {
    loading.value = true
    try {
      const res = await fetchGetBlacklist()
      blacklist.value = res || []
    } catch (e: any) {
      ElMessage.error(e.message || t('blacklistManage.errorLoad'))
    } finally {
      loading.value = false
    }
  }

  const openAddDialog = () => {
    addForm.token = ''
    addForm.platform = 'android'
    addForm.version = ''
    addForm.user_uuid = ''
    addDialogVisible.value = true
  }

  const submitAdd = async () => {
    addFormRef.value?.validate(async (valid: boolean) => {
      if (!valid) return
      try {
        await fetchAddBlacklist(addForm)
        ElMessage.success(t('blacklistManage.successAdd'))
        addDialogVisible.value = false
        loadBlacklist()
      } catch (e: any) {
        ElMessage.error(e.message || t('blacklistManage.errorAdd'))
      }
    })
  }

  const deleteBlacklist = (row: any) => {
    ElMessageBox.confirm(t('blacklistManage.removeConfirm'), t('blacklistManage.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      try {
        await fetchDeleteBlacklist({ id: row.id })
        ElMessage.success(t('blacklistManage.successRemove'))
        loadBlacklist()
      } catch (e: any) {
        ElMessage.error(e.message || t('blacklistManage.errorRemove'))
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
