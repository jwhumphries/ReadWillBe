import { useRef, useEffect, useState } from 'react';
import { DayPicker } from 'react-day-picker';
import { format, startOfWeek, getISOWeek, getYear } from 'date-fns';
import { Calendar, ChevronLeft, ChevronRight } from 'lucide-react';

export type DateType = 'day' | 'week' | 'month';

interface DatePickerProps {
  value?: Date;
  dateType?: DateType;
  onChange: (date: Date | undefined, dateType: DateType) => void;
  placeholder?: string;
  name?: string;
  required?: boolean;
}

// Format the display value based on date type
function formatDisplayValue(date: Date, dateType: DateType): string {
  switch (dateType) {
    case 'week':
      return `Week ${getISOWeek(date)}, ${getYear(date)}`;
    case 'month':
      return format(date, 'MMMM yyyy');
    case 'day':
    default:
      return format(date, 'MMM d, yyyy');
  }
}

// Format the date for form submission (matches Go parseDate expectations)
function formatForSubmission(date: Date, dateType: DateType): string {
  switch (dateType) {
    case 'week':
      const week = getISOWeek(date);
      const year = getYear(date);
      return `${year}-W${week.toString().padStart(2, '0')}`;
    case 'month':
      return format(date, 'MMMM yyyy');
    case 'day':
    default:
      return format(date, 'yyyy-MM-dd');
  }
}

export function DatePicker({
  value,
  dateType = 'day',
  onChange,
  placeholder = 'Select date',
  name,
  required,
}: DatePickerProps) {
  const detailsRef = useRef<HTMLDetailsElement>(null);
  const [month, setMonth] = useState<Date>(value || new Date());

  // Close on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (detailsRef.current && !detailsRef.current.contains(event.target as Node)) {
        detailsRef.current.removeAttribute('open');
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Handle day selection
  const handleDaySelect = (date: Date | undefined) => {
    if (date) {
      onChange(date, 'day');
      detailsRef.current?.removeAttribute('open');
    }
  };

  // Handle week selection (click on week number)
  const handleWeekClick = (weekDate: Date) => {
    // Use start of week (Monday) as the canonical date for the week
    const weekStart = startOfWeek(weekDate, { weekStartsOn: 1 });
    onChange(weekStart, 'week');
    detailsRef.current?.removeAttribute('open');
  };

  // Handle month selection (click on month caption)
  const handleMonthClick = (monthDate: Date) => {
    // Use first day of month as the canonical date
    const firstOfMonth = new Date(monthDate.getFullYear(), monthDate.getMonth(), 1);
    onChange(firstOfMonth, 'month');
    detailsRef.current?.removeAttribute('open');
  };

  // Navigate to today
  const handleTodayClick = () => {
    setMonth(new Date());
  };

  // Check if a day is in the selected week
  const isInSelectedWeek = (day: Date): boolean => {
    if (!value || dateType !== 'week') return false;
    const dayWeek = getISOWeek(day);
    const dayYear = getYear(day);
    const selectedWeek = getISOWeek(value);
    const selectedYear = getYear(value);
    return dayWeek === selectedWeek && dayYear === selectedYear;
  };

  // Check if a day is in the selected month
  const isInSelectedMonth = (day: Date): boolean => {
    if (!value || dateType !== 'month') return false;
    return day.getMonth() === value.getMonth() && day.getFullYear() === value.getFullYear();
  };

  return (
    <details className="dropdown w-full" ref={detailsRef}>
      <summary className="input input-bordered input-sm w-full text-left flex items-center gap-2 cursor-pointer list-none marker:hidden">
        <Calendar className="h-4 w-4 opacity-70" />
        <span className={value ? '' : 'opacity-50'}>
          {value ? formatDisplayValue(value, dateType) : placeholder}
        </span>
      </summary>

      {name && (
        <input
          type="hidden"
          name={name}
          value={value ? formatForSubmission(value, dateType) : ''}
          required={required}
        />
      )}

      <div className="dropdown-content bg-base-100 rounded-box shadow-xl z-[9999] p-2 mt-1">
        <DayPicker
          mode="single"
          month={month}
          onMonthChange={setMonth}
          selected={dateType === 'day' ? value : undefined}
          onSelect={handleDaySelect}
          showOutsideDays
          showWeekNumber
          weekStartsOn={1}
          className="react-day-picker"
          modifiers={{
            selectedWeek: dateType === 'week' ? isInSelectedWeek : () => false,
            selectedMonth: dateType === 'month' ? isInSelectedMonth : () => false,
          }}
          modifiersClassNames={{
            selectedWeek: 'bg-primary/20',
            selectedMonth: 'bg-primary/20',
            today: 'text-primary font-bold',
          }}
          components={{
            Chevron: ({ orientation }) =>
              orientation === 'left'
                ? <ChevronLeft className="h-4 w-4" />
                : <ChevronRight className="h-4 w-4" />,
            // Custom week number - clickable to select week
            WeekNumber: ({ week }) => (
              <button
                type="button"
                onClick={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  handleWeekClick(week.days[0].date);
                }}
                className="btn btn-ghost btn-xs w-8 h-8 text-xs font-normal opacity-60 hover:opacity-100 hover:bg-primary hover:text-primary-content"
                title={`Select Week ${getISOWeek(week.days[0].date)}`}
              >
                {getISOWeek(week.days[0].date)}
              </button>
            ),
          }}
          footer={
            <div className="flex justify-center gap-2 pt-2 border-t border-base-300 mt-2">
              <button
                type="button"
                onClick={handleTodayClick}
                className="btn btn-ghost btn-xs"
              >
                Today
              </button>
              <button
                type="button"
                onClick={(e) => {
                  e.preventDefault();
                  handleMonthClick(month);
                }}
                className="btn btn-ghost btn-xs"
              >
                Select Month
              </button>
            </div>
          }
        />
      </div>
    </details>
  );
}
