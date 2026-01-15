import { useRef, useEffect } from 'react';
import { DayPicker } from 'react-day-picker';
import { format } from 'date-fns';
import { Calendar, ChevronLeft, ChevronRight } from 'lucide-react';

interface DatePickerProps {
  value?: Date;
  onChange: (date: Date | undefined) => void;
  placeholder?: string;
  name?: string;
  required?: boolean;
}

export function DatePicker({
  value,
  onChange,
  placeholder = 'Select date',
  name,
  required,
}: DatePickerProps) {
  const detailsRef = useRef<HTMLDetailsElement>(null);

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

  const handleSelect = (date: Date | undefined) => {
    onChange(date);
    if (date) {
      detailsRef.current?.removeAttribute('open');
    }
  };

  return (
    <details className="dropdown w-full" ref={detailsRef}>
      <summary className="input input-bordered input-sm w-full text-left flex items-center gap-2 cursor-pointer list-none marker:hidden">
        <Calendar className="h-4 w-4 opacity-70" />
        <span className={value ? '' : 'opacity-50'}>
          {value ? format(value, 'MMM d, yyyy') : placeholder}
        </span>
      </summary>

      {name && (
        <input
          type="hidden"
          name={name}
          value={value ? format(value, 'yyyy-MM-dd') : ''}
          required={required}
        />
      )}

      <div className="dropdown-content bg-base-100 rounded-box shadow-xl z-[9999] p-2 mt-1">
        <DayPicker
          mode="single"
          selected={value}
          onSelect={handleSelect}
          showOutsideDays
          className="react-day-picker"
          components={{
            Chevron: ({ orientation }) =>
              orientation === 'left'
                ? <ChevronLeft className="h-4 w-4" />
                : <ChevronRight className="h-4 w-4" />,
          }}
        />
      </div>
    </details>
  );
}