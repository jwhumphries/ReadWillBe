import { useState, useRef, useEffect } from 'react';
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
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Close on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="input input-bordered input-sm w-full text-left flex items-center gap-2"
      >
        <Calendar className="h-4 w-4 opacity-70" />
        <span className={value ? '' : 'opacity-50'}>
          {value ? format(value, 'MMM d, yyyy') : placeholder}
        </span>
      </button>

      {name && (
        <input
          type="hidden"
          name={name}
          value={value ? format(value, 'yyyy-MM-dd') : ''}
          required={required}
        />
      )}

      {isOpen && (
        <div className="dropdown dropdown-open">
          <div className="dropdown-content bg-base-100 rounded-box shadow-xl z-[9999]">
            <DayPicker
              mode="single"
              selected={value}
              onSelect={(date: Date | undefined) => {
                onChange(date);
                setIsOpen(false);
              }}
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
        </div>
      )}
    </div>
  );
}
