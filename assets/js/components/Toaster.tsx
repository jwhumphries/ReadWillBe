import { useEffect } from 'react';
import { useToast } from '../context/ToastContext';

const alertClasses: Record<string, string> = {
  success: 'alert-success',
  error: 'alert-error',
  info: 'alert-info',
  warning: 'alert-warning',
};

export function Toaster() {
  const { toasts, addToast, removeToast } = useToast();

  // Listen for toast events (for use outside React components)
  useEffect(() => {
    const handler = (e: CustomEvent) => {
      addToast(e.detail.message, e.detail.type);
    };
    window.addEventListener('toast', handler as EventListener);
    return () => window.removeEventListener('toast', handler as EventListener);
  }, [addToast]);

  if (toasts.length === 0) return null;

  return (
    <div className="toast toast-top toast-end z-50">
      {toasts.map(t => (
        <div
          key={t.id}
          className={`alert ${alertClasses[t.type]} cursor-pointer`}
          onClick={() => removeToast(t.id)}
        >
          <span>{t.message}</span>
        </div>
      ))}
    </div>
  );
}

// Re-export toast for existing imports
export { toast } from '../context/ToastContext';
