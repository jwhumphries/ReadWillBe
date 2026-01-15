import React, { useState } from 'react';
import { getCsrfToken } from '../hooks/useCsrf';

interface Reading {
    id: number;
    date: string;
    content: string;
    status: 'pending' | 'completed';
    completedAt?: string;
    plan?: {
        id: number;
        title: string;
    };
}

interface ReadingListProps {
    readings: Reading[];
    showPlanTitle?: boolean;
    actionType: 'complete' | 'uncomplete';
}

export const ReadingList: React.FC<ReadingListProps> = ({
    readings: initialReadings,
    showPlanTitle = true,
    actionType,
}) => {
    const [readings, setReadings] = useState(initialReadings);
    const [loadingIds, setLoadingIds] = useState<Set<number>>(new Set());

    const handleAction = async (readingId: number) => {
        setLoadingIds(prev => new Set(prev).add(readingId));

        const endpoint = actionType === 'complete'
            ? `/reading/${readingId}/complete`
            : `/reading/${readingId}/uncomplete`;

        try {
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'X-CSRF-Token': getCsrfToken(),
                },
            });

            if (response.ok) {
                // Remove the reading from the list on success
                setReadings(prev => prev.filter(r => r.id !== readingId));
            }
        } catch (e) {
            console.error(`Failed to ${actionType} reading`, e);
        } finally {
            setLoadingIds(prev => {
                const next = new Set(prev);
                next.delete(readingId);
                return next;
            });
        }
    };

    if (readings.length === 0) {
        return (
            <div className="text-center py-8 text-base-content/60">
                {actionType === 'complete'
                    ? 'No pending readings'
                    : 'No completed readings'}
            </div>
        );
    }

    return (
        <div className="space-y-4">
            {readings.map((reading) => {
                const isLoading = loadingIds.has(reading.id);

                return (
                    <div
                        key={reading.id}
                        className="card bg-base-200 shadow-sm"
                    >
                        <div className="card-body p-4">
                            <div className="flex items-center justify-between gap-4">
                                <div className="flex-1 min-w-0">
                                    <p className="font-medium truncate">{reading.content}</p>
                                    <div className="flex gap-2 text-sm text-base-content/60 mt-1">
                                        <span>{reading.date}</span>
                                        {showPlanTitle && reading.plan && (
                                            <>
                                                <span>•</span>
                                                <a
                                                    href={`/plans/${reading.plan.id}`}
                                                    className="hover:underline"
                                                >
                                                    {reading.plan.title}
                                                </a>
                                            </>
                                        )}
                                    </div>
                                </div>
                                <button
                                    type="button"
                                    onClick={() => handleAction(reading.id)}
                                    disabled={isLoading}
                                    className={`btn btn-sm ${
                                        actionType === 'complete'
                                            ? 'btn-primary'
                                            : 'btn-ghost'
                                    }`}
                                >
                                    {isLoading ? (
                                        <span className="loading loading-spinner loading-xs" />
                                    ) : actionType === 'complete' ? (
                                        <>
                                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                            </svg>
                                            Complete
                                        </>
                                    ) : (
                                        <>
                                            <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                                            </svg>
                                            Undo
                                        </>
                                    )}
                                </button>
                            </div>
                        </div>
                    </div>
                );
            })}
        </div>
    );
};
