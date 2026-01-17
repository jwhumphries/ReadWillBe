import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getCsrfToken } from './useCsrf';
import { toast } from '../components/Toaster';
import type { Reading, Plan } from '../types';

// Fetch wrapper with CSRF token
async function fetchApi<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...init,
    headers: {
      'X-CSRF-Token': getCsrfToken(),
      ...init?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }

  return response.json();
}

// Hook for fetching notification count
export function useNotificationCount() {
  return useQuery({
    queryKey: ['notifications', 'count'],
    queryFn: () => fetchApi<{ count: number }>('/api/notifications/count'),
    refetchInterval: 60000, // Poll every 60s
  });
}

// Hook for fetching notifications
export function useNotifications() {
  return useQuery({
    queryKey: ['notifications', 'readings'],
    queryFn: () => fetchApi<{ readings: Reading[] }>('/api/notifications/readings'),
    enabled: false, // Only fetch on demand
  });
}

// Hook for completing a reading
export function useCompleteReading() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (readingId: number) =>
      fetchApi(`/reading/${readingId}/complete`, { method: 'POST' }),
    onSuccess: () => {
      toast.success('Reading completed!');
      queryClient.invalidateQueries({ queryKey: ['readings'] });
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
    onError: () => {
      toast.error('Failed to complete reading');
    },
  });
}

// Hook for uncompleting a reading
export function useUncompleteReading() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (readingId: number) =>
      fetchApi(`/reading/${readingId}/uncomplete`, { method: 'POST' }),
    onSuccess: () => {
      toast.success('Reading marked as incomplete');
      queryClient.invalidateQueries({ queryKey: ['readings'] });
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
    onError: () => {
      toast.error('Failed to undo reading');
    },
  });
}

// Hook for fetching plan status (with polling)
export function usePlanStatus(planId: number, enabled: boolean) {
  return useQuery({
    queryKey: ['plan', planId, 'status'],
    queryFn: () => fetchApi<{ status: string }>(`/api/plans/${planId}/status`),
    refetchInterval: enabled ? 5000 : false, // Poll every 5s while processing
    enabled,
  });
}

// Hook for deleting a plan
export function useDeletePlan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (planId: number) =>
      fetchApi(`/plans/${planId}`, { method: 'DELETE' }),
    onSuccess: () => {
      toast.success('Plan deleted');
      queryClient.invalidateQueries({ queryKey: ['plans'] });
    },
    onError: () => {
      toast.error('Failed to delete plan');
    },
  });
}

// Hook for saving draft
export function useSaveDraft() {
  return useMutation({
    mutationFn: async (data: { title?: string; readings?: Array<{ date: string; content: string }> }) => {
      const response = await fetch('/plans/draft', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getCsrfToken(),
        },
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error('Failed to save draft');
    },
    onSuccess: () => {
      toast.success('Draft saved');
    },
    onError: () => {
      toast.error('Failed to save draft');
    },
  });
}

// Hook for creating manual plan
export function useCreateManualPlan() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (data: { title: string; readings: Array<{ date: string; content: string }> }) => {
      const response = await fetch('/plans/create-manual', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getCsrfToken(),
        },
        body: JSON.stringify(data),
      });
      if (!response.ok) throw new Error('Failed to create plan');
      return response.json();
    },
    onSuccess: () => {
      toast.success('Plan created successfully!');
      queryClient.invalidateQueries({ queryKey: ['plans'] });
    },
    onError: () => {
      toast.error('Failed to create plan');
    },
  });
}
