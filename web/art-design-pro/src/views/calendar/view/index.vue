<template>
  <div class="calendar-view-page flex flex-col gap-4 pb-6">
    <!-- Top Filter & Control Card -->
    <ElCard shadow="never" class="art-card-control">
      <div class="flex flex-col lg:flex-row items-start lg:items-center justify-between gap-4">
        <!-- Title & Subtitle -->
        <div class="title-group">
          <div class="flex items-center gap-3">
            <div
              class="w-9 h-9 rounded-xl bg-gradient-to-br from-primary/20 to-primary/5 flex items-center justify-center text-primary text-xl shadow-xs"
            >
              <ArtSvgIcon icon="ri:calendar-todo-line" />
            </div>
            <div>
              <h3
                class="m-0 text-lg font-bold text-gray-800 dark:text-gray-100 flex items-center gap-2"
              >
                {{ t('calendarView.title') }}
                <span
                  class="text-xs px-2 py-0.5 rounded-full bg-rose-50 dark:bg-rose-950/30 text-rose-600 dark:text-rose-400 border border-rose-200 dark:border-rose-900/40 font-normal"
                >
                  {{ t('calendarView.monToSun') }}
                </span>
              </h3>
              <p class="m-0 text-xs text-gray-400 mt-0.5">
                {{ t('calendarView.subtitle') }}
              </p>
            </div>
          </div>
        </div>

        <!-- Filters & Navigation Buttons -->
        <div class="flex flex-wrap items-center gap-3 w-full lg:w-auto">
          <!-- Region Select -->
          <ElSelect
            v-model="currentRegion"
            size="default"
            class="w-32"
            @change="handleRegionChange"
          >
            <ElOption :label="t('calendarView.cn')" value="cn" />
            <ElOption :label="t('calendarView.hk')" value="hk" />
            <ElOption :label="t('calendarView.tw')" value="tw" />
            <ElOption :label="t('calendarView.allRegions')" value="all" />
          </ElSelect>

          <!-- Year Picker -->
          <ElDatePicker
            v-model="yearDate"
            type="year"
            :clearable="false"
            class="w-28"
            :format="locale === 'en' ? 'YYYY' : 'YYYY年'"
            value-format="YYYY"
            @change="handleYearChange"
          />

          <!-- Quick Navigation Buttons -->
          <ElButtonGroup>
            <ElButton :icon="ArrowLeft" @click="handlePrevMonth">
              {{ t('calendarView.prevMonth') }}
            </ElButton>
            <ElButton @click="handleToday">
              <ArtSvgIcon icon="ri:focus-3-line" class="mr-1" />
              {{ t('calendarView.today') }}
            </ElButton>
            <ElButton @click="handleNextMonth">
              {{ t('calendarView.nextMonth') }}
              <ElIcon class="el-icon--right"><ArrowRight /></ElIcon>
            </ElButton>
          </ElButtonGroup>

          <ElButton :icon="Refresh" circle @click="loadMonthData" />
        </div>
      </div>

      <!-- Month Tabs Bar -->
      <div
        class="month-tabs-bar mt-4 pt-3 border-t border-gray-100 dark:border-gray-800 flex flex-wrap items-center gap-1.5 select-none"
      >
        <span class="text-xs text-gray-400 font-medium mr-2 flex items-center shrink-0">
          <ArtSvgIcon icon="ri:calendar-event-line" class="mr-1 text-sm" />
          {{ formatYearLabel(currentYear) }}
        </span>
        <button
          v-for="m in 12"
          :key="m"
          type="button"
          class="month-chip px-3 py-1.5 text-xs rounded-lg transition-all shrink-0 cursor-pointer font-medium"
          :class="
            currentMonth === m
              ? 'bg-primary text-white shadow-sm shadow-primary/30 font-bold scale-105'
              : 'bg-gray-100/80 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700'
          "
          @click="selectMonth(m)"
        >
          {{ formatMonthTab(m) }}
        </button>
      </div>
    </ElCard>

    <!-- Month Statistics Row -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <ArtStatsCard
        :title="t('calendarView.statsTotalDays')"
        :count="monthStats.totalDays"
        icon="ri:calendar-line"
        icon-style="bg-blue-500"
        box-style="stats-card-compact"
        description=""
      />
      <ArtStatsCard
        :title="t('calendarView.statsWorkdays')"
        :count="monthStats.workdays"
        icon="ri:briefcase-line"
        icon-style="bg-emerald-500"
        box-style="stats-card-compact"
        description=""
      />
      <ArtStatsCard
        :title="t('calendarView.statsRestDays')"
        :count="monthStats.restDays"
        icon="ri:cup-line"
        icon-style="bg-rose-500"
        box-style="stats-card-compact"
        description=""
      />
      <ArtStatsCard
        :title="t('calendarView.statsExceptions')"
        :count="monthStats.exceptionCount"
        icon="ri:alarm-warning-line"
        icon-style="bg-amber-500"
        box-style="stats-card-compact"
        description=""
      />
    </div>

    <!-- Main Layout: Left Calendar Grid + Right Day & Month Details Panel -->
    <div class="grid grid-cols-1 xl:grid-cols-12 gap-4">
      <!-- Calendar Area (8 cols on XL) -->
      <ElCard shadow="never" class="xl:col-span-8 art-calendar-card flex flex-col">
        <!-- Calendar Legend & Header Bar -->
        <div
          class="flex flex-wrap items-center justify-between gap-3 mb-4 pb-3 border-b border-gray-100 dark:border-gray-800"
        >
          <div class="flex items-center gap-2">
            <span class="text-base font-extrabold text-gray-800 dark:text-gray-100">
              {{ formatYearMonth(currentYear, currentMonth) }}
            </span>
            <ElTag size="small" type="primary" effect="plain" class="rounded-full font-medium">
              {{ formatRegionLabel(currentRegion) }}
            </ElTag>
          </div>

          <!-- Color Legend -->
          <div class="flex flex-wrap items-center gap-3.5 text-xs">
            <div class="flex items-center gap-1.5">
              <span
                class="w-3 h-3 rounded bg-rose-100 border border-rose-300 dark:bg-rose-950/60 dark:border-rose-800 inline-block"
              ></span>
              <span class="text-rose-600 dark:text-rose-400 font-semibold">{{
                t('calendarView.legendRest')
              }}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <span
                class="w-3 h-3 rounded bg-emerald-100 border border-emerald-300 dark:bg-emerald-950/60 dark:border-emerald-800 inline-block"
              ></span>
              <span class="text-gray-900 dark:text-gray-100 font-semibold">{{
                t('calendarView.legendWork')
              }}</span>
            </div>
            <div class="flex items-center gap-1.5">
              <span
                class="w-3 h-3 rounded bg-blue-100 border border-blue-300 dark:bg-blue-950/60 dark:border-blue-800 inline-block"
              ></span>
              <span class="text-blue-600 dark:text-blue-400 font-medium">{{
                t('calendarView.legendFestival')
              }}</span>
            </div>
          </div>
        </div>

        <!-- ElCalendar Core -->
        <div v-loading="loading" class="calendar-wrapper">
          <ElCalendar v-model="calendarDate">
            <template #header>
              <!-- Hidden since we have our custom top header -->
              <span class="hidden"></span>
            </template>

            <template #date-cell="{ data }">
              <div
                class="calendar-day-cell relative w-full h-full flex flex-col justify-between p-2 rounded-xl transition-all select-none"
                :class="getDayCellClass(data)"
                @click="handleDayClick(data.day)"
              >
                <!-- Top Header: Day Number & Status Badges -->
                <div class="flex items-start justify-between w-full">
                  <!-- Date Number & Today Tag -->
                  <div class="flex items-center gap-1">
                    <div
                      class="day-number text-xs md:text-sm flex items-center justify-center rounded-lg transition-transform"
                      :class="getDayNumberClass(data)"
                    >
                      {{ getDayOfMonth(data.day) }}
                    </div>
                    <span
                      v-if="isDayToday(data.day)"
                      class="today-pill px-1.5 py-0.2 rounded text-[10px] font-extrabold text-white tracking-wide"
                    >
                      {{ t('calendarView.tagToday') }}
                    </span>
                  </div>

                  <!-- Right Exception/Weekend Badge (休 / 班 / 周末) -->
                  <div v-if="getDayBadge(data.day)" class="badge-tag">
                    <span
                      class="text-[10px] px-1.5 py-0.5 rounded-md font-bold uppercase tracking-tight shadow-xs"
                      :class="getDayBadge(data.day)!.class"
                    >
                      {{ getDayBadge(data.day)!.text }}
                    </span>
                  </div>
                </div>

                <!-- Bottom Tags: Holidays & Exception Reason -->
                <div class="cell-content mt-1.5 flex flex-col gap-1 w-full overflow-hidden">
                  <!-- Festivals / Holidays from `holiday` table -->
                  <div
                    v-for="(hName, idx) in getDayHolidays(data.day)"
                    :key="'h-' + idx"
                    class="truncate"
                  >
                    <ElTooltip
                      :content="getDayHolidayDesc(data.day, idx)"
                      placement="top"
                      :show-after="300"
                    >
                      <span
                        class="holiday-pill inline-flex items-center gap-1 max-w-full truncate px-1.5 py-0.5 rounded-md text-[10px] font-semibold border shadow-xs"
                        style="
                          background: rgba(239, 246, 255, 0.9);
                          color: #1d4ed8;
                          border-color: rgba(191, 219, 254, 0.8);
                        "
                      >
                        <ArtSvgIcon icon="ri:bookmark-3-line" class="shrink-0 text-[10px]" />
                        <span class="truncate">{{ hName }}</span>
                      </span>
                    </ElTooltip>
                  </div>

                  <!-- Exception Description from `calendar_exception` table -->
                  <div v-if="getDayExceptionDesc(data.day)" class="truncate">
                    <ElTooltip
                      :content="getDayExceptionDesc(data.day)"
                      placement="top"
                      :show-after="300"
                    >
                      <span
                        class="exception-pill inline-flex items-center gap-1 max-w-full truncate px-1.5 py-0.5 rounded-md text-[10px] font-semibold border shadow-xs"
                        :class="
                          isDayWorkday(data.day)
                            ? 'bg-emerald-100/80 dark:bg-emerald-950/60 text-emerald-800 dark:text-emerald-300 border-emerald-300/80 dark:border-emerald-800'
                            : 'bg-rose-100/90 dark:rose-950/60 text-rose-800 dark:text-rose-300 border-rose-300/80 dark:border-rose-800'
                        "
                      >
                        <ArtSvgIcon
                          :icon="isDayWorkday(data.day) ? 'ri:briefcase-line' : 'ri:umbrella-line'"
                          class="shrink-0 text-[10px]"
                        />
                        <span class="truncate">{{ getDayExceptionDesc(data.day) }}</span>
                      </span>
                    </ElTooltip>
                  </div>
                </div>
              </div>
            </template>
          </ElCalendar>
        </div>
      </ElCard>

      <!-- Right Detail Area (4 cols on XL) -->
      <div class="xl:col-span-4 flex flex-col gap-4">
        <!-- Selected Date Details Card -->
        <ElCard shadow="never" class="art-selected-day-card">
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div
                  class="w-6 h-6 rounded-md bg-primary/10 flex items-center justify-center text-primary text-sm"
                >
                  <ArtSvgIcon icon="ri:information-line" />
                </div>
                <span class="font-bold text-sm text-gray-800 dark:text-gray-100">
                  {{ t('calendarView.dayDetailTitle') }}
                </span>
              </div>
              <ElTag
                v-if="selectedDayInfo"
                :type="selectedDayInfo.is_workday ? 'success' : 'danger'"
                size="small"
                class="font-medium"
              >
                {{
                  selectedDayInfo.is_workday
                    ? t('calendarView.statusWorkday')
                    : t('calendarView.statusWeekend')
                }}
              </ElTag>
            </div>
          </template>

          <div v-if="selectedDayInfo" class="flex flex-col gap-3 text-sm">
            <!-- Big Date & Weekday Display -->
            <div
              class="p-3.5 rounded-xl border flex items-center justify-between transition-colors shadow-xs"
              :class="
                selectedDayInfo.is_workday
                  ? 'bg-emerald-50/40 dark:bg-emerald-950/20 border-emerald-200/70 dark:border-emerald-900/40'
                  : 'bg-rose-50/60 dark:bg-rose-950/20 border-rose-200/70 dark:border-rose-900/40'
              "
            >
              <div>
                <div class="text-xl font-extrabold text-gray-900 dark:text-gray-100">
                  {{ selectedDateStr }}
                </div>
                <div
                  class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 font-medium flex items-center gap-1.5"
                >
                  <span
                    :class="
                      isWeekendDay(selectedDateStr) && !isCompensatoryWorkday(selectedDateStr)
                        ? 'text-rose-500 font-bold'
                        : 'text-gray-600 dark:text-gray-300'
                    "
                  >
                    {{ getDayWeekdayText(selectedDateStr) }}
                  </span>
                  <span
                    v-if="isWeekendDay(selectedDateStr) && isCompensatoryWorkday(selectedDateStr)"
                    class="text-[11px] px-1.5 py-0.2 rounded bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-300 font-bold"
                  >
                    {{ t('calendarView.statusCompensatoryWeekend') }}
                  </span>
                  <span
                    v-if="isDayToday(selectedDateStr)"
                    class="text-[11px] px-1.5 py-0.2 rounded bg-primary/10 text-primary font-bold border border-primary/25"
                  >
                    {{ t('calendarView.tagToday') }}
                  </span>
                </div>
              </div>
              <div
                class="w-12 h-12 rounded-xl flex items-center justify-center text-2xl font-black shadow-sm"
                :class="
                  selectedDayInfo.is_workday
                    ? 'bg-emerald-500 text-white'
                    : 'bg-rose-500 text-white'
                "
              >
                {{
                  selectedDayInfo.is_workday ? t('calendarView.tagWork') : t('calendarView.tagRest')
                }}
              </div>
            </div>

            <!-- Attributes Breakdown -->
            <div class="space-y-2.5 pt-1">
              <!-- Workday Status -->
              <div
                class="flex items-start justify-between py-1.5 border-b border-gray-100 dark:border-gray-800"
              >
                <span class="text-gray-400 text-xs">{{ t('calendarView.workStatusLabel') }}</span>
                <span
                  class="font-semibold text-xs"
                  :class="
                    selectedDayInfo.is_workday
                      ? 'text-emerald-600 dark:text-emerald-400'
                      : 'text-rose-600 dark:text-rose-400'
                  "
                >
                  {{ formatWorkdayStatus(selectedDayInfo) }}
                </span>
              </div>

              <!-- Exception Reason (from calendar_exception table) -->
              <div
                v-if="selectedDayInfo.is_exception"
                class="flex flex-col gap-1 py-1.5 border-b border-gray-100 dark:border-gray-800"
              >
                <span class="text-gray-400 text-xs flex items-center gap-1 font-medium">
                  <ArtSvgIcon icon="ri:file-text-line" />
                  {{ t('calendarView.exceptionReason') }}
                </span>
                <div
                  class="p-2.5 rounded-lg text-xs border font-medium"
                  :class="
                    selectedDayInfo.is_workday
                      ? 'bg-emerald-50 dark:bg-emerald-950/30 text-emerald-800 dark:text-emerald-300 border-emerald-200/80 dark:border-emerald-800'
                      : 'bg-rose-50 dark:bg-rose-950/30 text-rose-800 dark:text-rose-300 border-rose-200/80 dark:border-rose-800'
                  "
                >
                  {{ selectedDayInfo.exception_desc || t('calendarView.exceptionDefaultDesc') }}
                </div>
              </div>

              <!-- Associated Festivals (from holiday table) -->
              <div
                v-if="selectedDayInfo.holidays && selectedDayInfo.holidays.length > 0"
                class="flex flex-col gap-1.5 py-1.5 border-b border-gray-100 dark:border-gray-800"
              >
                <span class="text-gray-400 text-xs flex items-center gap-1 font-medium">
                  <ArtSvgIcon icon="ri:gift-line" />
                  {{ t('calendarView.festivals') }}
                </span>
                <div class="flex flex-col gap-1.5">
                  <div
                    v-for="(h, idx) in selectedDayInfo.holidays"
                    :key="idx"
                    class="p-2.5 rounded-lg bg-blue-50/70 dark:bg-blue-950/30 border border-blue-100 dark:border-blue-800/50 text-xs"
                  >
                    <div class="font-bold text-blue-700 dark:text-blue-300">
                      {{ h }}
                    </div>
                    <div
                      v-if="selectedDayInfo.holiday_descs && selectedDayInfo.holiday_descs[idx]"
                      class="text-[11px] text-gray-500 dark:text-gray-400 mt-0.5"
                    >
                      {{ selectedDayInfo.holiday_descs[idx] }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Quick Action Button -->
            <div class="pt-2">
              <ElButton type="primary" plain class="w-full font-medium" @click="goToArrangePage">
                <ArtSvgIcon icon="ri:settings-line" class="mr-1" />
                {{ t('calendarView.manageArrangeBtn') }}
              </ElButton>
            </div>
          </div>

          <div v-else class="text-center py-10 text-gray-400 text-xs">
            <ArtSvgIcon
              icon="ri:calendar-2-line"
              class="text-4xl mb-2 text-gray-300 dark:text-gray-600 block mx-auto"
            />
            {{ t('calendarView.noSelection') }}
          </div>
        </ElCard>

        <!-- Monthly Holidays & Exceptions List -->
        <ElCard shadow="never" class="art-month-list-card flex-1">
          <template #header>
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <div
                  class="w-6 h-6 rounded-md bg-amber-500/10 flex items-center justify-center text-amber-500 text-sm"
                >
                  <ArtSvgIcon icon="ri:list-check-2" />
                </div>
                <span class="font-bold text-sm text-gray-800 dark:text-gray-100">
                  {{ t('calendarView.monthHolidaysList') }}
                </span>
              </div>
              <ElTag size="small" type="info" class="font-medium">
                {{ monthExceptionsList.length }} {{ t('calendarView.itemSuffix') }}
              </ElTag>
            </div>
          </template>

          <div
            v-if="monthExceptionsList.length > 0"
            class="flex flex-col gap-2 max-h-[320px] overflow-y-auto pr-1"
          >
            <div
              v-for="item in monthExceptionsList"
              :key="item.date"
              class="p-2.5 rounded-lg border transition-all cursor-pointer flex items-center justify-between text-xs hover:shadow-xs"
              :class="
                item.is_workday
                  ? 'bg-emerald-50/40 dark:bg-emerald-950/20 border-emerald-100 dark:border-emerald-900/30 hover:border-emerald-300'
                  : 'bg-rose-50/50 dark:bg-rose-950/20 border-rose-100 dark:border-rose-900/30 hover:border-rose-300'
              "
              @click="handleDayClick(item.date)"
            >
              <div class="flex items-center gap-2.5">
                <span
                  class="w-6 h-6 rounded-md flex items-center justify-center text-[11px] font-bold shrink-0 text-white shadow-xs"
                  :class="item.is_workday ? 'bg-emerald-500' : 'bg-rose-500'"
                >
                  {{ item.is_workday ? t('calendarView.tagWork') : t('calendarView.tagRest') }}
                </span>
                <div>
                  <div class="font-bold text-gray-800 dark:text-gray-200">
                    {{ item.date }}
                    <span class="text-[11px] font-normal text-gray-400 ml-1">
                      {{ getDayWeekdayText(item.date) }}
                    </span>
                  </div>
                  <div class="text-gray-500 dark:text-gray-400 text-[11px] mt-0.5">
                    {{
                      item.description ||
                      (item.is_workday
                        ? t('calendarView.statusCompensatoryWork')
                        : t('calendarView.statusHolidayRest'))
                    }}
                  </div>
                </div>
              </div>

              <ArtSvgIcon icon="ri:arrow-right-s-line" class="text-gray-400 text-base" />
            </div>
          </div>

          <div v-else class="text-center py-8 text-gray-400 text-xs">
            <ArtSvgIcon
              icon="ri:inbox-line"
              class="text-3xl mb-1 text-gray-300 dark:text-gray-600 block mx-auto"
            />
            {{ t('calendarView.noMonthHolidays') }}
          </div>
        </ElCard>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, watch, onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import { useI18n } from 'vue-i18n'
  import { ArrowLeft, ArrowRight, Refresh } from '@element-plus/icons-vue'
  import { fetchGetCalendarMonth, CalendarDayItem } from '@/api/calendar'

  // Configure dayjs for Monday as the first day of week
  import dayjs from 'dayjs'
  import 'dayjs/locale/zh-cn'
  import localeData from 'dayjs/plugin/localeData'
  import updateLocale from 'dayjs/plugin/updateLocale'

  dayjs.extend(localeData)
  dayjs.extend(updateLocale)
  dayjs.locale('zh-cn')
  dayjs.updateLocale('zh-cn', { weekStart: 1 })
  dayjs.updateLocale('en', { weekStart: 1 })

  const { t, locale } = useI18n()
  const router = useRouter()

  const enShortMonths = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec'
  ]
  const enFullMonths = [
    'January',
    'February',
    'March',
    'April',
    'May',
    'June',
    'July',
    'August',
    'September',
    'October',
    'November',
    'December'
  ]

  function formatMonthTab(m: number): string {
    if (locale.value === 'en') {
      return enShortMonths[m - 1]
    }
    return `${m}月`
  }

  function formatYearLabel(year: number): string {
    if (locale.value === 'en') {
      return `${year}:`
    }
    return `${year}年:`
  }

  function formatYearMonth(year: number, month: number): string {
    if (locale.value === 'en') {
      return `${enFullMonths[month - 1]} ${year}`
    }
    return `${year} 年 ${month} 月`
  }

  // Current date state
  const today = new Date()
  const calendarDate = ref<Date>(new Date())
  const currentYear = ref<number>(today.getFullYear())
  const currentMonth = ref<number>(today.getMonth() + 1)
  const currentRegion = ref<string>('cn')
  const yearDate = ref<string>(String(today.getFullYear()))

  // Selected day state
  const selectedDateStr = ref<string>('')

  // Data state
  const loading = ref<boolean>(false)
  const daysMap = ref<Record<string, CalendarDayItem>>({})
  const monthExceptionsList = ref<any[]>([])
  const monthStats = ref({
    totalDays: 0,
    workdays: 0,
    restDays: 0,
    exceptionCount: 0,
    holidayCount: 0
  })

  // Selected day info computed from daysMap
  const selectedDayInfo = computed(() => {
    if (!selectedDateStr.value) return null
    return daysMap.value[selectedDateStr.value] || null
  })

  // Format region label
  function formatRegionLabel(region: string) {
    switch (region) {
      case 'cn':
        return t('calendarView.cn')
      case 'hk':
        return t('calendarView.hk')
      case 'tw':
        return t('calendarView.tw')
      case 'all':
        return t('calendarView.allRegions')
      default:
        return region
    }
  }

  // Extract Day number from 'YYYY-MM-DD'
  function getDayOfMonth(dateStr: string) {
    if (!dateStr) return ''
    const parts = dateStr.split('-')
    return parseInt(parts[2], 10)
  }

  // Check if date is Saturday or Sunday
  function isWeekendDay(dateStr: string): boolean {
    if (!dateStr) return false
    const d = new Date(dateStr)
    const weekday = d.getDay()
    return weekday === 0 || weekday === 6
  }

  // Check if date is compensatory workday (被调休上班)
  function isCompensatoryWorkday(dateStr: string): boolean {
    const day = daysMap.value[dateStr]
    return !!(day && day.is_exception && day.is_workday)
  }

  // Check if day is workday overall
  function isDayWorkday(dateStr: string): boolean {
    const day = daysMap.value[dateStr]
    if (day) {
      return day.is_workday
    }
    return !isWeekendDay(dateStr)
  }

  // Badges: "休", "班"
  function getDayBadge(dateStr: string) {
    const day = daysMap.value[dateStr]

    // 1. If explicitly in exception table
    if (day && day.is_exception) {
      if (day.is_workday) {
        return { text: t('calendarView.tagWork'), class: 'bg-emerald-600 text-white shadow-xs' }
      } else {
        return { text: t('calendarView.tagRest'), class: 'bg-rose-500 text-white shadow-xs' }
      }
    }

    return null
  }

  // Day cell dynamic class: Background & borders
  function getDayCellClass(data: { day: string; type: string }) {
    const day = daysMap.value[data.day]
    const isSelected = selectedDateStr.value === data.day
    const isCurrentMonth = data.type === 'current-month'
    const isCompensatory = isCompensatoryWorkday(data.day)
    const isToday = isDayToday(data.day)

    const classes: string[] = []

    if (!isCurrentMonth) {
      classes.push('opacity-30 hover:opacity-75')
    }

    // Selected state
    if (isSelected) {
      classes.push('ring-2 ring-primary ring-offset-2 dark:ring-offset-gray-900 shadow-md')
    } else {
      classes.push('hover:shadow-md hover:-translate-y-0.5')
    }

    // Background & border rules:
    // 重点美化“今天”：显著的主题色边框、光感渐变背景与微光阴影
    if (isToday) {
      classes.push('calendar-day-cell--today')
      if (day && day.is_exception && !day.is_workday) {
        // 今天恰逢法定节假日放假：在淡红底色基础上叠加高亮品牌主色边框和柔和渐变
        classes.push(
          'bg-gradient-to-br from-rose-100/95 via-rose-50/80 to-primary/10 dark:from-rose-950/70 dark:via-rose-950/40 dark:to-primary/20'
        )
      } else if (isCompensatory) {
        // 今天恰逢调休补班：浅绿底叠加主色边框
        classes.push(
          'bg-gradient-to-br from-emerald-100/85 via-emerald-50/60 to-primary/10 dark:from-emerald-950/60 dark:via-emerald-950/30 dark:to-primary/20'
        )
      } else {
        // 今天普通工作日或双休：品牌主色清爽微蓝半透明渐变底色
        classes.push(
          'bg-gradient-to-br from-primary/15 via-primary/6 to-transparent dark:from-primary/25 dark:via-primary/10 dark:to-transparent'
        )
      }
    } else if (day && day.is_exception && !day.is_workday) {
      // 法定节假日放假安排改为淡红色背景
      classes.push(
        'bg-rose-50/90 dark:bg-rose-950/35 border border-rose-300/80 dark:border-rose-900/60 shadow-xs'
      )
    } else if (isCompensatory) {
      // 调休补班：微浅绿/灰底
      classes.push(
        'bg-emerald-50/30 dark:bg-emerald-950/15 border border-emerald-200/70 dark:border-emerald-900/40'
      )
    } else {
      // 普通工作日及正常双休日：不改变背景，保持标准背景
      classes.push('bg-white dark:bg-gray-800/50 border border-gray-100 dark:border-gray-800')
    }

    return classes.join(' ')
  }

  // Day number font class:
  // "同时周六与周日的字体为红色，如果当前周六或周日被调休则字体为黑色"
  function getDayNumberClass(data: { day: string }) {
    const isToday = isDayToday(data.day)
    const isWeekend = isWeekendDay(data.day)
    const isCompensatory = isCompensatoryWorkday(data.day)

    // Today highlight
    if (isToday) {
      return 'w-6 h-6 rounded-full bg-primary text-white shadow-sm font-black scale-105'
    }

    // Saturday or Sunday
    if (isWeekend) {
      if (isCompensatory) {
        // 当前周六或周日被调休 -> 字体为黑色
        return 'text-gray-900 dark:text-gray-100 font-bold'
      } else {
        // 周六与周日 -> 字体为红色
        return 'text-rose-600 dark:text-rose-400 font-bold'
      }
    }

    // Weekday statutory holiday: also subtle festive rose
    const day = daysMap.value[data.day]
    if (day && day.is_exception && !day.is_workday) {
      return 'text-rose-600 dark:text-rose-400 font-bold'
    }

    // Normal Monday - Friday workday
    return 'text-gray-700 dark:text-gray-200 font-semibold'
  }

  // Format Date object to 'YYYY-MM-DD'
  function formatDate(d: Date): string {
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${y}-${m}-${day}`
  }

  const todayStr = formatDate(today)

  function isDayToday(dateStr: string): boolean {
    return dateStr === todayStr
  }

  // Get day holidays
  function getDayHolidays(dateStr: string): string[] {
    const day = daysMap.value[dateStr]
    return day?.holidays || []
  }

  // Get day holiday description
  function getDayHolidayDesc(dateStr: string, idx: number): string {
    const day = daysMap.value[dateStr]
    return day?.holiday_descs?.[idx] || day?.holidays?.[idx] || ''
  }

  // Get day exception description
  function getDayExceptionDesc(dateStr: string): string {
    const day = daysMap.value[dateStr]
    return day?.exception_desc || ''
  }

  // Format weekday text
  function getDayWeekdayText(dateStr: string): string {
    const d = new Date(dateStr)
    const weekdayKeys = [
      'calendarView.weekdaySun',
      'calendarView.weekdayMon',
      'calendarView.weekdayTue',
      'calendarView.weekdayWed',
      'calendarView.weekdayThu',
      'calendarView.weekdayFri',
      'calendarView.weekdaySat'
    ]
    return t(weekdayKeys[d.getDay()])
  }

  // Format workday status description
  function formatWorkdayStatus(day: CalendarDayItem): string {
    if (day.is_exception) {
      return day.is_workday
        ? t('calendarView.statusCompensatoryWork')
        : t('calendarView.statusHolidayRest')
    }
    return day.is_workday ? t('calendarView.statusWorkday') : t('calendarView.statusWeekend')
  }

  // Load month data from backend API
  async function loadMonthData() {
    loading.value = true
    try {
      const res = await fetchGetCalendarMonth({
        year: currentYear.value,
        month: currentMonth.value,
        region: currentRegion.value
      })

      if (res && res.days) {
        daysMap.value = res.days
        monthExceptionsList.value = res.exceptions || []
        if (res.stats) {
          monthStats.value = res.stats
        }
      }
    } catch (err) {
      console.error('Failed to load month calendar data:', err)
    } finally {
      loading.value = false
    }
  }

  // Handle Day Click
  function handleDayClick(dayStr: string) {
    selectedDateStr.value = dayStr
  }

  // Handle Year Change from DatePicker
  function handleYearChange(newYear: string) {
    if (newYear) {
      currentYear.value = parseInt(newYear, 10)
      calendarDate.value = new Date(currentYear.value, currentMonth.value - 1, 1)
      loadMonthData()
    }
  }

  // Handle Month selection
  function selectMonth(m: number) {
    currentMonth.value = m
    calendarDate.value = new Date(currentYear.value, m - 1, 1)
    loadMonthData()
  }

  // Handle Region Change
  function handleRegionChange() {
    loadMonthData()
  }

  // Handle Previous Month
  function handlePrevMonth() {
    let y = currentYear.value
    let m = currentMonth.value - 1
    if (m < 1) {
      m = 12
      y -= 1
    }
    currentYear.value = y
    currentMonth.value = m
    yearDate.value = String(y)
    calendarDate.value = new Date(y, m - 1, 1)
    loadMonthData()
  }

  // Handle Next Month
  function handleNextMonth() {
    let y = currentYear.value
    let m = currentMonth.value + 1
    if (m > 12) {
      m = 1
      y += 1
    }
    currentYear.value = y
    currentMonth.value = m
    yearDate.value = String(y)
    calendarDate.value = new Date(y, m - 1, 1)
    loadMonthData()
  }

  // Handle Today Click
  function handleToday() {
    const now = new Date()
    currentYear.value = now.getFullYear()
    currentMonth.value = now.getMonth() + 1
    yearDate.value = String(now.getFullYear())
    calendarDate.value = now
    selectedDateStr.value = formatDate(now)
    loadMonthData()
  }

  // Watch calendarDate in case user clicks inside ElCalendar internal buttons
  watch(calendarDate, (newVal) => {
    if (newVal) {
      const y = newVal.getFullYear()
      const m = newVal.getMonth() + 1
      if (y !== currentYear.value || m !== currentMonth.value) {
        currentYear.value = y
        currentMonth.value = m
        yearDate.value = String(y)
        loadMonthData()
      }
    }
  })

  // Jump to Arrange Page
  function goToArrangePage() {
    router.push('/calendar/arrange')
  }

  onMounted(() => {
    selectedDateStr.value = formatDate(new Date())
    loadMonthData()
  })
</script>

<style scoped>
  .calendar-view-page {
    animation: fadeIn 0.3s ease-in-out;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .calendar-wrapper :deep(.el-calendar) {
    --el-calendar-border: 1px solid var(--el-border-color-lighter);
    background-color: transparent;
  }

  .calendar-wrapper :deep(.el-calendar__header) {
    display: none !important;
  }

  .calendar-wrapper :deep(.el-calendar__body) {
    padding: 0;
  }

  .calendar-wrapper :deep(.el-calendar-table) {
    border-collapse: separate;
    border-spacing: 5px;
  }

  /* Weekday Header: Mon-Fri default, Sat & Sun red */
  .calendar-wrapper :deep(.el-calendar-table thead th) {
    padding: 10px 0;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    font-weight: 600;
    text-align: center;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  /* Saturday (col 6) and Sunday (col 7) header: Red Font */
  .calendar-wrapper :deep(.el-calendar-table thead th:nth-child(6)),
  .calendar-wrapper :deep(.el-calendar-table thead th:nth-child(7)) {
    color: #f43f5e !important;
    font-weight: 700;
  }

  .calendar-wrapper :deep(.el-calendar-table td) {
    border: none;
    padding: 0;
    height: 106px;
    vertical-align: top;
    border-radius: 12px;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  }

  .calendar-wrapper :deep(.el-calendar-table td.is-selected) {
    background-color: transparent;
  }

  .calendar-wrapper :deep(.el-calendar-table .el-calendar-day) {
    height: 100%;
    padding: 0;
    box-sizing: border-box;
  }

  .calendar-wrapper :deep(.el-calendar-table .el-calendar-day:hover) {
    background-color: transparent;
  }

  .calendar-day-cell {
    min-height: 102px;
    box-sizing: border-box;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.03);
  }

  /* 今天日历块深度美化：高亮边框、光感渐变、左侧装饰条与微光阴影 */
  .calendar-day-cell--today {
    position: relative;
    border: 2px solid var(--el-color-primary) !important;
    box-shadow:
      0 0 0 1px rgba(var(--el-color-primary-rgb, 64, 158, 255), 0.15),
      0 6px 16px -2px rgba(var(--el-color-primary-rgb, 64, 158, 255), 0.25) !important;
  }

  .today-pill {
    background: linear-gradient(135deg, var(--el-color-primary) 0%, #3b82f6 100%);
    box-shadow: 0 1px 3px rgba(64, 158, 255, 0.4);
    animation: todayPulse 2.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }

  @keyframes todayPulse {
    0%,
    100% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.88;
      transform: scale(0.96);
    }
  }
</style>
