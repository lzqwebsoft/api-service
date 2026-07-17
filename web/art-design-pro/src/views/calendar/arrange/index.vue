<template>
  <div class="calendar-arrange-page flex flex-col gap-4 pb-5">
    <!-- Stats Row -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
      <ArtStatsCard
        :title="t('calendarArrange.totalExceptions')"
        :count="stats.totalCount"
        icon="ri:calendar-line"
        icon-style="bg-blue-500"
        box-style="stats-card-compact"
        description=""
      />
      <ArtStatsCard
        :title="t('calendarArrange.holidaysCountLabel')"
        :count="stats.holidayCount"
        icon="ri:umbrella-line"
        icon-style="bg-red-500"
        box-style="stats-card-compact"
        description=""
      />
      <ArtStatsCard
        :title="t('calendarArrange.workdaysCountLabel')"
        :count="stats.workdayCount"
        icon="ri:briefcase-line"
        icon-style="bg-green-500"
        box-style="stats-card-compact"
        description=""
      />
    </div>

    <!-- 搜索区域 -->
    <ArtSearchBar
      ref="searchBarRef"
      v-model="searchFormState"
      :items="searchItems"
      :rules="searchRules"
      :is-expand="false"
      :show-expand="true"
      :show-reset-button="true"
      :show-search-button="true"
      :disabled-search-button="false"
      :button-left-limit="1"
      label-width="90px"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- Years Filter Row (Collapsible) -->
    <ElCard shadow="never" class="mb-4 mt-4" :body-style="{ padding: '16px' }">
      <div class="flex flex-col gap-2">
        <div
          class="flex items-center justify-between cursor-pointer select-none"
          @click="isYearsExpanded = !isYearsExpanded"
        >
          <div class="flex items-center gap-2">
            <span class="text-gray-600 text-sm font-medium">{{
              t('calendarArrange.yearsCovered')
            }}</span>
            <span class="text-gray-400 text-xs font-normal"
              >({{ t('calendarArrange.clickToFilter') }})</span
            >
            <ElTag
              v-if="selectedYear !== undefined"
              size="small"
              closable
              @close.stop="filterByYear(undefined)"
            >
              {{ selectedYear }}
            </ElTag>
          </div>
          <div
            class="flex items-center gap-1 text-gray-400 hover:text-gray-600 text-xs font-normal transition-colors"
          >
            <span>{{
              isYearsExpanded ? t('calendarArrange.collapse') : t('calendarArrange.expand')
            }}</span>
            <ArtSvgIcon
              :icon="isYearsExpanded ? 'ri:arrow-up-s-line' : 'ri:arrow-down-s-line'"
              class="text-sm"
            />
          </div>
        </div>
        <div
          v-show="isYearsExpanded"
          class="flex flex-wrap gap-2 mt-1 border-t border-gray-50 pt-2 transition-all duration-200"
        >
          <ElTag
            size="default"
            :effect="selectedYear === undefined ? 'dark' : 'plain'"
            :type="selectedYear === undefined ? 'primary' : 'info'"
            class="cursor-pointer select-none transition-all"
            @click="filterByYear(undefined)"
          >
            {{ t('calendarArrange.all') }}
          </ElTag>
          <ElTag
            v-for="year in stats.years"
            :key="year"
            size="default"
            :effect="selectedYear === Number(year) ? 'dark' : 'plain'"
            :type="selectedYear === Number(year) ? 'primary' : 'info'"
            class="cursor-pointer select-none transition-all"
            @click="filterByYear(Number(year))"
          >
            {{ year }}
          </ElTag>
        </div>
      </div>
    </ElCard>

    <ElCard class="art-table-card">
      <template #header>
        <div class="flex-cb">
          <div class="title-group">
            <h4 class="m-0 font-semibold text-lg">{{ t('calendarArrange.title') }}</h4>
            <p class="text-gray-400 text-xs mt-1">{{ t('calendarArrange.subtitle') }}</p>
          </div>
          <div class="flex gap-2">
            <ElTag type="success">
              {{ t('calendarArrange.dataCount', { count: pagination.total }) }}
            </ElTag>
          </div>
        </div>
      </template>

      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadExceptions">
        <template #left>
          <ElSpace wrap>
            <ElButton type="primary" @click="openAddDialog" v-ripple>
              <i class="fas fa-plus mr-1"></i> {{ t('calendarArrange.addBtn') }}
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- Table of Calendar exceptions -->
      <ArtTable
        v-loading="loading"
        :data="exceptions"
        :columns="columns"
        :pagination="pagination"
        :height="computedTableHeight"
        empty-height="360px"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
        <!-- Custom slot: region -->
        <template #region="{ row }">
          <ElTag size="small" type="info">
            {{ t('calendarArrange.' + row.region) }}
          </ElTag>
        </template>

        <!-- Custom slot: type -->
        <template #type="{ row }">
          <ElTag :type="row.is_workday ? 'success' : 'danger'">
            {{ row.is_workday ? t('calendarArrange.workday') : t('calendarArrange.holiday') }}
          </ElTag>
        </template>

        <!-- Custom slot: operation -->
        <template #operation="{ row }">
          <ArtButtonTable type="edit" @click="openEditDialog(row)" />
          <ArtButtonTable type="delete" @click="deleteException(row)" />
        </template>
      </ArtTable>
    </ElCard>

    <!-- Dialog: Add / Edit Exception -->
    <ElDialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? t('calendarArrange.addBtn') : t('calendarArrange.editBtn')"
      width="500px"
    >
      <ElForm :model="form" label-width="100px" :rules="formRules" ref="formRef">
        <ElFormItem :label="t('calendarArrange.date')" prop="date">
          <ElDatePicker
            v-model="form.date"
            type="date"
            :placeholder="t('calendarArrange.placeholderDate')"
            value-format="YYYY-MM-DD"
            style="width: 100%"
            :disabled="dialogType === 'edit'"
          />
        </ElFormItem>
        <ElFormItem :label="t('calendarArrange.region')" prop="region">
          <ElSelect
            v-model="form.region"
            :placeholder="t('calendarArrange.placeholderRegion')"
            style="width: 100%"
            :disabled="dialogType === 'edit'"
          >
            <ElOption :label="t('calendarArrange.cn')" value="cn" />
            <ElOption :label="t('calendarArrange.hk')" value="hk" />
            <ElOption :label="t('calendarArrange.tw')" value="tw" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('calendarArrange.type')" prop="is_workday">
          <ElSelect
            v-model="form.is_workday"
            :placeholder="t('calendarArrange.placeholderType')"
            style="width: 100%"
          >
            <ElOption :label="t('calendarArrange.holiday')" :value="false" />
            <ElOption :label="t('calendarArrange.workday')" :value="true" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('calendarArrange.description')" prop="description">
          <ElInput
            v-model="form.description"
            type="textarea"
            :placeholder="t('calendarArrange.placeholderDescription')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <span class="dialog-footer">
          <ElButton @click="dialogVisible = false">{{ t('common.cancel') }}</ElButton>
          <ElButton type="primary" @click="submitForm">{{ t('common.confirm') }}</ElButton>
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
  import ArtStatsCard from '@/components/core/cards/art-stats-card/index.vue'
  import {
    fetchGetCalendarList,
    fetchAddCalendar,
    fetchUpdateCalendar,
    fetchDeleteCalendar
  } from '@/api/calendar'

  defineOptions({ name: 'Arrange' })

  const { t } = useI18n()

  const computedTableHeight = computed(() => {
    return ''
  })

  const loading = ref(false)
  const exceptions = ref<any[]>([])

  // Pagination state
  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  // Statistics state
  const stats = reactive({
    totalCount: 0,
    holidayCount: 0,
    workdayCount: 0,
    years: [] as string[]
  })

  const selectedYear = ref<number | undefined>(undefined)
  const isYearsExpanded = ref(true)

  // Search Area configuration
  const searchBarRef = ref()
  const searchFormState = reactive({
    region: 'all',
    is_workday: undefined
  })

  const searchItems = computed(() => [
    {
      label: t('calendarArrange.region'),
      key: 'region',
      type: 'select',
      props: {
        placeholder: '请选择地区',
        clearable: false,
        options: [
          { label: t('calendarArrange.all'), value: 'all' },
          { label: t('calendarArrange.cn'), value: 'cn' },
          { label: t('calendarArrange.hk'), value: 'hk' },
          { label: t('calendarArrange.tw'), value: 'tw' }
        ]
      }
    },
    {
      label: t('calendarArrange.type'),
      key: 'is_workday',
      type: 'select',
      props: {
        placeholder: '请选择工作性质',
        clearable: true,
        options: [
          { label: t('calendarArrange.holiday'), value: false },
          { label: t('calendarArrange.workday'), value: true }
        ]
      }
    }
  ])

  const searchRules = reactive({})

  const handleSearch = () => {
    pagination.current = 1
    loadExceptions()
  }

  const handleReset = () => {
    searchFormState.region = 'all'
    searchFormState.is_workday = undefined
    selectedYear.value = undefined
    pagination.current = 1
    loadExceptions()
  }

  const filterByYear = (year: number | undefined) => {
    selectedYear.value = year
    pagination.current = 1
    loadExceptions()
  }

  // Dialog and Form configs
  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')

  const form = reactive({
    date: '',
    region: 'cn',
    is_workday: false,
    description: ''
  })

  const formRules = computed(() => ({
    date: [{ required: true, message: t('calendarArrange.ruleDate'), trigger: 'change' }],
    region: [{ required: true, message: t('calendarArrange.ruleRegion'), trigger: 'change' }],
    description: [
      { required: true, message: t('calendarArrange.ruleDescription'), trigger: 'blur' }
    ]
  }))

  const formRef = ref()

  // Use the useTableColumns hook to manage visible/hidden columns, table checks and icons
  const { columns, columnChecks } = useTableColumns(() => [
    { prop: 'date', label: t('calendarArrange.date'), minWidth: 120 },
    {
      prop: 'region',
      label: t('calendarArrange.region'),
      width: 100,
      useSlot: true,
      slotName: 'region'
    },
    {
      prop: 'is_workday',
      label: t('calendarArrange.type'),
      width: 120,
      useSlot: true,
      slotName: 'type'
    },
    { prop: 'description', label: t('calendarArrange.description'), minWidth: 200 },
    {
      prop: 'operation',
      label: t('calendarArrange.operations'),
      width: 150,
      fixed: 'right',
      useSlot: true,
      slotName: 'operation'
    }
  ])

  onMounted(() => {
    loadExceptions()
  })

  const loadExceptions = async () => {
    loading.value = true
    try {
      const params: any = {
        current: pagination.current,
        size: pagination.size
      }
      if (searchFormState.region !== 'all') {
        params.region = searchFormState.region
      }
      if (searchFormState.is_workday !== undefined && searchFormState.is_workday !== null) {
        params.is_workday = searchFormState.is_workday
      }
      if (selectedYear.value !== undefined) {
        params.year = selectedYear.value
      }
      const res = await fetchGetCalendarList(params)
      exceptions.value = res.list || []
      pagination.total = res.total || 0

      // Update statistics
      stats.totalCount = res.totalCount || 0
      stats.holidayCount = res.holidayCount || 0
      stats.workdayCount = res.workdayCount || 0
      stats.years = res.years || []
    } catch (e: any) {
      ElMessage.error(e.message || t('calendarArrange.errorLoad'))
    } finally {
      loading.value = false
    }
  }

  const handleSizeChange = (val: number) => {
    pagination.size = val
    pagination.current = 1
    loadExceptions()
  }

  const handleCurrentChange = (val: number) => {
    pagination.current = val
    loadExceptions()
  }

  const openAddDialog = () => {
    dialogType.value = 'add'
    form.date = ''
    form.region = searchFormState.region === 'all' ? 'cn' : searchFormState.region
    form.is_workday = false
    form.description = ''
    dialogVisible.value = true
  }

  const openEditDialog = (row: any) => {
    dialogType.value = 'edit'
    form.date = row.date
    form.region = row.region
    form.is_workday = row.is_workday
    form.description = row.description
    dialogVisible.value = true
  }

  const submitForm = async () => {
    formRef.value?.validate(async (valid: boolean) => {
      if (!valid) return
      try {
        if (dialogType.value === 'add') {
          await fetchAddCalendar(form)
          ElMessage.success(t('calendarArrange.successAdd'))
        } else {
          await fetchUpdateCalendar(form)
          ElMessage.success(t('calendarArrange.successUpdate'))
        }
        dialogVisible.value = false
        loadExceptions()
      } catch (e: any) {
        ElMessage.error(e.message || t('calendarArrange.errorSubmit'))
      }
    })
  }

  const deleteException = (row: any) => {
    ElMessageBox.confirm(
      t('calendarArrange.deleteConfirm', { date: row.date }),
      t('calendarArrange.warning'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    ).then(async () => {
      try {
        await fetchDeleteCalendar({ date: row.date, region: row.region })
        ElMessage.success(t('calendarArrange.successDelete'))
        loadExceptions()
      } catch (e: any) {
        ElMessage.error(e.message || t('calendarArrange.errorDelete'))
      }
    })
  }
</script>

<style scoped>
  :deep(.stats-card-compact) {
    height: 6rem !important;
    padding-left: 0.75rem !important;
    padding-right: 0.75rem !important;
  }
</style>
