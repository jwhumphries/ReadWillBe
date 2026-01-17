import React, { createContext, useContext, useState, useCallback, useRef } from 'react';

type ToastType = 'success' | 'error' | 'info' | 'warning';

interface Toast {
  id: number;
  message: string;
  type: ToastType;
}

interface ToastContextValue {
  toasts: Toast[];
  addToast: (message: string, type?: ToastType) => void;
  removeToast: (id: number) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const idRef = useRef(0);

  const removeToast = useCallback((id: number) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  }, []);

  const addToast = useCallback((message: string, type: ToastType = 'info') => {
    const id = ++idRef.current;
    setToasts(prev => [...prev, { id, message, type }]);

    // Auto-dismiss after 3 seconds
    setTimeout(() => removeToast(id), 3000);
  }, [removeToast]);

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast }}>
      {children}
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
}

// Convenience methods matching Sonner's API
export const toast = {
  success: (message: string) => {
    window.dispatchEvent(new CustomEvent('toast', { detail: { message, type: 'success' } }));
  },
  error: (message: string) => {
    window.dispatchEvent(new CustomEvent('toast', { detail: { message, type: 'error' } }));
  },
  info: (message: string) => {
    window.dispatchEvent(new CustomEvent('toast', { detail: { message, type: 'info' } }));
  },
  warning: (message: string) => {
    window.dispatchEvent(new CustomEvent('toast', { detail: { message, type: 'warning' } }));
  },
};
