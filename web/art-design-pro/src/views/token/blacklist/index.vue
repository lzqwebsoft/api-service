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
    <ElDialog v-model="addDialogVisible" :title="t('blacklistManage.addBtn')" width="520px">
      <ElForm :model="addForm" label-width="120px" :rules="formRules" ref="addFormRef">
        <!-- Step 1: 应用名称 -->
        <ElFormItem :label="t('blacklistManage.appName')" prop="appName">
          <ElSelect
            v-model="addForm.appName"
            placeholder="请选择应用名称"
            style="width: 100%"
            @change="handleAppNameChange"
          >
            <ElOption v-for="name in appNameOptions" :key="name" :label="name" :value="name" />
          </ElSelect>
        </ElFormItem>

        <!-- Step 2: 平台 -->
        <ElFormItem :label="t('blacklistManage.platform')" prop="platform">
          <ElSelect
            v-model="addForm.platform"
            placeholder="请选择平台"
            style="width: 100%"
            :disabled="!addForm.appName"
            @change="handlePlatformChange"
          >
            <ElOption v-for="p in platformOptions" :key="p" :label="p" :value="p" />
          </ElSelect>
        </ElFormItem>

        <!-- Step 3: 版本 -->
        <ElFormItem :label="t('blacklistManage.version')" prop="version">
          <ElSelect
            v-model="addForm.version"
            placeholder="请选择版本"
            style="width: 100%"
            :disabled="!addForm.platform"
            @change="handleVersionChange"
          >
            <ElOption v-for="v in versionOptions" :key="v" :label="v" :value="v" />
          </ElSelect>
        </ElFormItem>

        <!-- Step 4: Token -->
        <ElFormItem :label="t('blacklistManage.token')" prop="token_id">
          <ElSelect
            v-model="addForm.token_id"
            placeholder="请选择 Access Token"
            style="width: 100%"
            :disabled="!addForm.version"
            @change="handleTokenChange"
          >
            <ElOption
              v-for="item in tokenOptions"
              :key="item.id"
              :label="item.token"
              :value="item.id"
            />
          </ElSelect>
        </ElFormItem>

        <!-- Step 5: 用户 UUID -->
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
  import {
    fetchGetBlacklist,
    fetchAddBlacklist,
    fetchDeleteBlacklist,
    fetchGetTokens
  } from '@/api/token'

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
  const allTokensList = ref<any[]>([])

  const addForm = reactive({
    appName: '',
    platform: '',
    version: '',
    token_id: undefined as number | undefined,
    token: '',
    user_uuid: ''
  })

  // Reactive translation rules
  const formRules = computed(() => ({
    appName: [{ required: true, message: '请选择应用名称', trigger: 'change' }],
    platform: [{ required: true, message: '请选择平台', trigger: 'change' }],
    version: [{ required: true, message: '请选择版本', trigger: 'change' }],
    token_id: [{ required: true, message: '请选择 Access Token', trigger: 'change' }],
    user_uuid: [{ required: true, message: t('blacklistManage.ruleUserUuid'), trigger: 'blur' }]
  }))

  const addFormRef = ref()

  // Computed cascaded options
  const appNameOptions = computed(() => {
    const set = new Set<string>()
    allTokensList.value.forEach((item) => {
      if (item.app_name) set.add(item.app_name)
    })
    return Array.from(set)
  })

  const platformOptions = computed(() => {
    if (!addForm.appName) return []
    const set = new Set<string>()
    allTokensList.value
      .filter((item) => item.app_name === addForm.appName)
      .forEach((item) => {
        if (item.platform) set.add(item.platform)
      })
    return Array.from(set)
  })

  const versionOptions = computed(() => {
    if (!addForm.appName || !addForm.platform) return []
    const set = new Set<string>()
    allTokensList.value
      .filter((item) => item.app_name === addForm.appName && item.platform === addForm.platform)
      .forEach((item) => {
        if (item.version) set.add(item.version)
      })
    return Array.from(set)
  })

  const tokenOptions = computed(() => {
    if (!addForm.appName || !addForm.platform || !addForm.version) return []
    return allTokensList.value.filter(
      (item) =>
        item.app_name === addForm.appName &&
        item.platform === addForm.platform &&
        item.version === addForm.version
    )
  })

  const handleAppNameChange = () => {
    addForm.platform = ''
    addForm.version = ''
    addForm.token_id = undefined
    addForm.token = ''
  }

  const handlePlatformChange = () => {
    addForm.version = ''
    addForm.token_id = undefined
    addForm.token = ''
  }

  const handleVersionChange = () => {
    addForm.token_id = undefined
    addForm.token = ''
  }

  const handleTokenChange = (val: number) => {
    const found = allTokensList.value.find((item) => item.id === val)
    if (found) {
      addForm.token = found.token
    }
  }

  // Use the useTableColumns hook to manage visible/hidden columns, table checks and icons
  const { columns, columnChecks } = useTableColumns(() => [
    { type: 'globalIndex', label: t('blacklistManage.index'), width: 80, align: 'center' },
    { prop: 'app_name', label: t('blacklistManage.appName'), minWidth: 140 },
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

  const openAddDialog = async () => {
    addForm.appName = ''
    addForm.platform = ''
    addForm.version = ''
    addForm.token_id = undefined
    addForm.token = ''
    addForm.user_uuid = ''

    try {
      const res = await fetchGetTokens()
      allTokensList.value = Array.isArray(res) ? res : res?.tokens || []
    } catch (e: any) {
      ElMessage.error('加载 Token 选项失败: ' + (e.message || ''))
    }

    addDialogVisible.value = true
  }

  const submitAdd = async () => {
    addFormRef.value?.validate(async (valid: boolean) => {
      if (!valid) return
      try {
        await fetchAddBlacklist({
          token_id: addForm.token_id,
          token: addForm.token,
          user_uuid: addForm.user_uuid
        })
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
