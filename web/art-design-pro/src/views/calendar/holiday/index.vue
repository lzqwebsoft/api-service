<template>
  <div class="holiday-definition-page flex flex-col gap-4 pb-5">
    <!-- Stats Row -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
      <ArtStatsCard
        :title="t('holiday.statsTotal')"
        :count="stats.totalCount"
        icon="ri:umbrella-line"
        icon-style="bg-blue-500"
        box-style="stats-card-compact"
        description=""
      />
      <ArtStatsCard
        :title="t('holiday.statsSolar')"
        :count="stats.solarCount"
        icon="ri:calendar-line"
        icon-style="bg-red-500"
        box-style="stats-card-compact"
        description=""
      />
      <ArtStatsCard
        :title="t('holiday.statsWeekday')"
        :count="stats.weekdayCount"
        icon="ri:calendar-todo-line"
        icon-style="bg-purple-500"
        box-style="stats-card-compact"
        description=""
      />
      <ArtStatsCard
        :title="t('holiday.statsIndustry')"
        :count="stats.industryCount"
        icon="ri:award-line"
        icon-style="bg-green-500"
        box-style="stats-card-compact"
        description=""
      />
    </div>

    <!-- Search Area -->
    <ArtSearchBar
      ref="searchBarRef"
      v-model="searchFormState"
      :items="searchItems"
      :rules="searchRules"
      :is-expand="false"
      :show-expand="false"
      :show-reset-button="true"
      :show-search-button="true"
      :disabled-search-button="false"
      label-width="90px"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card mt-4">
      <template #header>
        <div class="flex-cb">
          <div class="title-group">
            <h4 class="m-0 font-semibold text-lg">{{ t('holiday.tableTitle') }}</h4>
            <p class="text-gray-400 text-xs mt-1">{{ t('holiday.tableSubtitle') }}</p>
          </div>
          <div class="flex gap-2">
            <ElTag type="success">
              {{ t('calendarArrange.dataCount', { count: pagination.total }) }}
            </ElTag>
          </div>
        </div>
      </template>

      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadHolidays">
        <template #left>
          <ElSpace wrap>
            <ElButton type="primary" @click="openAddDialog" v-ripple>
              <i class="fas fa-plus mr-1"></i> {{ t('holiday.addBtn') }}
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- Table of Holidays -->
      <ArtTable
        v-loading="loading"
        :data="holidays"
        :columns="columns"
        :pagination="pagination"
        empty-height="360px"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
        <!-- Custom slot: type -->
        <template #type="{ row }">
          <ElTag :type="getTypeTag(row.type)">
            {{ getTypeText(row.type) }}
          </ElTag>
        </template>

        <!-- Custom slot: rule -->
        <template #rule="{ row }">
          <span v-if="row.type === 'solar' || row.type === 'industry'">
            {{ t('holiday.ruleEveryYear', { month: row.month, day: row.day }) }}
          </span>
          <span v-else-if="row.type === 'weekday'">
            {{
              t('holiday.ruleWeekday', {
                month: row.month,
                weekNumber: row.week_number,
                dayOfWeek: getDayOfWeekText(row.day_of_week)
              })
            }}
          </span>
          <span v-else> {{ t('holiday.ruleCustom') }} </span>
        </template>

        <!-- Custom slot: regions -->
        <template #regions="{ row }">
          <div class="flex flex-wrap gap-1">
            <template v-if="row.regions">
              <ElTag
                v-for="r in row.regions
                  .split(',')
                  .map((s: string) => s.trim())
                  .filter(Boolean)"
                :key="r"
                size="small"
                type="info"
              >
                {{ getRegionLabel(r) }}
              </ElTag>
            </template>
            <ElTag v-else size="small" type="info">{{ t('holiday.allRegions') }}</ElTag>
          </div>
        </template>

        <!-- Custom slot: operation -->
        <template #operation="{ row }">
          <ArtButtonTable type="edit" @click="openEditDialog(row)" />
          <ArtButtonTable type="delete" @click="deleteHoliday(row)" />
        </template>
      </ArtTable>
    </ElCard>

    <!-- Dialog: Add / Edit Holiday -->
    <ElDialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? t('holiday.dialogAddTitle') : t('holiday.dialogEditTitle')"
      width="550px"
    >
      <ElForm :model="form" label-width="120px" :rules="formRules" ref="formRef">
        <ElFormItem :label="t('holiday.formName')" prop="name">
          <ElInput v-model="form.name" :placeholder="t('holiday.formNamePlaceholder')" />
        </ElFormItem>
        <ElFormItem :label="t('holiday.formType')" prop="type">
          <ElSelect
            v-model="form.type"
            :placeholder="t('holiday.formTypePlaceholder')"
            style="width: 100%"
          >
            <ElOption :label="t('holiday.formTypeSolar')" value="solar" />
            <ElOption :label="t('holiday.formTypeWeekday')" value="weekday" />
            <ElOption :label="t('holiday.formTypeIndustry')" value="industry" />
          </ElSelect>
        </ElFormItem>

        <!-- Conditional fields: solar type -->
        <template v-if="form.type === 'solar' || form.type === 'industry'">
          <ElFormItem :label="t('holiday.formFixedMonth')" prop="month">
            <ElSelect
              v-model="form.month"
              :placeholder="t('holiday.formMonthPlaceholder')"
              style="width: 100%"
            >
              <ElOption
                v-for="m in 12"
                :key="m"
                :label="t('holiday.formMonthLabel', { m })"
                :value="m"
              />
            </ElSelect>
          </ElFormItem>
          <ElFormItem :label="t('holiday.formFixedDay')" prop="day">
            <ElSelect
              v-model="form.day"
              :placeholder="t('holiday.formDayPlaceholder')"
              style="width: 100%"
            >
              <ElOption
                v-for="d in 31"
                :key="d"
                :label="t('holiday.formDayLabel', { d })"
                :value="d"
              />
            </ElSelect>
          </ElFormItem>
        </template>

        <!-- Conditional fields: weekday type -->
        <template v-if="form.type === 'weekday'">
          <ElFormItem :label="t('holiday.formSpecifyMonth')" prop="month">
            <ElSelect
              v-model="form.month"
              :placeholder="t('holiday.formMonthPlaceholder')"
              style="width: 100%"
            >
              <ElOption
                v-for="m in 12"
                :key="m"
                :label="t('holiday.formMonthLabel', { m })"
                :value="m"
              />
            </ElSelect>
          </ElFormItem>
          <ElFormItem :label="t('holiday.formWeekNumber')" prop="week_number">
            <ElSelect
              v-model="form.week_number"
              :placeholder="t('holiday.formWeekNumberPlaceholder')"
              style="width: 100%"
            >
              <ElOption :label="t('holiday.formWeekFirst')" :value="1" />
              <ElOption :label="t('holiday.formWeekSecond')" :value="2" />
              <ElOption :label="t('holiday.formWeekThird')" :value="3" />
              <ElOption :label="t('holiday.formWeekFourth')" :value="4" />
              <ElOption :label="t('holiday.formWeekLast')" :value="5" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem :label="t('holiday.formDayOfWeek')" prop="day_of_week">
            <ElSelect
              v-model="form.day_of_week"
              :placeholder="t('holiday.formDayOfWeekPlaceholder')"
              style="width: 100%"
            >
              <ElOption :label="t('holiday.dayMon')" :value="1" />
              <ElOption :label="t('holiday.dayTue')" :value="2" />
              <ElOption :label="t('holiday.dayWed')" :value="3" />
              <ElOption :label="t('holiday.dayThu')" :value="4" />
              <ElOption :label="t('holiday.dayFri')" :value="5" />
              <ElOption :label="t('holiday.daySat')" :value="6" />
              <ElOption :label="t('holiday.daySun')" :value="7" />
            </ElSelect>
          </ElFormItem>
        </template>

        <ElFormItem :label="t('holiday.formRegions')" prop="regions">
          <ElSelect
            v-model="dialogRegions"
            multiple
            clearable
            :placeholder="t('holiday.formRegionsPlaceholder')"
            style="width: 100%"
          >
            <ElOption :label="t('holiday.regionCn')" value="cn" />
            <ElOption :label="t('holiday.regionHk')" value="hk" />
            <ElOption :label="t('holiday.regionTw')" value="tw" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('holiday.formDescription')" prop="description">
          <ElInput
            v-model="form.description"
            type="textarea"
            :placeholder="t('holiday.formDescriptionPlaceholder')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <span class="dialog-footer">
          <ElButton @click="dialogVisible = false">{{ t('holiday.cancel') }}</ElButton>
          <ElButton type="primary" @click="submitForm">{{ t('holiday.confirm') }}</ElButton>
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
    fetchGetHolidayList,
    fetchAddHoliday,
    fetchUpdateHoliday,
    fetchDeleteHoliday
  } from '@/api/calendar'

  defineOptions({ name: 'Holiday' })

  const { t } = useI18n()

  const regionKeyMap: Record<string, string> = {
    cn: 'holiday.regionCn',
    hk: 'holiday.regionHk',
    tw: 'holiday.regionTw'
  }

  const getRegionLabel = (key: string) => {
    const i18nKey = regionKeyMap[key.toLowerCase()]
    return i18nKey ? t(i18nKey) : key
  }

  const loading = ref(false)
  const holidays = ref<any[]>([])

  // Pagination state
  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  // Statistics state
  const stats = reactive({
    totalCount: 0,
    solarCount: 0,
    weekdayCount: 0,
    industryCount: 0
  })

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  let currentId = 0

  const form = reactive({
    name: '',
    type: 'solar',
    month: 1,
    day: 1,
    week_number: 1,
    day_of_week: 1,
    regions: 'cn',
    description: ''
  })

  const formRules = computed(() => ({
    name: [{ required: true, message: t('holiday.ruleName'), trigger: 'blur' }],
    type: [{ required: true, message: t('holiday.ruleType'), trigger: 'change' }]
  }))

  const formRef = ref()

  const dialogRegions = computed({
    get() {
      return form.regions
        ? form.regions
            .split(',')
            .map((s) => s.trim())
            .filter(Boolean)
        : []
    },
    set(val: string[]) {
      form.regions = val.join(',')
    }
  })

  // Search Area configuration
  const searchBarRef = ref()
  const searchFormState = reactive({
    name: '',
    type: 'all',
    regions: ''
  })

  const searchItems = computed(() => [
    {
      label: t('holiday.searchName'),
      key: 'name',
      type: 'input',
      props: {
        placeholder: t('holiday.searchNamePlaceholder')
      }
    },
    {
      label: t('holiday.searchType'),
      key: 'type',
      type: 'select',
      props: {
        placeholder: t('holiday.searchTypePlaceholder'),
        options: [
          { label: t('holiday.searchTypeAll'), value: 'all' },
          { label: t('holiday.searchTypeSolar'), value: 'solar' },
          { label: t('holiday.searchTypeWeekday'), value: 'weekday' },
          { label: t('holiday.searchTypeIndustry'), value: 'industry' }
        ]
      }
    },
    {
      label: t('holiday.searchRegions'),
      key: 'regions',
      type: 'select',
      props: {
        placeholder: t('holiday.searchRegionsPlaceholder'),
        clearable: true,
        options: [
          { label: t('holiday.regionCn'), value: 'cn' },
          { label: t('holiday.regionHk'), value: 'hk' },
          { label: t('holiday.regionTw'), value: 'tw' }
        ]
      }
    }
  ])

  const searchRules = reactive({})

  // Columns definition
  const { columns, columnChecks } = useTableColumns(() => [
    { label: t('holiday.colId'), prop: 'id', width: '80' },
    { label: t('holiday.colName'), prop: 'name', minWidth: '150' },
    {
      label: t('holiday.colType'),
      prop: 'type',
      width: 120,
      useSlot: true,
      slotName: 'type'
    },
    {
      label: t('holiday.colRule'),
      prop: 'rule',
      minWidth: 180,
      useSlot: true,
      slotName: 'rule'
    },
    {
      label: t('holiday.colRegions'),
      prop: 'regions',
      width: 120,
      useSlot: true,
      slotName: 'regions'
    },
    { label: t('holiday.colDescription'), prop: 'description', minWidth: 180 },
    {
      label: t('holiday.colOperation'),
      prop: 'operation',
      width: 120,
      fixed: 'right',
      useSlot: true,
      slotName: 'operation'
    }
  ])

  const handleSearch = () => {
    pagination.current = 1
    loadHolidays()
  }

  const handleReset = () => {
    searchFormState.name = ''
    searchFormState.type = 'all'
    searchFormState.regions = ''
    pagination.current = 1
    loadHolidays()
  }

  const handleSizeChange = (val: number) => {
    pagination.size = val
    pagination.current = 1
    loadHolidays()
  }

  const handleCurrentChange = (val: number) => {
    pagination.current = val
    loadHolidays()
  }

  onMounted(() => {
    loadHolidays()
  })

  const loadHolidays = async () => {
    loading.value = true
    try {
      const params: any = {
        current: pagination.current,
        size: pagination.size
      }
      if (searchFormState.name) {
        params.name = searchFormState.name
      }
      if (searchFormState.type !== 'all') {
        params.type = searchFormState.type
      }
      if (searchFormState.regions) {
        params.regions = searchFormState.regions
      }
      const res = await fetchGetHolidayList(params)
      holidays.value = res.list || []
      pagination.total = res.total || 0

      // Update stats
      stats.totalCount = res.totalCount || 0
      stats.solarCount = res.solarCount || 0
      stats.weekdayCount = res.weekdayCount || 0
      stats.industryCount = res.industryCount || 0
    } catch (e: any) {
      ElMessage.error(e.message || t('holiday.errorLoad'))
    } finally {
      loading.value = false
    }
  }

  const getTypeText = (type: string) => {
    switch (type) {
      case 'solar':
        return t('holiday.typeSolar')
      case 'weekday':
        return t('holiday.typeWeekday')
      case 'industry':
        return t('holiday.typeIndustry')
      default:
        return t('holiday.typeUnknown')
    }
  }

  const getTypeTag = (type: string) => {
    switch (type) {
      case 'solar':
        return 'danger'
      case 'weekday':
        return 'primary'
      case 'industry':
        return 'success'
      default:
        return 'warning'
    }
  }

  const getDayOfWeekText = (val: number) => {
    return t(`holiday.dayOfWeekShort${val}`)
  }

  const openAddDialog = () => {
    dialogType.value = 'add'
    form.name = ''
    form.type = 'solar'
    form.month = 1
    form.day = 1
    form.week_number = 1
    form.day_of_week = 1
    form.regions = 'cn'
    form.description = ''
    dialogVisible.value = true
  }

  const openEditDialog = (row: any) => {
    dialogType.value = 'edit'
    currentId = row.id
    form.name = row.name
    form.type = row.type
    form.month = row.month || 1
    form.day = row.day || 1
    form.week_number = row.week_number || 1
    form.day_of_week = row.day_of_week || 1
    form.regions = row.regions
    form.description = row.description
    dialogVisible.value = true
  }

  const submitForm = async () => {
    formRef.value?.validate(async (valid: boolean) => {
      if (!valid) return
      try {
        if (dialogType.value === 'add') {
          await fetchAddHoliday(form)
          ElMessage.success(t('holiday.successAdd'))
        } else {
          await fetchUpdateHoliday({ id: currentId, ...form })
          ElMessage.success(t('holiday.successUpdate'))
        }
        dialogVisible.value = false
        loadHolidays()
      } catch (e: any) {
        ElMessage.error(e.message || t('holiday.errorSubmit'))
      }
    })
  }

  const deleteHoliday = (row: any) => {
    ElMessageBox.confirm(t('holiday.deleteConfirm', { name: row.name }), t('holiday.deleteTitle'), {
      confirmButtonText: t('holiday.confirm'),
      cancelButtonText: t('holiday.cancel'),
      type: 'warning'
    }).then(async () => {
      try {
        await fetchDeleteHoliday({ id: row.id })
        ElMessage.success(t('holiday.successDelete'))
        loadHolidays()
      } catch (e: any) {
        ElMessage.error(e.message || t('holiday.errorDelete'))
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
