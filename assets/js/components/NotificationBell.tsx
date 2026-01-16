import React, { useState, useRef, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Bell } from 'lucide-react';
import { getCsrfToken } from '../hooks/useCsrf';

interface Reading {
    id: number;
    date: string;
    content: string;
    plan?: {
        id: number;
        title: string;
    };
}

interface NotificationBellProps {
    initialCount?: number;
    pollInterval?: number;
}

export const NotificationBell: React.FC<NotificationBellProps> = ({
    initialCount = 0,
    pollInterval = 60000,
}) => {
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);

    // Fetch notification count with React Query
    const { data: countData } = useQuery({
        queryKey: ['notifications', 'count'],
        queryFn: async () => {
            const response = await fetch('/api/notifications/count', {
                headers: { 'X-CSRF-Token': getCsrfToken() },
            });
            if (!response.ok) throw new Error('Failed to fetch count');
            return response.json() as Promise<{ count: number }>;
        },
        initialData: { count: initialCount },
        refetchInterval: pollInterval,
    });

    // Fetch readings when dropdown is opened
    const { data: readingsData, isLoading, refetch } = useQuery({
        queryKey: ['notifications', 'readings'],
        queryFn: async () => {
            const response = await fetch('/api/notifications/readings', {
                headers: { 'X-CSRF-Token': getCsrfToken() },
            });
            if (!response.ok) throw new Error('Failed to fetch readings');
            return response.json() as Promise<{ readings: Reading[] }>;
        },
        enabled: false, // Only fetch on demand
    });

    const count = countData?.count ?? 0;
    const readings = readingsData?.readings ?? [];

    // Handle click outside to close dropdown
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleToggle = () => {
        const newIsOpen = !isOpen;
        setIsOpen(newIsOpen);
        if (newIsOpen) {
            refetch();
        }
    };

    return (
        <div className="dropdown dropdown-end" ref={dropdownRef}>
            <button
                type="button"
                className="btn btn-ghost btn-circle"
                onClick={handleToggle}
                aria-label="Notifications"
            >
                <div className="indicator">
                    <Bell className="h-5 w-5" />
                    {count > 0 && (
                        <span className="badge badge-sm badge-primary indicator-item">
                            {count > 99 ? '99+' : count}
                        </span>
                    )}
                </div>
            </button>

            {isOpen && (
                <div className="dropdown-content menu bg-base-200 rounded-box z-50 w-80 p-2 shadow-lg mt-2">
                    <div className="px-2 py-1 font-semibold border-b border-base-300 mb-2">
                        Today's Readings
                    </div>
                    {isLoading ? (
                        <div className="flex justify-center py-4">
                            <span className="loading loading-spinner loading-sm"></span>
                        </div>
                    ) : readings.length === 0 ? (
                        <div className="text-center py-4 text-base-content/60">
                            No readings for today
                        </div>
                    ) : (
                        <ul className="max-h-64 overflow-y-auto">
                            {readings.map((reading: Reading) => (
                                <li key={reading.id}>
                                    <a
                                        href={`/plans/${reading.plan?.id}`}
                                        className="flex flex-col items-start gap-1"
                                    >
                                        <span className="text-sm font-medium truncate w-full">
                                            {reading.content}
                                        </span>
                                        {reading.plan && (
                                            <span className="text-xs opacity-60">
                                                {reading.plan.title}
                                            </span>
                                        )}
                                    </a>
                                </li>
                            ))}
                        </ul>
                    )}
                    <div className="border-t border-base-300 mt-2 pt-2">
                        <a href="/dashboard" className="btn btn-ghost btn-sm w-full">
                            View Dashboard
                        </a>
                    </div>
                </div>
            )}
        </div>
    );
};
