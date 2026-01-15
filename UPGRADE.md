# Component Upgrade Guide: React Libraries & UX Improvements

This document outlines opportunities to upgrade current JavaScript functionality to use established React libraries, and suggests common UX improvements that could enhance the ReadWillBe application.

## Table of Contents

1. [Current JavaScript Inventory](#current-javascript-inventory)
2. [Recommended React Libraries](#recommended-react-libraries)
3. [Component Upgrades](#component-upgrades)
4. [UX Improvements](#ux-improvements)
5. [Package Additions](#package-additions)
6. [Implementation Examples](#implementation-examples)

---

## Current JavaScript Inventory

### Standalone Files

| File | Purpose | Migration Recommendation |
|------|---------|-------------------------|
| `static/push-setup.js` | Web Push API integration | **Keep as vanilla JS** - Must work with Service Worker |
| `static/serviceWorker.js` | Push notification handling | **Keep as vanilla JS** - Required format |

### Inline JavaScript (layout.templ)

| Function | Purpose | Migration Recommendation |
|----------|---------|-------------------------|
| `togglePasswordVisibility()` | Show/hide password | Replace with React component |
| `getCsrfToken()` | Extract CSRF from cookie | Convert to React hook |
| `deleteDraftAndGoBack()` | Delete draft, redirect | Replace with React handler |
| `handleDateSelect()` | Calendar date selection | Replace with date picker library |
| `toggleCalendar()` | Show/hide calendar | Replace with date picker library |
| Modal trigger handlers | Open/close dialogs | Replace with React modal library |

---

## Recommended React Libraries

### High Priority (Significant UX Improvement)

| Library | Purpose | Bundle Size | Recommendation |
|---------|---------|-------------|----------------|
| **@tanstack/react-query** | Data fetching, caching | ~13KB | Strongly recommended for API calls |
| **sonner** | Toast notifications | ~5KB | Better UX than alerts |
| **react-hook-form** | Form management | ~9KB | Cleaner form handling |
| **lucide-react** | Icons | Tree-shakeable | Replace custom SVGs |

### Medium Priority (Good Enhancements)

| Library | Purpose | Bundle Size | Recommendation |
|---------|---------|-------------|----------------|
| **date-fns** | Date formatting | Tree-shakeable | Consistent date handling |
| **react-day-picker** | Calendar/date selection | ~15KB | Replace cally library |
| **@headlessui/react** | Accessible UI primitives | ~10KB | Optional - DaisyUI already good |

### Low Priority (Nice to Have)

| Library | Purpose | Bundle Size | Recommendation |
|---------|---------|-------------|----------------|
| **react-loading-skeleton** | Loading states | ~3KB | Better perceived performance |
| **framer-motion** | Animations | ~30KB | Only if animations needed |
| **zod** | Schema validation | ~13KB | Type-safe form validation |

---

## Component Upgrades

### 1. Password Input with Visibility Toggle

**Current Implementation (layout.templ:114-128):**
```javascript
function togglePasswordVisibility(input) {
  const btn = input.nextElementSibling;
  const eye = btn.querySelector(".icon-eye");
  const slash = btn.querySelector(".icon-eye-slash");

  if (input.type === "password") {
    input.type = "text";
    eye.classList.add("hidden");
    slash.classList.remove("hidden");
  } else {
    input.type = "password";
    eye.classList.remove("hidden");
    slash.classList.add("hidden");
  }
}
```

**Upgraded React Component:**

Create `react/components/PasswordInput.tsx`:
```tsx
import { useState, forwardRef } from 'react';
import { Eye, EyeOff } from 'lucide-react';

interface PasswordInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string;
  error?: string;
}

export const PasswordInput = forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ label, error, className = '', ...props }, ref) => {
    const [visible, setVisible] = useState(false);

    return (
      <div className="space-y-1">
        {label && (
          <label htmlFor={props.id} className="text-sm font-medium">
            {label}
          </label>
        )}
        <label className="input input-bordered flex items-center gap-2 pr-0 w-full">
          <input
            ref={ref}
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
```

**Usage:**
```tsx
<PasswordInput
  name="password"
  id="signin-password"
  required
  placeholder="••••••••"
  label="Password"
/>
```

### 2. Date Picker (Replace Cally)

**Current Implementation (layout.templ:171-191):**
```javascript
function handleDateSelect(event) {
  const date = new Date(event.target.value);
  document.getElementById("date-input").value = event.target.value;
  document.getElementById("date-display").textContent =
    date.toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  document.getElementById("calendar-dropdown").classList.add("hidden");
}

function toggleCalendar() {
  const dropdown = document.getElementById("calendar-dropdown");
  dropdown.classList.toggle("hidden");
}
```

**Upgraded with react-day-picker:**

Create `react/components/DatePicker.tsx`:
```tsx
import { useState, useRef, useEffect } from 'react';
import { DayPicker } from 'react-day-picker';
import { format } from 'date-fns';
import { Calendar, ChevronLeft, ChevronRight } from 'lucide-react';
import 'react-day-picker/dist/style.css';

interface DatePickerProps {
  value?: Date;
  onChange: (date: Date | undefined) => void;
  placeholder?: string;
  name?: string;
  required?: boolean;
}

export function DatePicker({
  value,
  onChange,
  placeholder = 'Select date',
  name,
  required,
}: DatePickerProps) {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Close on outside click
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="input input-bordered input-sm w-full text-left flex items-center gap-2"
      >
        <Calendar className="h-4 w-4 opacity-70" />
        <span className={value ? '' : 'opacity-50'}>
          {value ? format(value, 'MMM d, yyyy') : placeholder}
        </span>
      </button>

      {name && (
        <input
          type="hidden"
          name={name}
          value={value ? format(value, 'yyyy-MM-dd') : ''}
          required={required}
        />
      )}

      {isOpen && (
        <div className="absolute z-[9999] mt-1 card card-compact bg-base-100 shadow-xl">
          <div className="card-body p-4">
            <DayPicker
              mode="single"
              selected={value}
              onSelect={(date) => {
                onChange(date);
                setIsOpen(false);
              }}
              showOutsideDays
              className="!m-0"
              classNames={{
                day_selected: 'bg-primary text-primary-content rounded-btn',
                day_today: 'font-bold',
                day: 'btn btn-ghost btn-sm rounded-btn',
                head_cell: 'text-sm font-medium opacity-70',
                caption_label: 'text-lg font-bold',
                nav_button: 'btn btn-ghost btn-sm btn-circle',
              }}
              components={{
                IconLeft: () => <ChevronLeft className="h-4 w-4" />,
                IconRight: () => <ChevronRight className="h-4 w-4" />,
              }}
            />
          </div>
        </div>
      )}
    </div>
  );
}
```

**Usage:**
```tsx
const [date, setDate] = useState<Date>();

<DatePicker
  value={date}
  onChange={setDate}
  name="date"
  required
  placeholder="Select date"
/>
```

### 3. Toast Notifications (Replace Alerts)

**Current pattern** uses `alert()` or DaisyUI alert components inline.

**Upgraded with Sonner:**

Create `react/components/Toaster.tsx`:
```tsx
import { Toaster as SonnerToaster, toast } from 'sonner';

export function Toaster() {
  return (
    <SonnerToaster
      position="top-right"
      toastOptions={{
        className: 'toast',
        style: {
          background: 'hsl(var(--b2))',
          color: 'hsl(var(--bc))',
          border: '1px solid hsl(var(--b3))',
        },
      }}
    />
  );
}

// Export toast functions for use throughout app
export { toast };
```

**Add to layout:**
```tsx
// In your React root or layout
import { Toaster } from './components/Toaster';

function App() {
  return (
    <>
      <Toaster />
      {/* rest of app */}
    </>
  );
}
```

**Usage in components:**
```tsx
import { toast } from '../components/Toaster';

// Success toast
toast.success('Reading marked complete!');

// Error toast
toast.error('Failed to save changes');

// Loading toast with promise
toast.promise(
  fetch('/api/reading/1/complete', { method: 'POST' }),
  {
    loading: 'Saving...',
    success: 'Reading completed!',
    error: 'Failed to complete reading',
  }
);

// Custom toast
toast('Reading due today', {
  description: 'Genesis 1-3',
  action: {
    label: 'View',
    onClick: () => window.location.href = '/dashboard',
  },
});
```

### 4. Data Fetching with React Query

**Current pattern** uses raw fetch with useEffect for loading states.

**Upgraded with TanStack Query:**

Create `react/lib/queryClient.ts`:
```tsx
import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 1 minute
      refetchOnWindowFocus: false,
    },
  },
});
```

Create `react/hooks/useApi.ts`:
```tsx
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getCsrfToken } from './useCsrf';
import { toast } from '../components/Toaster';

// Fetch wrapper with CSRF
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
      fetchApi(`/api/reading/${readingId}/complete`, { method: 'POST' }),
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

// Hook for fetching plan status (with polling)
export function usePlanStatus(planId: number, enabled: boolean) {
  return useQuery({
    queryKey: ['plan', planId, 'status'],
    queryFn: () => fetchApi<{ status: string }>(`/api/plans/${planId}/status`),
    refetchInterval: enabled ? 5000 : false, // Poll every 5s while processing
    enabled,
  });
}
```

**Usage:**
```tsx
import { useCompleteReading, useNotificationCount } from '../hooks/useApi';

function ReadingRow({ reading }: { reading: Reading }) {
  const { mutate: complete, isPending } = useCompleteReading();

  return (
    <button
      onClick={() => complete(reading.id)}
      disabled={isPending}
      className="btn btn-primary btn-sm"
    >
      {isPending ? (
        <span className="loading loading-spinner loading-xs" />
      ) : (
        <CheckIcon className="h-5 w-5" />
      )}
      Complete
    </button>
  );
}

function NotificationBell() {
  const { data, isLoading } = useNotificationCount();
  const count = data?.count ?? 0;

  return (
    <button className="btn btn-ghost btn-circle">
      <div className="indicator">
        <BellIcon className="h-5 w-5" />
        {count > 0 && (
          <span className="badge badge-sm badge-primary indicator-item">
            {count}
          </span>
        )}
      </div>
    </button>
  );
}
```

### 5. Form Handling with React Hook Form

**Current pattern** uses uncontrolled forms with FormData.

**Upgraded with react-hook-form:**

```tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { PasswordInput } from './PasswordInput';
import { toast } from './Toaster';
import { getCsrfToken } from '../hooks/useCsrf';

const signInSchema = z.object({
  email: z.string().email('Please enter a valid email'),
  password: z.string().min(1, 'Password is required'),
});

type SignInFormData = z.infer<typeof signInSchema>;

interface SignInFormProps {
  csrf: string;
}

export function SignInForm({ csrf }: SignInFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SignInFormData>({
    resolver: zodResolver(signInSchema),
  });

  const onSubmit = async (data: SignInFormData) => {
    try {
      const formData = new FormData();
      formData.append('email', data.email);
      formData.append('password', data.password);
      formData.append('_csrf', csrf);

      const response = await fetch('/auth/sign-in', {
        method: 'POST',
        body: formData,
        headers: {
          'X-CSRF-Token': getCsrfToken(),
        },
      });

      if (response.ok) {
        window.location.href = '/dashboard';
      } else if (response.status === 422) {
        toast.error('Invalid email or password');
      } else {
        toast.error('Sign in failed. Please try again.');
      }
    } catch {
      toast.error('Network error. Please check your connection.');
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-1">
        <label htmlFor="email" className="text-sm font-medium">
          Email
        </label>
        <input
          type="email"
          id="email"
          autoComplete="email"
          placeholder="email@example.com"
          className={`input input-bordered w-full ${errors.email ? 'input-error' : ''}`}
          {...register('email')}
        />
        {errors.email && (
          <p className="text-sm text-error">{errors.email.message}</p>
        )}
      </div>

      <PasswordInput
        id="password"
        autoComplete="current-password"
        placeholder="••••••••"
        label="Password"
        error={errors.password?.message}
        {...register('password')}
      />

      <button
        type="submit"
        disabled={isSubmitting}
        className="btn btn-primary w-full"
      >
        {isSubmitting && <span className="loading loading-spinner loading-sm" />}
        Sign In
      </button>
    </form>
  );
}
```

### 6. Icons with Lucide React

**Current implementation** uses inline SVGs in `icons.templ`.

**Upgraded with lucide-react:**

```tsx
// Instead of custom SVG icons, use lucide-react
import {
  Bell,
  Book,
  Calendar,
  Check,
  ChevronLeft,
  ChevronRight,
  Eye,
  EyeOff,
  Home,
  History,
  Menu,
  Pencil,
  Plus,
  Save,
  Settings,
  Trash2,
  Upload,
  User,
  X,
  AlertTriangle,
  Info,
  LogOut,
} from 'lucide-react';

// Icon mapping for easy replacement
export const Icons = {
  Bell,
  Book,
  Calendar,
  Check,
  ChevronLeft,
  ChevronRight,
  Eye,
  EyeOff,
  Dashboard: Home,
  History,
  Menu,
  Pencil,
  Plus,
  Save,
  Settings,
  Trash: Trash2,
  Upload,
  User,
  Close: X,
  Warning: AlertTriangle,
  Info,
  SignOut: LogOut,
};

// Usage
import { Icons } from './Icons';

<Icons.Bell className="h-5 w-5" />
<Icons.Check className="h-5 w-5" />
```

---

## UX Improvements

### 1. Loading Skeletons

Replace spinner-only loading states with skeleton screens:

```tsx
import Skeleton from 'react-loading-skeleton';
import 'react-loading-skeleton/dist/skeleton.css';

function ReadingCardSkeleton() {
  return (
    <div className="card bg-base-200 shadow-xl">
      <div className="card-body">
        <Skeleton height={24} width="60%" /> {/* Title */}
        <Skeleton height={16} count={2} /> {/* Content */}
        <div className="flex gap-2 mt-4">
          <Skeleton height={24} width={100} /> {/* Date badge */}
          <Skeleton height={24} width={120} /> {/* Status badge */}
        </div>
      </div>
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="space-y-4">
      <Skeleton height={36} width={200} /> {/* Heading */}
      <ReadingCardSkeleton />
      <ReadingCardSkeleton />
    </div>
  );
}
```

### 2. Optimistic Updates

Show changes immediately before server confirmation:

```tsx
function useOptimisticComplete() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (readingId: number) =>
      fetchApi(`/api/reading/${readingId}/complete`, { method: 'POST' }),

    // Optimistically update the UI
    onMutate: async (readingId) => {
      await queryClient.cancelQueries({ queryKey: ['readings'] });

      const previousReadings = queryClient.getQueryData(['readings']);

      queryClient.setQueryData(['readings'], (old: Reading[]) =>
        old.map((r) =>
          r.id === readingId ? { ...r, status: 'completed' } : r
        )
      );

      return { previousReadings };
    },

    // Rollback on error
    onError: (err, readingId, context) => {
      queryClient.setQueryData(['readings'], context?.previousReadings);
      toast.error('Failed to complete reading');
    },

    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['readings'] });
    },
  });
}
```

### 3. Keyboard Shortcuts

Add common keyboard shortcuts for power users:

```tsx
import { useHotkeys } from 'react-hotkeys-hook';

function Dashboard() {
  useHotkeys('mod+enter', () => {
    // Complete first pending reading
    const firstPending = readings.find(r => r.status === 'pending');
    if (firstPending) complete(firstPending.id);
  }, { enableOnFormTags: false });

  useHotkeys('/', (e) => {
    e.preventDefault();
    // Focus search (if you add one)
    document.getElementById('search')?.focus();
  });

  useHotkeys('g h', () => {
    window.location.href = '/dashboard';
  });

  useHotkeys('g p', () => {
    window.location.href = '/plans';
  });

  // ...
}
```

### 4. Error Boundaries

Gracefully handle React errors:

```tsx
import { Component, ReactNode } from 'react';

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<
  { children: ReactNode; fallback?: ReactNode },
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? (
        <div className="alert alert-error">
          <span>Something went wrong. Please refresh the page.</span>
        </div>
      );
    }
    return this.props.children;
  }
}

// Usage
<ErrorBoundary>
  <ManualPlanForm />
</ErrorBoundary>
```

### 5. Debounced Auto-Save

For the manual plan form, auto-save draft changes:

```tsx
import { useDebouncedCallback } from 'use-debounce';

function ManualPlanForm({ initialTitle }: { initialTitle: string }) {
  const [title, setTitle] = useState(initialTitle);

  const saveTitle = useDebouncedCallback(async (value: string) => {
    await fetch('/plans/draft/title', {
      method: 'POST',
      body: new URLSearchParams({ title: value }),
      headers: { 'X-CSRF-Token': getCsrfToken() },
    });
    toast.success('Draft saved', { duration: 2000 });
  }, 1000);

  return (
    <input
      type="text"
      value={title}
      onChange={(e) => {
        setTitle(e.target.value);
        saveTitle(e.target.value);
      }}
      placeholder="Enter plan title"
      className="input input-bordered w-full"
    />
  );
}
```

### 6. Confirmation Dialog with Undo

Instead of confirm dialogs, use toast with undo:

```tsx
function useDeleteWithUndo() {
  const queryClient = useQueryClient();

  const handleDelete = async (readingId: number) => {
    // Optimistically remove
    const previousData = queryClient.getQueryData(['readings']);
    queryClient.setQueryData(['readings'], (old: Reading[]) =>
      old.filter((r) => r.id !== readingId)
    );

    // Show toast with undo
    const toastId = toast('Reading deleted', {
      action: {
        label: 'Undo',
        onClick: () => {
          queryClient.setQueryData(['readings'], previousData);
          toast.dismiss(toastId);
        },
      },
      duration: 5000,
    });

    // Actually delete after toast duration
    setTimeout(async () => {
      try {
        await fetchApi(`/api/reading/${readingId}`, { method: 'DELETE' });
        queryClient.invalidateQueries({ queryKey: ['readings'] });
      } catch {
        queryClient.setQueryData(['readings'], previousData);
        toast.error('Failed to delete reading');
      }
    }, 5000);
  };

  return handleDelete;
}
```

### 7. Progress Indicators for Long Operations

Better feedback during CSV processing:

```tsx
function PlanProcessingCard({ planId }: { planId: number }) {
  const { data } = usePlanStatus(planId, true);
  const [progress, setProgress] = useState(0);

  // Animate progress smoothly
  useEffect(() => {
    if (data?.status === 'processing') {
      const timer = setInterval(() => {
        setProgress((p) => Math.min(p + Math.random() * 10, 90));
      }, 500);
      return () => clearInterval(timer);
    }
    if (data?.status === 'active') {
      setProgress(100);
    }
  }, [data?.status]);

  return (
    <div className="mt-4">
      <progress
        className="progress progress-primary w-full"
        value={progress}
        max={100}
      />
      <div className="flex justify-between text-sm opacity-70 mt-1">
        <span>Importing readings...</span>
        <span>{Math.round(progress)}%</span>
      </div>
    </div>
  );
}
```

### 8. Accessible Focus Management

Improve accessibility for keyboard users:

```tsx
import { useRef, useEffect } from 'react';

function ConfirmModal({ isOpen, onClose, onConfirm, title }: Props) {
  const confirmButtonRef = useRef<HTMLButtonElement>(null);
  const closeButtonRef = useRef<HTMLButtonElement>(null);

  // Focus trap and initial focus
  useEffect(() => {
    if (isOpen) {
      confirmButtonRef.current?.focus();

      // Trap focus within modal
      const handleTab = (e: KeyboardEvent) => {
        if (e.key === 'Tab') {
          if (e.shiftKey && document.activeElement === confirmButtonRef.current) {
            e.preventDefault();
            closeButtonRef.current?.focus();
          } else if (!e.shiftKey && document.activeElement === closeButtonRef.current) {
            e.preventDefault();
            confirmButtonRef.current?.focus();
          }
        }
        if (e.key === 'Escape') {
          onClose();
        }
      };

      document.addEventListener('keydown', handleTab);
      return () => document.removeEventListener('keydown', handleTab);
    }
  }, [isOpen, onClose]);

  // ...
}
```

---

## Package Additions

Update `package.json` with recommended libraries:

```json
{
  "dependencies": {
    "react": "^19.1.0",
    "react-dom": "^19.1.0",
    "@tanstack/react-query": "^5.80.6",
    "sonner": "^2.0.3",
    "react-hook-form": "^7.56.4",
    "lucide-react": "^0.511.0",
    "date-fns": "^4.1.0",
    "react-day-picker": "^9.7.0"
  },
  "devDependencies": {
    "@tailwindcss/cli": "4.1.18",
    "@tailwindcss/typography": "0.5.19",
    "@types/react": "^19.1.8",
    "@types/react-dom": "^19.1.6",
    "daisyui": "5.5.14",
    "tailwindcss": "4.1.18",
    "typescript": "^5.8.3",
    "zod": "^3.25.36",
    "@hookform/resolvers": "^5.0.1"
  }
}
```

### Optional Packages

```json
{
  "dependencies": {
    "react-loading-skeleton": "^3.5.0",
    "use-debounce": "^10.0.4",
    "react-hotkeys-hook": "^4.6.1"
  }
}
```

---

## Implementation Examples

### Complete Component: Manual Plan Form

Here's a fully upgraded version of the manual plan creation form:

```tsx
// react/components/ManualPlanForm.tsx
import { useState } from 'react';
import { useForm, useFieldArray } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Pencil, Check, X, Save } from 'lucide-react';
import { toast } from './Toaster';
import { DatePicker } from './DatePicker';
import { getCsrfToken } from '../hooks/useCsrf';

const readingSchema = z.object({
  id: z.string(),
  date: z.date({ required_error: 'Date is required' }),
  content: z.string().min(1, 'Reading content is required'),
});

const formSchema = z.object({
  title: z.string().min(1, 'Plan title is required'),
  readings: z.array(readingSchema).min(1, 'Add at least one reading'),
});

type FormData = z.infer<typeof formSchema>;

interface ManualPlanFormProps {
  initialTitle?: string;
  initialReadings?: Array<{ id: string; date: string; content: string }>;
}

export function ManualPlanForm({
  initialTitle = '',
  initialReadings = [],
}: ManualPlanFormProps) {
  const queryClient = useQueryClient();
  const [editingIndex, setEditingIndex] = useState<number | null>(null);

  const {
    register,
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
    reset,
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      title: initialTitle,
      readings: initialReadings.map((r) => ({
        ...r,
        date: new Date(r.date),
      })),
    },
  });

  const { fields, append, remove, update } = useFieldArray({
    control,
    name: 'readings',
  });

  // Add new reading
  const [newDate, setNewDate] = useState<Date>();
  const [newContent, setNewContent] = useState('');

  const addReading = () => {
    if (!newDate || !newContent.trim()) {
      toast.error('Please enter both date and content');
      return;
    }

    append({
      id: crypto.randomUUID(),
      date: newDate,
      content: newContent.trim(),
    });

    setNewDate(undefined);
    setNewContent('');
    toast.success('Reading added');
  };

  // Save draft
  const saveDraft = useMutation({
    mutationFn: async (data: FormData) => {
      const response = await fetch('/plans/draft', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getCsrfToken(),
        },
        body: JSON.stringify({
          title: data.title,
          readings: data.readings.map((r) => ({
            ...r,
            date: r.date.toISOString().split('T')[0],
          })),
        }),
      });
      if (!response.ok) throw new Error('Failed to save draft');
    },
    onSuccess: () => toast.success('Draft saved'),
    onError: () => toast.error('Failed to save draft'),
  });

  // Create plan
  const createPlan = useMutation({
    mutationFn: async (data: FormData) => {
      const response = await fetch('/plans/create-manual', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-CSRF-Token': getCsrfToken(),
        },
        body: JSON.stringify({
          title: data.title,
          readings: data.readings.map((r) => ({
            date: r.date.toISOString().split('T')[0],
            content: r.content,
          })),
        }),
      });
      if (!response.ok) throw new Error('Failed to create plan');
      return response.json();
    },
    onSuccess: () => {
      toast.success('Plan created successfully!');
      queryClient.invalidateQueries({ queryKey: ['plans'] });
      window.location.href = '/plans';
    },
    onError: () => toast.error('Failed to create plan'),
  });

  // Delete draft and go back
  const deleteDraft = async () => {
    await fetch('/plans/draft', {
      method: 'DELETE',
      headers: { 'X-CSRF-Token': getCsrfToken() },
    });
    window.location.href = '/plans';
  };

  return (
    <div className="card bg-base-200 shadow-xl">
      <form onSubmit={handleSubmit((data) => createPlan.mutate(data))}>
        <div className="card-body">
          {/* Title Input */}
          <div className="space-y-1 mb-6">
            <label htmlFor="plan-title" className="text-sm font-bold">
              Plan Title
            </label>
            <input
              type="text"
              id="plan-title"
              placeholder="Enter plan title"
              className={`input input-bordered w-full ${errors.title ? 'input-error' : ''}`}
              {...register('title')}
              onBlur={() => saveDraft.mutate(control._formValues as FormData)}
            />
            {errors.title && (
              <p className="text-sm text-error">{errors.title.message}</p>
            )}
          </div>

          <div className="divider my-2">Readings</div>

          {/* Add Reading Form */}
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr>
                  <th className="w-48">Date</th>
                  <th>Reading Content</th>
                  <th className="w-24">Actions</th>
                </tr>
              </thead>
              <tbody>
                {/* New reading row */}
                <tr className="bg-base-300">
                  <td>
                    <DatePicker
                      value={newDate}
                      onChange={setNewDate}
                      placeholder="Select date"
                    />
                  </td>
                  <td>
                    <input
                      type="text"
                      value={newContent}
                      onChange={(e) => setNewContent(e.target.value)}
                      placeholder="What to read..."
                      className="input input-bordered input-sm w-full"
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          addReading();
                        }
                      }}
                    />
                  </td>
                  <td>
                    <button
                      type="button"
                      onClick={addReading}
                      className="btn btn-primary btn-sm"
                    >
                      <Plus className="h-5 w-5" />
                    </button>
                  </td>
                </tr>

                {/* Existing readings */}
                {fields.map((field, index) => (
                  <ReadingRow
                    key={field.id}
                    reading={field}
                    isEditing={editingIndex === index}
                    onEdit={() => setEditingIndex(index)}
                    onSave={(updated) => {
                      update(index, updated);
                      setEditingIndex(null);
                      saveDraft.mutate(control._formValues as FormData);
                    }}
                    onCancel={() => setEditingIndex(null)}
                    onDelete={() => {
                      remove(index);
                      saveDraft.mutate(control._formValues as FormData);
                    }}
                  />
                ))}
              </tbody>
            </table>
          </div>

          {fields.length === 0 && (
            <div className="text-center py-8 text-base-content/60">
              No readings added yet. Add your first reading above.
            </div>
          )}

          {errors.readings && (
            <div className="alert alert-error mt-4">
              <span>{errors.readings.message}</span>
            </div>
          )}

          {/* Actions */}
          <div className="card-actions justify-end gap-2 mt-6">
            <button
              type="button"
              onClick={deleteDraft}
              className="btn btn-ghost"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={createPlan.isPending || fields.length === 0}
              className="btn btn-primary gap-2"
            >
              {createPlan.isPending ? (
                <span className="loading loading-spinner loading-sm" />
              ) : (
                <Plus className="h-5 w-5" />
              )}
              Create Plan
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}

// Reading row component
interface ReadingRowProps {
  reading: { id: string; date: Date; content: string };
  isEditing: boolean;
  onEdit: () => void;
  onSave: (reading: { id: string; date: Date; content: string }) => void;
  onCancel: () => void;
  onDelete: () => void;
}

function ReadingRow({
  reading,
  isEditing,
  onEdit,
  onSave,
  onCancel,
  onDelete,
}: ReadingRowProps) {
  const [editDate, setEditDate] = useState(reading.date);
  const [editContent, setEditContent] = useState(reading.content);

  if (isEditing) {
    return (
      <tr className="bg-base-300">
        <td>
          <DatePicker value={editDate} onChange={(d) => d && setEditDate(d)} />
        </td>
        <td>
          <input
            type="text"
            value={editContent}
            onChange={(e) => setEditContent(e.target.value)}
            className="input input-bordered input-sm w-full"
          />
        </td>
        <td>
          <div className="flex gap-1">
            <button
              type="button"
              onClick={() => onSave({ ...reading, date: editDate, content: editContent })}
              className="btn btn-success btn-xs"
              title="Save"
            >
              <Check className="h-4 w-4" />
            </button>
            <button
              type="button"
              onClick={onCancel}
              className="btn btn-ghost btn-xs"
              title="Cancel"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        </td>
      </tr>
    );
  }

  return (
    <tr>
      <td>
        <span className="font-medium">
          {reading.date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
          })}
        </span>
      </td>
      <td>
        <span>{reading.content}</span>
      </td>
      <td>
        <div className="flex gap-1">
          <button
            type="button"
            onClick={onEdit}
            className="btn btn-ghost btn-xs"
            title="Edit"
          >
            <Pencil className="h-4 w-4" />
          </button>
          <button
            type="button"
            onClick={onDelete}
            className="btn btn-ghost btn-xs text-error"
            title="Delete"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </td>
    </tr>
  );
}
```

---

## Migration Checklist for Upgrades

### Phase 1: Core Libraries
- [ ] Install `@tanstack/react-query`
- [ ] Set up QueryClientProvider
- [ ] Install and configure `sonner`
- [ ] Add Toaster component to layout

### Phase 2: Form Libraries
- [ ] Install `react-hook-form` and `zod`
- [ ] Create form schemas
- [ ] Convert auth forms
- [ ] Convert settings forms

### Phase 3: UI Components
- [ ] Install `lucide-react`
- [ ] Replace custom SVG icons
- [ ] Install `react-day-picker` and `date-fns`
- [ ] Replace cally calendar component

### Phase 4: UX Enhancements
- [ ] Add loading skeletons
- [ ] Implement optimistic updates
- [ ] Add keyboard shortcuts (optional)
- [ ] Implement error boundaries

### Phase 5: Testing & Cleanup
- [ ] Test all forms
- [ ] Test all mutations
- [ ] Remove unused dependencies
- [ ] Update documentation

---

## Summary

This upgrade guide recommends:

1. **Keep vanilla JS**: `push-setup.js` and `serviceWorker.js` must remain vanilla JavaScript
2. **Use React Query**: For all data fetching, caching, and mutations
3. **Use Sonner**: For toast notifications instead of alerts
4. **Use React Hook Form + Zod**: For form handling and validation
5. **Use Lucide React**: For consistent, tree-shakeable icons
6. **Use react-day-picker**: To replace the cally calendar component

The key principle is to upgrade incrementally, starting with the core infrastructure (React Query, Toaster) before moving to individual components. Each upgrade improves the user experience while maintaining compatibility with the existing DaisyUI styling.
