import React, { useState, useRef, useEffect } from 'react';
import { Plus, X, Save } from 'lucide-react';
import { DatePicker, DateType } from './DatePicker';
import { format, parse, getISOWeek, getYear } from 'date-fns';

interface Reading {
    id: string;
    date: string;
    dateType: DateType;
    content: string;
}

interface PlanEditorProps {
    initialReadings?: Reading[];
    initialTitle?: string;
    buttonText?: string;
}

// Format date for storage based on type (matches Go parseDate expectations)
function formatDateForStorage(date: Date, dateType: DateType): string {
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

// Parse stored date string to Date object
function parseStoredDate(dateStr: string, dateType: DateType): Date | undefined {
    try {
        switch (dateType) {
            case 'week': {
                // Parse "2025-W03" format
                const match = dateStr.match(/^(\d{4})-W(\d{2})$/);
                if (match) {
                    const year = parseInt(match[1], 10);
                    const week = parseInt(match[2], 10);
                    // Get first day of the ISO week
                    const jan4 = new Date(year, 0, 4);
                    const dayOfWeek = jan4.getDay() || 7;
                    const firstMonday = new Date(jan4);
                    firstMonday.setDate(jan4.getDate() - dayOfWeek + 1);
                    firstMonday.setDate(firstMonday.getDate() + (week - 1) * 7);
                    return firstMonday;
                }
                break;
            }
            case 'month': {
                // Parse "January 2025" format
                const parsed = parse(dateStr, 'MMMM yyyy', new Date());
                if (!isNaN(parsed.getTime())) {
                    return parsed;
                }
                break;
            }
            case 'day':
            default: {
                // Parse "2025-01-15" format
                const date = new Date(dateStr + 'T00:00:00');
                if (!isNaN(date.getTime())) {
                    return date;
                }
                break;
            }
        }
    } catch {
        // Fall through to return undefined
    }
    return undefined;
}

// Detect date type from string format
function detectDateType(dateStr: string): DateType {
    if (/^\d{4}-W\d{2}$/.test(dateStr)) {
        return 'week';
    }
    if (/^[A-Z][a-z]+ \d{4}$/.test(dateStr)) {
        return 'month';
    }
    return 'day';
}

export const PlanEditor: React.FC<PlanEditorProps> = ({
    initialReadings = [],
    initialTitle = '',
    buttonText = 'Create Plan'
}) => {
    const [title, setTitle] = useState(initialTitle);
    const [readings, setReadings] = useState<Reading[]>(() => {
        if (initialReadings.length > 0) {
            return initialReadings.map(r => ({
                ...r,
                id: r.id || crypto.randomUUID(),
                // Detect dateType from date format if not provided
                dateType: r.dateType || detectDateType(r.date)
            }));
        }
        return [{
            id: crypto.randomUUID(),
            date: format(new Date(), 'yyyy-MM-dd'),
            dateType: 'day' as DateType,
            content: ''
        }];
    });

    // Ref to track the last added reading ID for auto-focus
    const lastAddedIdRef = useRef<string | null>(null);
    const inputRefs = useRef<{ [key: string]: HTMLInputElement | null }>({});

    useEffect(() => {
        if (lastAddedIdRef.current && inputRefs.current[lastAddedIdRef.current]) {
            inputRefs.current[lastAddedIdRef.current]?.focus();
            lastAddedIdRef.current = null;
        }
    }, [readings]);

    const addReading = () => {
        const newId = crypto.randomUUID();
        // Use the date and dateType of the last reading if available
        const lastReading = readings[readings.length - 1];
        const lastDate = lastReading?.date || format(new Date(), 'yyyy-MM-dd');
        const lastDateType = lastReading?.dateType || 'day';

        setReadings([...readings, {
            id: newId,
            date: lastDate,
            dateType: lastDateType,
            content: ''
        }]);
        lastAddedIdRef.current = newId;
    };

    const handleDateChange = (id: string, date: Date | undefined, dateType: DateType) => {
        if (date) {
            setReadings(readings.map(r =>
                r.id === id
                    ? { ...r, date: formatDateForStorage(date, dateType), dateType }
                    : r
            ));
        }
    };

    const handleContentKeyDown = (e: React.KeyboardEvent, id: string) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            addReading();
        }
    };

    const removeReading = (id: string) => {
        if (readings.length <= 1) {
            // If it's the last row, just clear the content instead of removing
            setReadings(readings.map(r =>
                r.id === id ? { ...r, content: '' } : r
            ));
            return;
        }
        setReadings(readings.filter(r => r.id !== id));
    };

    const hasValidReadings = readings.some(r => r.content.trim() !== '');

    return (
        <div className="card bg-base-200 shadow-xl">
            <div className="card-body">
                {/* Title input */}
                <div className="form-control w-full mb-6">
                    <label className="label">
                        <span className="label-text font-bold">Plan Title</span>
                    </label>
                    <input
                        type="text"
                        name="title"
                        value={title}
                        onChange={(e) => setTitle(e.target.value)}
                        placeholder="Enter plan title"
                        className="input input-bordered w-full"
                        required
                    />
                </div>

                {/* Readings table */}
                <div className="overflow-visible">
                    <table className="table w-full">
                        <thead>
                            <tr>
                                <th className="w-52">Date</th>
                                <th>Reading Content</th>
                                <th className="w-16"></th>
                            </tr>
                        </thead>
                        <tbody>
                            {readings.map((reading) => (
                                <tr key={reading.id} className="hover:bg-base-300/50 transition-colors">
                                    <td className="overflow-visible align-top">
                                        <DatePicker
                                            value={parseStoredDate(reading.date, reading.dateType)}
                                            dateType={reading.dateType}
                                            onChange={(date, dateType) => handleDateChange(reading.id, date, dateType)}
                                            placeholder="Select date"
                                        />
                                    </td>
                                    <td className="align-top">
                                        <input
                                            ref={el => { inputRefs.current[reading.id] = el }}
                                            type="text"
                                            value={reading.content}
                                            onChange={(e) => setReadings(readings.map(r =>
                                                r.id === reading.id ? { ...r, content: e.target.value } : r
                                            ))}
                                            onKeyDown={(e) => handleContentKeyDown(e, reading.id)}
                                            placeholder="What to read..."
                                            className="input input-bordered input-sm w-full"
                                            autoComplete="off"
                                        />
                                    </td>
                                    <td className="align-top">
                                        <div className="tooltip tooltip-error tooltip-bottom" data-tip="Remove reading">
                                            <button
                                                type="button"
                                                onClick={() => removeReading(reading.id)}
                                                className="btn btn-ghost btn-xs text-error"
                                                aria-label="Remove reading"
                                                disabled={readings.length === 1 && !reading.content}
                                            >
                                                <X className="h-4 w-4" />
                                            </button>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                <div className="mt-4">
                    <button
                        type="button"
                        onClick={addReading}
                        className="btn btn-ghost btn-sm gap-2"
                    >
                        <Plus className="h-4 w-4" />
                        Add Reading
                    </button>
                </div>

                {/* HIDDEN INPUT: Serializes React state for form submission */}
                <input
                    type="hidden"
                    name="readingsJSON"
                    value={JSON.stringify(readings.filter(r => r.content.trim() !== ''))}
                />

                <div className="card-actions justify-end mt-6">
                    <a href="/plans" className="btn btn-ghost">Cancel</a>
                    <button
                        type="submit"
                        className="btn btn-primary gap-2"
                        disabled={!hasValidReadings || !title}
                    >
                        <Save className="h-4 w-4" />
                        {buttonText}
                    </button>
                </div>
            </div>
        </div>
    );
};
