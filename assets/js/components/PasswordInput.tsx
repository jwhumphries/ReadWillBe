import { useState, forwardRef } from 'react';
import { Eye, EyeOff } from 'lucide-react';

interface PasswordInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string;
  error?: string;
}

export const PasswordInput = forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ label, error, className = '', id, ...props }, ref) => {
    const [visible, setVisible] = useState(false);

    return (
      <div className="space-y-1">
        {label && (
          <label htmlFor={id} className="text-sm font-medium">
            {label}
          </label>
        )}
        <label className="input input-bordered flex items-center gap-2 pr-0 w-full">
          <input
            ref={ref}
            id={id}
            type={visible ? 'text' : 'password'}
            className={`grow ${className}`}
            {...props}
          />
          <button
            type="button"
            className="btn btn-ghost btn-sm btn-circle"
            onClick={() => setVisible(!visible)}
            aria-label={visible ? 'Hide password' : 'Show password'}
          >
            {visible ? (
              <EyeOff className="h-5 w-5 opacity-70" />
            ) : (
              <Eye className="h-5 w-5 opacity-70" />
            )}
          </button>
        </label>
        {error && (
          <p className="text-sm text-error">{error}</p>
        )}
      </div>
    );
  }
);

PasswordInput.displayName = 'PasswordInput';
