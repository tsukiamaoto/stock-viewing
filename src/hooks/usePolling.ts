import { useEffect, useRef } from 'react';

/**
 * Custom hook to handle recurrent polling/API fetching logic.
 * @param callback The function to execute on an interval
 * @param delay The interval delay in milliseconds. If null, polling is paused.
 * @param deps Dependencies that, when changed, should reset the interval and trigger an immediate run.
 */
export function usePolling(callback: () => void, delay: number | null, deps: React.DependencyList = []) {
  const savedCallback = useRef(callback);

  // Remember the latest callback if it changes.
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  // Set up the interval.
  useEffect(() => {
    // Don't schedule if delay is null
    if (delay === null) {
      return;
    }

    // Run immediately
    savedCallback.current();

    const id = setInterval(() => savedCallback.current(), delay);
    return () => clearInterval(id);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [delay, ...deps]);
}
