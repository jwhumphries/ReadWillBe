import React, { useState } from 'react';
import { Plus, X, Save } from 'lucide-react';

interface Reading {
    id: string;
    date: string;
    content: string;
}

interface PlanEditorProps {
    initialReadings?: Reading[];
    initialTitle?: string;
}

export const PlanEditor: React.FC<PlanEditorProps> = ({
    initialReadings = [],
    initialTitle = ''
}) => {
    const [title, setTitle] = useState(initialTitle);
    const [readings, setReadings] = useState<Reading[]>(
        initialReadings.map(r => ({...r, id: r.id || crypto.randomUUID()}))
    );

    const addReading = () => {
        setReadings([...readings, { id: crypto.randomUUID(), date: '', content: '' }]);
    };

    const removeReading = (id: string) => {
        setReadings(readings.filter(r => r.id !== id));
    };

    const updateReading = (id: string, field: keyof Reading, value: string) => {
        setReadings(readings.map(r => r.id === id ? { ...r, [field]: value } : r));
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
                <div className="overflow-x-auto">
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
                                <tr key={reading.id}>
                                    <td>
                                        <input
                                            type="date"
                                            value={reading.date}
                                            onChange={(e) => updateReading(reading.id, 'date', e.target.value)}
                                            className="input input-bordered input-sm w-full"
                                            required
                                        />
                                    </td>
                                    <td>
                                        <input
                                            type="text"
                                            value={reading.content}
                                            onChange={(e) => updateReading(reading.id, 'content', e.target.value)}
                                            placeholder="What to read..."
                                            className="input input-bordered input-sm w-full"
                                            required
                                        />
                                    </td>
                                    <td>
                                        <button
                                            type="button"
                                            onClick={() => removeReading(reading.id)}
                                            className="btn btn-ghost btn-xs text-error"
                                            aria-label="Remove reading"
                                        >
                                            <X className="h-4 w-4" />
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                <button type="button" onClick={addReading} className="btn btn-outline btn-primary btn-sm gap-2">
                    <Plus className="h-4 w-4" />
                    Add Reading
                </button>

                {/* HIDDEN INPUT: Serializes React state for form submission */}
                <input type="hidden" name="readingsJSON" value={JSON.stringify(readings)} />

                <div className="card-actions justify-end mt-6">
                    <a href="/plans" className="btn btn-ghost">Cancel</a>
                    <button
                        type="submit"
                        className="btn btn-primary gap-2"
                        disabled={readings.length === 0 || !title}
                    >
                        <Save className="h-4 w-4" />
                        Create Plan
                    </button>
                </div>
            </div>
        </div>
    );
};
