import { DateRangePicker, type DateRange } from "@yca-software/design-system";
import { useTranslationNamespace } from "@/helpers";

interface DateRangeFilterProps {
  label: string;
  value: DateRange | undefined;
  onChange: (value: DateRange | undefined) => void;
  onApply: (value: DateRange | undefined) => void;
  minDate?: Date;
  maxDate?: Date;
}

export function DateRangeFilter({
  label,
  value,
  onChange,
  onApply,
  minDate,
  maxDate,
}: DateRangeFilterProps) {
  const { t } = useTranslationNamespace(["common"]);

  return (
    <div className="flex max-w-sm flex-col gap-1">
      <label className="text-sm font-medium">{label}</label>
      <DateRangePicker
        value={value}
        onChange={onChange}
        minDate={minDate}
        maxDate={maxDate}
        translations={{
          applyButton: t("common:dateRange.apply"),
          cancelButton: t("common:dateRange.cancel"),
          startLabel: t("common:dateRange.startDate"),
          endLabel: t("common:dateRange.endDate"),
          placeholder: t("common:dateRange.placeholder"),
          ariaLabel: t("common:dateRange.ariaLabel"),
          presetLabels: {
            today: t("common:dateRange.presets.today"),
            yesterday: t("common:dateRange.presets.yesterday"),
            last7: t("common:dateRange.presets.last7"),
            last14: t("common:dateRange.presets.last14"),
            last30: t("common:dateRange.presets.last30"),
            thisWeek: t("common:dateRange.presets.thisWeek"),
            lastWeek: t("common:dateRange.presets.lastWeek"),
            thisMonth: t("common:dateRange.presets.thisMonth"),
            lastMonth: t("common:dateRange.presets.lastMonth"),
          },
        }}
        onApply={onApply}
      />
    </div>
  );
}
