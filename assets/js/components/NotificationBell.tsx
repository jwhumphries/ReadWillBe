import React, { useState, useRef, useEffect } from 'react';
import { getCsrfToken } from '../hooks/useCsrf';
import { usePolling } from '../hooks/usePolling';

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
    pollInterval = 30000,
}) => {
    const [count, setCount] = useState(initialCount);
    const [readings, setReadings] = useState<Reading[]>([]);
    const [isOpen, setIsOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);

    // Fetch notification count
    const fetchCount = async () => {
        try {
            const response = await fetch('/api/notifications/count', {
                headers: {
                    'X-CSRF-Token': getCsrfToken(),
                },
            });
            if (response.ok) {
                const data = await response.json();
                setCount(data.count);
            }
        } catch (e) {
            console.error('Failed to fetch notification count', e);
        }
    };

    // Poll for notification count
    usePolling(fetchCount, {
        interval: pollInterval,
        enabled: true,
        immediate: false,
    });

    // Fetch dropdown content when opened
    const fetchReadings = async () => {
        setLoading(true);
        try {
            const response = await fetch('/api/notifications/readings', {
                headers: {
                    'X-CSRF-Token': getCsrfToken(),
                },
            });
            if (response.ok) {
                const data = await response.json();
                setReadings(data.readings || []);
            }
        } catch (e) {
            console.error('Failed to fetch notifications', e);
        } finally {
            setLoading(false);
        }
    };

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
            fetchReadings();
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
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
                    </svg>
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
                    {loading ? (
                        <div className="flex justify-center py-4">
                            <span className="loading loading-spinner loading-sm"></span>
                        </div>
                    ) : readings.length === 0 ? (
                        <div className="text-center py-4 text-base-content/60">
                            No readings for today
                        </div>
                    ) : (
                        <ul className="max-h-64 overflow-y-auto">
                            {readings.map((reading) => (
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
