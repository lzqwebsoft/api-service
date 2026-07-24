<template>
  <div class="apps-page art-full-height">
    <ElCard class="art-table-card">
      <template #header>
        <div class="flex-cb">
          <div class="title-group">
            <h4 class="m-0 font-semibold text-lg">{{ t('tokenManage.title') }}</h4>
            <p class="text-gray-400 text-xs mt-1">{{ t('tokenManage.subtitle') }}</p>
          </div>
          <div class="flex gap-2">
            <ElTag type="success">{{ t('tokenManage.dataCount', { count: apps.length }) }}</ElTag>
          </div>
        </div>
      </template>

      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadApps">
        <template #left>
          <ElSpace wrap>
            <ElButton type="primary" @click="openRegisterDialog" v-ripple>
              <i class="fas fa-plus mr-1"></i> {{ t('tokenManage.register') }}
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- Table of Apps -->
      <ArtTable
        v-loading="loading"
        :data="apps"
        :columns="columns"
        :show-pagination="false"
        :height="computedTableHeight"
        empty-height="360px"
      >
        <!-- Custom slot: version -->
        <template #version="{ row }">
          <ElTag size="small">{{ row.version }}</ElTag>
        </template>

        <!-- Custom slot: isActive -->
        <template #isActive="{ row }">
          <ElSwitch v-model="row.is_active" @change="toggleAppStatus(row)" />
        </template>

        <!-- Custom slot: tokenCount -->
        <template #tokenCount="{ row }">
          <ElTag type="info">{{ row.token_count }}</ElTag>
        </template>

        <!-- Custom slot: operation -->
        <template #operation="{ row }">
          <ElTooltip :content="t('tokenManage.generate')" placement="top">
            <ArtButtonTable
              icon="ri:key-2-line"
              icon-class="bg-theme/12 text-theme"
              @click="openGenerateTokenDialog(row)"
            />
          </ElTooltip>
          <ElTooltip :content="t('tokenManage.manage')" placement="top">
            <ArtButtonTable type="view" @click="openTokensDrawer(row)" />
          </ElTooltip>
          <ElTooltip :content="t('tokenManage.delete')" placement="top">
            <ArtButtonTable type="delete" @click="deleteAppVersion(row)" />
          </ElTooltip>
        </template>
      </ArtTable>
    </ElCard>

    <!-- Dialog: Register App -->
    <ElDialog v-model="registerDialogVisible" :title="t('tokenManage.register')" width="500px">
      <ElForm :model="registerForm" label-width="140px" :rules="formRules" ref="registerFormRef">
        <ElFormItem :label="t('tokenManage.appId')" prop="app_id">
          <ElInput v-model="registerForm.app_id" :placeholder="t('tokenManage.placeholderAppId')" />
        </ElFormItem>
        <ElFormItem :label="t('tokenManage.appName')" prop="name">
          <ElInput v-model="registerForm.name" :placeholder="t('tokenManage.placeholderAppName')" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <span class="dialog-footer">
          <ElButton @click="registerDialogVisible = false">{{ t('common.cancel') }}</ElButton>
          <ElButton type="primary" @click="submitRegister">{{ t('common.confirm') }}</ElButton>
        </span>
      </template>
    </ElDialog>

    <!-- Dialog: Generate Token -->
    <ElDialog v-model="generateTokenDialogVisible" :title="t('tokenManage.generate')" width="500px">
      <ElForm :model="tokenForm" label-width="140px">
        <ElFormItem :label="t('tokenManage.appId')">
          <ElInput :value="tokenForm.app_id" disabled />
        </ElFormItem>
        <ElFormItem :label="t('tokenManage.selectPlatform')">
          <ElSelect
            v-model="tokenForm.platform"
            :placeholder="t('tokenManage.placeholderPlatform')"
            style="width: 100%"
          >
            <ElOption label="android" value="android" />
            <ElOption label="iOS" value="iOS" />
            <ElOption label="windows" value="windows" />
            <ElOption label="Linux" value="Linux" />
            <ElOption label="mac" value="mac" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('tokenManage.versionOperator')">
          <ElSelect
            v-model="tokenForm.version_operator"
            :placeholder="t('tokenManage.placeholderVersionOperator')"
            style="width: 100%"
          >
            <ElOption label="=" value="=" />
            <ElOption label=">=" value=">=" />
            <ElOption label=">" value=">" />
            <ElOption label="<=" value="<=" />
            <ElOption label="<" value="<" />
            <ElOption label="all" value="all" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('tokenManage.version')">
          <ElInput v-model="tokenForm.version" :placeholder="t('tokenManage.placeholderVersion')" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <span class="dialog-footer">
          <ElButton @click="generateTokenDialogVisible = false">{{ t('common.cancel') }}</ElButton>
          <ElButton type="primary" @click="submitGenerateToken">{{
            t('tokenManage.generate')
          }}</ElButton>
        </span>
      </template>
    </ElDialog>

    <!-- Dialog: Generated Token Result -->
    <ElDialog
      v-model="tokenResultDialogVisible"
      :title="t('tokenManage.copySuccess')"
      width="500px"
    >
      <div class="mb-4 text-sm text-gray-400">
        {{ t('tokenManage.copyTip') }}
      </div>
      <div
        class="flex items-center gap-2 p-3 bg-zinc-800 rounded border border-zinc-700 font-mono text-emerald-400 break-all select-all"
      >
        <span>{{ generatedToken }}</span>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <ElButton type="primary" @click="tokenResultDialogVisible = false">{{
            t('tokenManage.done')
          }}</ElButton>
        </span>
      </template>
    </ElDialog>

    <!-- Dialog: Edit Token Version Constraint -->
    <ElDialog
      v-model="editTokenDialogVisible"
      :title="t('tokenManage.editTokenTitle')"
      width="500px"
    >
      <ElForm :model="editTokenForm" label-width="140px">
        <ElFormItem :label="t('tokenManage.versionOperator')">
          <ElSelect
            v-model="editTokenForm.version_operator"
            :placeholder="t('tokenManage.placeholderVersionOperator')"
            style="width: 100%"
          >
            <ElOption label="=" value="=" />
            <ElOption label=">=" value=">=" />
            <ElOption label=">" value=">" />
            <ElOption label="<=" value="<=" />
            <ElOption label="<" value="<" />
            <ElOption label="all" value="all" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('tokenManage.version')">
          <ElInput
            v-model="editTokenForm.version"
            :placeholder="t('tokenManage.placeholderVersion')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <span class="dialog-footer">
          <ElButton @click="editTokenDialogVisible = false">{{ t('common.cancel') }}</ElButton>
          <ElButton type="primary" @click="submitUpdateTokenVersion">{{
            t('common.confirm')
          }}</ElButton>
        </span>
      </template>
    </ElDialog>

    <!-- Drawer: Tokens List -->
    <ElDrawer
      v-model="tokensDrawerVisible"
      :title="`${t('tokenManage.manage')}: ${currentApp.app_id} (${currentApp.name})`"
      size="60%"
    >
      <div v-loading="drawerLoading" class="p-4">
        <ElTable
          :data="tokens"
          stripe
          border
          style="width: 100%"
          class="rounded-lg overflow-hidden"
        >
          <!-- Platform -->
          <ElTableColumn
            prop="platform"
            :label="t('tokenManage.platform')"
            width="90"
            align="center"
          >
            <template #default="scope">
              <ElTooltip :content="scope.row.platform" placement="top">
                <ElTag
                  :type="getPlatformTagType(scope.row.platform)"
                  size="small"
                  class="capitalize font-medium cursor-pointer"
                >
                  <ArtSvgIcon :icon="getPlatformIcon(scope.row.platform)" class="text-xs" />
                </ElTag>
              </ElTooltip>
            </template>
          </ElTableColumn>

          <!-- Version Constraint -->
          <ElTableColumn :label="t('tokenManage.version')" width="130" align="center">
            <template #default="scope">
              <div class="relative flex items-center justify-center w-full px-4">
                <ElTag size="small" type="info">
                  {{ scope.row.version_operator || '=' }} {{ scope.row.version || '*' }}
                </ElTag>
                <ElButton
                  v-if="!scope.row.is_revoked"
                  circle
                  size="small"
                  type="primary"
                  link
                  class="absolute right-0"
                  :title="t('tokenManage.editTokenTitle')"
                  @click="openEditTokenDialog(scope.row)"
                >
                  <ArtSvgIcon icon="ri:edit-line" class="text-xs" />
                </ElButton>
              </div>
            </template>
          </ElTableColumn>

          <!-- Token -->
          <ElTableColumn prop="token" label="Token (Access Key)" min-width="220">
            <template #default="scope">
              <div class="flex items-center justify-between gap-2 group">
                <span
                  class="font-mono text-xs text-gray-700 dark:text-gray-300 bg-gray-50 dark:bg-zinc-800 px-2 py-1 rounded border border-gray-200 dark:border-zinc-700 truncate select-all"
                  :title="scope.row.token"
                >
                  {{ scope.row.token }}
                </span>
                <ElButton
                  circle
                  size="small"
                  type="primary"
                  link
                  class="opacity-60 group-hover:opacity-100 transition-opacity flex-shrink-0"
                  :title="t('common.copy')"
                  @click="copyToken(scope.row.token)"
                >
                  <ArtSvgIcon icon="ri:file-copy-line" class="text-xs" />
                </ElButton>
              </div>
            </template>
          </ElTableColumn>

          <!-- Created At -->
          <ElTableColumn :label="t('tokenManage.createdAt')" width="120" align="center">
            <template #default="scope">
              <span
                class="text-xs text-gray-500 dark:text-gray-400 flex items-center justify-center gap-1"
              >
                {{ formatTime(scope.row.created_at) }}
              </span>
            </template>
          </ElTableColumn>

          <!-- Status -->
          <ElTableColumn :label="t('tokenManage.statusLabel')" width="95" align="center">
            <template #default="scope">
              <ElTag
                :type="scope.row.is_revoked ? 'danger' : 'success'"
                effect="light"
                size="small"
                class="rounded-full px-3 font-medium"
              >
                <i
                  class="fas fa-circle text-[6px] mr-1"
                  :class="scope.row.is_revoked ? 'text-rose-500' : 'text-emerald-500'"
                ></i>
                {{ scope.row.is_revoked ? t('tokenManage.revoked') : t('tokenManage.active') }}
              </ElTag>
            </template>
          </ElTableColumn>

          <!-- Operations -->
          <ElTableColumn
            :label="t('tokenManage.operations')"
            width="100"
            fixed="right"
            align="center"
          >
            <template #default="scope">
              <ElButton
                v-if="!scope.row.is_revoked"
                size="small"
                type="danger"
                class="font-medium hover:opacity-80"
                @click="revokeToken(scope.row)"
              >
                <ArtSvgIcon icon="ri:forbid-line" class="mr-1 text-xs" />
                {{ t('tokenManage.revokeBtn') }}
              </ElButton>
              <span v-else class="text-gray-400 text-xs italic">-</span>
            </template>
          </ElTableColumn>
        </ElTable>
      </div>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, onMounted, computed } from 'vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useWindowSize, useClipboard } from '@vueuse/core'
  import {
    fetchGetApps,
    fetchRegisterApp,
    fetchToggleApp,
    fetchDeleteApp,
    fetchGetTokens,
    fetchGenerateToken,
    fetchRevokeToken,
    fetchUpdateTokenVersion
  } from '@/api/token'

  defineOptions({ name: 'Apps' })

  const { t } = useI18n()
  const { height: windowHeight } = useWindowSize()
  const { copy } = useClipboard()

  const computedTableHeight = computed(() => {
    const val = windowHeight.value - 360
    return val > 300 ? val : 300
  })

  const loading = ref(false)
  const drawerLoading = ref(false)
  const apps = ref<any[]>([])
  const tokens = ref<any[]>([])

  // Dialog / Drawer visibility states
  const registerDialogVisible = ref(false)
  const generateTokenDialogVisible = ref(false)
  const tokenResultDialogVisible = ref(false)
  const tokensDrawerVisible = ref(false)
  const editTokenDialogVisible = ref(false)

  const editTokenForm = reactive({
    id: 0,
    version: '',
    version_operator: '='
  })

  const generatedToken = ref('')
  const currentApp = ref<any>({})

  // Form states
  const registerForm = reactive({
    app_id: '',
    name: ''
  })

  const tokenForm = reactive({
    app_id: '',
    version: '',
    version_operator: '=',
    platform: 'android'
  })

  // Reactive translation rules
  const formRules = computed(() => ({
    app_id: [{ required: true, message: t('tokenManage.ruleAppId'), trigger: 'blur' }],
    name: [{ required: true, message: t('tokenManage.ruleAppName'), trigger: 'blur' }]
  }))

  const registerFormRef = ref()

  // Use the useTableColumns hook to manage visible/hidden columns, table checks and icons
  const { columns, columnChecks } = useTableColumns(() => [
    { type: 'globalIndex', label: t('tokenManage.index'), width: 80, align: 'center' },
    { prop: 'app_id', label: t('tokenManage.appId'), minWidth: 150 },
    { prop: 'name', label: t('tokenManage.appName'), minWidth: 180 },
    {
      prop: 'is_active',
      label: t('tokenManage.status'),
      width: 120,
      useSlot: true,
      slotName: 'isActive'
    },
    {
      prop: 'token_count',
      label: t('tokenManage.tokenCount'),
      width: 120,
      useSlot: true,
      slotName: 'tokenCount'
    },
    {
      prop: 'operation',
      label: t('tokenManage.operations'),
      minWidth: 250,
      fixed: 'right',
      useSlot: true,
      slotName: 'operation'
    }
  ])

  onMounted(() => {
    loadApps()
  })

  const loadApps = async () => {
    loading.value = true
    try {
      const res = await fetchGetApps()
      apps.value = res || []
    } catch (e: any) {
      ElMessage.error(e.message || t('tokenManage.errorLoadApps'))
    } finally {
      loading.value = false
    }
  }

  const openRegisterDialog = () => {
    registerForm.app_id = ''
    registerForm.name = ''
    registerDialogVisible.value = true
  }

  const submitRegister = async () => {
    registerFormRef.value?.validate(async (valid: boolean) => {
      if (!valid) return
      try {
        await fetchRegisterApp(registerForm)
        ElMessage.success(t('tokenManage.successRegister'))
        registerDialogVisible.value = false
        loadApps()
      } catch (e: any) {
        ElMessage.error(e.message || t('tokenManage.errorRegister'))
      }
    })
  }

  const toggleAppStatus = async (row: any) => {
    try {
      await fetchToggleApp({
        app_id: row.app_id,
        is_active: row.is_active
      })
      ElMessage.success(row.is_active ? t('tokenManage.appEnabled') : t('tokenManage.appDisabled'))
    } catch (e: any) {
      row.is_active = !row.is_active // Revert UI
      ElMessage.error(e.message || t('tokenManage.errorToggle'))
    }
  }

  const deleteAppVersion = (row: any) => {
    ElMessageBox.confirm(
      t('tokenManage.deleteConfirmAppOnly', { name: row.name }),
      t('tokenManage.warning'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    ).then(async () => {
      try {
        await fetchDeleteApp({ app_id: row.app_id })
        ElMessage.success(t('tokenManage.successDelete'))
        loadApps()
      } catch (e: any) {
        ElMessage.error(e.message || t('tokenManage.errorDelete'))
      }
    })
  }

  const openGenerateTokenDialog = (row: any) => {
    tokenForm.app_id = row.app_id
    tokenForm.version = ''
    tokenForm.version_operator = '='
    tokenForm.platform = 'android'
    generateTokenDialogVisible.value = true
  }

  const submitGenerateToken = async () => {
    try {
      const res = await fetchGenerateToken(tokenForm)
      generatedToken.value = res.token
      generateTokenDialogVisible.value = false
      tokenResultDialogVisible.value = true
      loadApps()
    } catch (e: any) {
      ElMessage.error(e.message || t('tokenManage.errorGenerate'))
    }
  }

  const openTokensDrawer = async (row: any) => {
    currentApp.value = row
    tokensDrawerVisible.value = true
    loadTokens(row.app_id)
  }

  const loadTokens = async (appId: string) => {
    drawerLoading.value = true
    try {
      const res = await fetchGetTokens({ app_id: appId })
      tokens.value = res.tokens || []
    } catch (e: any) {
      ElMessage.error(e.message || t('tokenManage.errorLoadTokens'))
    } finally {
      drawerLoading.value = false
    }
  }

  const revokeToken = (row: any) => {
    ElMessageBox.confirm(t('tokenManage.revokeConfirm'), t('tokenManage.warning'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    }).then(async () => {
      try {
        await fetchRevokeToken({ token: row.token })
        ElMessage.success(t('tokenManage.successRevoke'))
        row.is_revoked = true
        loadTokens(currentApp.value.app_id)
      } catch (e: any) {
        ElMessage.error(e.message || t('tokenManage.errorRevoke'))
      }
    })
  }

  const openEditTokenDialog = (row: any) => {
    editTokenForm.id = row.id
    editTokenForm.version = row.version
    editTokenForm.version_operator = row.version_operator || '='
    editTokenDialogVisible.value = true
  }

  const submitUpdateTokenVersion = async () => {
    try {
      await fetchUpdateTokenVersion(editTokenForm)
      ElMessage.success(t('tokenManage.successUpdateTokenVersion'))
      editTokenDialogVisible.value = false
      loadTokens(currentApp.value.app_id)
    } catch (e: any) {
      ElMessage.error(e.message || t('tokenManage.errorUpdateTokenVersion'))
    }
  }

  const formatTime = (timeStr: string) => {
    if (!timeStr) return '-'
    const d = new Date(timeStr)
    return d.toLocaleString()
  }

  const copyToken = async (text: string) => {
    try {
      await copy(text)
      ElMessage.success(t('common.copySuccess'))
    } catch {
      ElMessage.error(t('common.copyFailed'))
    }
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
