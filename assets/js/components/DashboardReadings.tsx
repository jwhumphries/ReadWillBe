import React, { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Check, AlertTriangle, Calendar } from 'lucide-react';
import { getCsrfToken } from '../hooks/useCsrf';
import { toast } from './Toaster';

interface Reading {
  id: number;
  date: string;
  content: string;
  status: 'pending' | 'completed';
  isOverdue: boolean;
}

interface Plan {
  id: number;
  title: string;
}

interface PlanGroup {
  plan: Plan;
  readings: Reading[];
  hasOverdue: boolean;
}

interface DashboardReadingsProps {
  planGroups: PlanGroup[];
}

export const DashboardReadings: React.FC<DashboardReadingsProps> = ({
  planGroups: initialPlanGroups,
}) => {
  const [planGroups, setPlanGroups] = useState(initialPlanGroups);
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: async (readingId: number) => {
      const response = await fetch(`/reading/${readingId}/complete`, {
        method: 'POST',
        headers: {
          'X-CSRF-Token': getCsrfToken(),
        },
      });

      if (!response.ok) {
        throw new Error('Failed to complete reading');
      }

      return readingId;
    },
    onSuccess: (readingId: number) => {
      // Remove the reading from its plan group
      setPlanGroups(prev =>
        prev
          .map(group => ({
            ...group,
            readings: group.readings.filter(r => r.id !== readingId),
            hasOverdue: group.readings
              .filter(r => r.id !== readingId)
              .some(r => r.isOverdue),
          }))
          .filter(group => group.readings.length > 0)
      );
      toast.success('Reading completed!');
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
    onError: () => {
      toast.error('Failed to complete reading');
    },
  });

  if (planGroups.length === 0) {
    return (
      <div className="hero bg-base-200 rounded-box py-12">
        <div className="hero-content text-center">
          <div className="max-w-md">
            <div className="mx-auto w-24 h-24 mb-4 opacity-20 text-base-content">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="w-full h-full">
                <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"></path>
                <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"></path>
              </svg>
            </div>
            <h3 className="text-2xl font-bold">All caught up!</h3>
            <p className="py-6">No readings scheduled for today. You can relax or check your reading plans.</p>
            <a href="/plans" className="btn btn-primary">View Plans</a>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {planGroups.map((group) => (
        <div
          key={group.plan.id}
          className={`card bg-base-200 shadow-xl ${group.hasOverdue ? 'card-border border-error' : ''}`}
        >
          <div className="card-body p-0">
            <div className="flex items-center gap-2 p-6 pb-0">
              <h3 className="card-title flex-1">{group.plan.title}</h3>
              {group.hasOverdue && (
                <div className="badge badge-error gap-2">
                  <svg xmlns="http://www.w3.org/2000/svg" className="inline-block h-4 w-4 stroke-current" fill="none" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  Overdue
                </div>
              )}
            </div>
            <div className="divide-y divide-base-300">
              {group.readings.map((reading) => {
                const isLoading = mutation.isPending && mutation.variables === reading.id;

                return (
                  <div key={reading.id} className="p-6">
                    <div className="flex items-start gap-3">
                      {reading.isOverdue && (
                        <div className="text-error mt-1" title="Overdue">
                          <AlertTriangle className="h-5 w-5" />
                        </div>
                      )}
                      <div className="flex-1">
                        <p className="leading-relaxed">{reading.content}</p>
                        <div className="flex items-center justify-between mt-4">
                          <div className="flex items-center gap-2 text-sm opacity-75">
                            <Calendar className="h-4 w-4" />
                            <span>{reading.date}</span>
                          </div>
                          <button
                            type="button"
                            onClick={() => mutation.mutate(reading.id)}
                            disabled={isLoading}
                            className="btn btn-primary btn-sm gap-2"
                          >
                            {isLoading ? (
                              <span className="loading loading-spinner loading-xs" />
                            ) : (
                              <>
                                <Check className="h-5 w-5" />
                                Complete
                              </>
                            )}
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};
