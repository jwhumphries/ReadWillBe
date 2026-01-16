import { useState, useCallback } from 'react';
import { getCsrfToken } from './useCsrf';

interface UseFetchOptions {
  onSuccess?: (data: unknown) => void;
  onError?: (error: Error) => void;
}

export function useFetch<T = unknown>(options: UseFetchOptions = {}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [data, setData] = useState<T | null>(null);

  const execute = useCallback(async (url: string, init: RequestInit = {}) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(url, {
        ...init,
        headers: {
          'X-CSRF-Token': getCsrfToken(),
          ...init.headers,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // Check if response is HTML (for full page responses)
      const contentType = response.headers.get('content-type');
      if (contentType?.includes('text/html')) {
        // For HTML responses, replace the body
        const html = await response.text();
        document.body.innerHTML = html;
        return null;
      }

      const result = await response.json();
      setData(result as T);
      options.onSuccess?.(result);
      return result as T;
    } catch (err) {
      const error = err instanceof Error ? err : new Error(String(err));
      setError(error);
      options.onError?.(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [options]);

  return { execute, loading, error, data };
}
