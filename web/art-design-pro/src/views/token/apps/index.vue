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
      <ElForm :model="registerForm" label-width="120px" :rules="formRules" ref="registerFormRef">
        <ElFormItem :label="t('tokenManage.appId')" prop="app_id">
          <ElInput v-model="registerForm.app_id" :placeholder="t('tokenManage.placeholderAppId')" />
        </ElFormItem>
        <ElFormItem :label="t('tokenManage.appName')" prop="name">
          <ElInput v-model="registerForm.name" :placeholder="t('tokenManage.placeholderAppName')" />
        </ElFormItem>
        <ElFormItem :label="t('tokenManage.version')" prop="version">
          <ElInput
            v-model="registerForm.version"
            :placeholder="t('tokenManage.placeholderVersion')"
          />
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
      <ElForm :model="tokenForm" label-width="120px">
        <ElFormItem :label="t('tokenManage.appName')">
          <ElInput :value="`${tokenForm.app_id} (v${tokenForm.version})`" disabled />
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

    <!-- Drawer: Tokens List -->
    <ElDrawer
      v-model="tokensDrawerVisible"
      :title="`${t('tokenManage.manage')}: ${currentApp.app_id} (v${currentApp.version})`"
      size="60%"
    >
      <div v-loading="drawerLoading" class="p-4">
        <ElTable :data="tokens" stripe style="width: 100%">
          <ElTableColumn prop="platform" :label="t('tokenManage.platform')" width="100" />
          <ElTableColumn prop="token" label="Token" min-width="180" show-overflow-tooltip />
          <ElTableColumn :label="t('tokenManage.createdAt')" width="160">
            <template #default="scope">
              {{ formatTime(scope.row.created_at) }}
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('tokenManage.statusLabel')" width="100">
            <template #default="scope">
              <ElTag :type="scope.row.revoked ? 'danger' : 'success'">
                {{ scope.row.revoked ? t('tokenManage.revoked') : t('tokenManage.active') }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn :label="t('tokenManage.operations')" width="120" fixed="right">
            <template #default="scope">
              <ElButton
                v-if="!scope.row.revoked"
                size="small"
                type="danger"
                @click="revokeToken(scope.row)"
              >
                {{ t('tokenManage.revokeBtn') }}
              </ElButton>
              <span v-else class="text-gray-500">-</span>
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
  import { useWindowSize } from '@vueuse/core'
  import {
    fetchGetApps,
    fetchRegisterApp,
    fetchToggleApp,
    fetchDeleteApp,
    fetchGetTokens,
    fetchGenerateToken,
    fetchRevokeToken
  } from '@/api/token'

  defineOptions({ name: 'Apps' })

  const { t } = useI18n()
  const { height: windowHeight } = useWindowSize()

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

  const generatedToken = ref('')
  const currentApp = ref<any>({})

  // Form states
  const registerForm = reactive({
    app_id: '',
    name: '',
    version: ''
  })

  const tokenForm = reactive({
    app_id: '',
    version: '',
    platform: 'android'
  })

  // Reactive translation rules
  const formRules = computed(() => ({
    app_id: [{ required: true, message: t('tokenManage.ruleAppId'), trigger: 'blur' }],
    name: [{ required: true, message: t('tokenManage.ruleAppName'), trigger: 'blur' }],
    version: [{ required: true, message: t('tokenManage.ruleVersion'), trigger: 'blur' }]
  }))

  const registerFormRef = ref()

  // Use the useTableColumns hook to manage visible/hidden columns, table checks and icons
  const { columns, columnChecks } = useTableColumns(() => [
    { type: 'globalIndex', label: t('tokenManage.index'), width: 80, align: 'center' },
    { prop: 'app_id', label: t('tokenManage.appId'), minWidth: 150 },
    { prop: 'name', label: t('tokenManage.appName'), minWidth: 180 },
    {
      prop: 'version',
      label: t('tokenManage.version'),
      width: 120,
      useSlot: true,
      slotName: 'version'
    },
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
    registerForm.version = ''
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
        version: row.version,
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
      t('tokenManage.deleteConfirm', { name: row.name, version: row.version }),
      t('tokenManage.warning'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    ).then(async () => {
      try {
        await fetchDeleteApp({ app_id: row.app_id, version: row.version })
        ElMessage.success(t('tokenManage.successDelete'))
        loadApps()
      } catch (e: any) {
        ElMessage.error(e.message || t('tokenManage.errorDelete'))
      }
    })
  }

  const openGenerateTokenDialog = (row: any) => {
    tokenForm.app_id = row.app_id
    tokenForm.version = row.version
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
    loadTokens(row.app_id, row.version)
  }

  const loadTokens = async (appId: string, version: string) => {
    drawerLoading.value = true
    try {
      const res = await fetchGetTokens({ app_id: appId, version })
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
        loadTokens(currentApp.value.app_id, currentApp.value.version)
        loadApps()
      } catch (e: any) {
        ElMessage.error(e.message || t('tokenManage.errorRevoke'))
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
