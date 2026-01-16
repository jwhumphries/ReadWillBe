import React, { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Check, Undo2 } from 'lucide-react';
import { getCsrfToken } from '../hooks/useCsrf';
import { toast } from './Toaster';

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
    const queryClient = useQueryClient();

    const mutation = useMutation({
        mutationFn: async (readingId: number) => {
            const endpoint = actionType === 'complete'
                ? `/reading/${readingId}/complete`
                : `/reading/${readingId}/uncomplete`;

            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'X-CSRF-Token': getCsrfToken(),
                },
            });

            if (!response.ok) {
                throw new Error(`Failed to ${actionType} reading`);
            }

            return readingId;
        },
        onSuccess: (readingId: number) => {
            // Remove the reading from the list on success
            setReadings(prev => prev.filter(r => r.id !== readingId));
            toast.success(actionType === 'complete' ? 'Reading completed!' : 'Reading marked as incomplete');
            queryClient.invalidateQueries({ queryKey: ['notifications'] });
        },
        onError: () => {
            toast.error(`Failed to ${actionType} reading`);
        },
    });

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
                const isLoading = mutation.isPending && mutation.variables === reading.id;

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
                                    onClick={() => mutation.mutate(reading.id)}
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
                                            <Check className="h-4 w-4" />
                                            Complete
                                        </>
                                    ) : (
                                        <>
                                            <Undo2 className="h-4 w-4" />
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
