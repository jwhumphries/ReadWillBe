import {useEffect, useRef, useCallback} from 'react';

interface UsePollingOptions {
  interval: number;
  enabled?: boolean;
  immediate?: boolean;
}

export function usePolling(
  callback: () => void | Promise<void>,
  options: UsePollingOptions,
) {
  const {interval, enabled = true, immediate = true} = options;
  const savedCallback = useRef(callback);
  const intervalRef = useRef<number | null>(null);

  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  const start = useCallback(() => {
    if (intervalRef.current) return;

    const runCallback = () => {
      Promise.resolve(savedCallback.current()).catch(err => {
        console.error('usePolling: callback rejected', err);
      });
    };

    if (immediate) {
      runCallback();
    }

    intervalRef.current = window.setInterval(runCallback, interval);
  }, [interval, immediate]);

  const stop = useCallback(() => {
    if (intervalRef.current) {
      window.clearInterval(intervalRef.current);
      intervalRef.current = null;
    }
  }, []);

  useEffect(() => {
    if (enabled) {
      start();
    } else {
      stop();
    }
    return stop;
  }, [enabled, start, stop]);

  return {start, stop};
}
