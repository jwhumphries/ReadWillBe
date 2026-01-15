import React, { useState, useRef, useEffect } from 'react';
import { Plus, X, Save } from 'lucide-react';
import { DatePicker } from './DatePicker';
import { format } from 'date-fns';

interface Reading {
    id: string;
    date: string;
    content: string;
}

interface PlanEditorProps {
    initialReadings?: Reading[];
    initialTitle?: string;
    buttonText?: string;
}

export const PlanEditor: React.FC<PlanEditorProps> = ({
    initialReadings = [],
    initialTitle = '',
    buttonText = 'Create Plan'
}) => {
    const [title, setTitle] = useState(initialTitle);
    const [readings, setReadings] = useState<Reading[]>(
        initialReadings.length > 0
            ? initialReadings.map(r => ({ ...r, id: r.id || crypto.randomUUID() }))
            : [{ id: crypto.randomUUID(), date: format(new Date(), 'yyyy-MM-dd'), content: '' }]
    );

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
        // Use the date of the last reading if available, otherwise today
        const lastDate = readings.length > 0 ? readings[readings.length - 1].date : format(new Date(), 'yyyy-MM-dd');
        
        setReadings([...readings, {
            id: newId,
            date: lastDate,
            content: ''
        }]);
        lastAddedIdRef.current = newId;
    };

    const updateReading = (id: string, field: keyof Reading, value: string) => {
        setReadings(readings.map(r => r.id === id ? { ...r, [field]: value } : r));
    };

    const handleDateChange = (id: string, date: Date | undefined) => {
        if (date) {
            updateReading(id, 'date', format(date, 'yyyy-MM-dd'));
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
            // If it's the last row, just clear it instead of removing
            updateReading(id, 'content', '');
            return;
        }
        setReadings(readings.filter(r => r.id !== id));
    };

    const handleFormSubmit = (e: React.FormEvent) => {
        // Filter out empty rows before submitting
        const validReadings = readings.filter(r => r.content.trim() !== '');
        
        if (validReadings.length === 0 || !title) {
            e.preventDefault();
            return;
        }
        
        // Update hidden input with valid readings only? 
        // No, the hidden input is bound to state. We should probably update state or let the server handle it?
        // Let's just block submit if no valid readings.
        // Actually, let's allow submitting but maybe warn? 
        // The previous logic was: disabled={readings.length === 0 || !title}
        // Now readings is never empty.
    };

    const hasValidReadings = readings.some(r => r.content.trim() !== '');

    const parseDate = (dateStr: string) => {
        try {
            return new Date(dateStr + 'T00:00:00');
        } catch {
            return undefined;
        }
    };

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
                                <th className="w-48">Date</th>
                                <th>Reading Content</th>
                                <th className="w-16"></th>
                            </tr>
                        </thead>
                        <tbody>
                            {readings.map((reading) => (
                                <tr key={reading.id} className="hover:bg-base-300/50 transition-colors">
                                    <td className="overflow-visible align-top">
                                        <DatePicker
                                            value={parseDate(reading.date)}
                                            onChange={(date) => handleDateChange(reading.id, date)}
                                            placeholder="Select date"
                                        />
                                    </td>
                                    <td className="align-top">
                                        <input
                                            ref={el => { inputRefs.current[reading.id] = el }}
                                            type="text"
                                            value={reading.content}
                                            onChange={(e) => updateReading(reading.id, 'content', e.target.value)}
                                            onKeyDown={(e) => handleContentKeyDown(e, reading.id)}
                                            placeholder="What to read..."
                                            className="input input-bordered input-sm w-full"
                                            autoComplete="off"
                                        />
                                    </td>
                                    <td className="align-top">
                                        <button
                                            type="button"
                                            onClick={() => removeReading(reading.id)}
                                            className="btn btn-ghost btn-xs text-error"
                                            aria-label="Remove reading"
                                            disabled={readings.length === 1 && !reading.content}
                                        >
                                            <X className="h-4 w-4" />
                                        </button>
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
                        onClick={handleFormSubmit}
                    >
                        <Save className="h-4 w-4" />
                        {buttonText}
                    </button>
                </div>
            </div>
        </div>
    );
};