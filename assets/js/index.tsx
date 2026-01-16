import React from 'react';
import { createRoot } from 'react-dom/client';
import { QueryClientProvider } from '@tanstack/react-query';
import { queryClient } from './lib/queryClient';
import { Toaster } from './components/Toaster';

// Import components
import { PlanEditor } from './components/PlanEditor';
import { NotificationBell } from './components/NotificationBell';
import { ConfirmModal } from './components/ConfirmModal';
import { ReadingList } from './components/ReadingList';
import { PasswordInput } from './components/PasswordInput';
import { DatePicker } from './components/DatePicker';
import { ErrorBoundary } from './components/ErrorBoundary';
import { DashboardReadings } from './components/DashboardReadings';

// Component registry - maps component names to React components
const components: Record<string, React.ComponentType<any>> = {
    'PlanEditor': PlanEditor,
    'NotificationBell': NotificationBell,
    'ConfirmModal': ConfirmModal,
    'ReadingList': ReadingList,
    'PasswordInput': PasswordInput,
    'DatePicker': DatePicker,
    'ErrorBoundary': ErrorBoundary,
    'DashboardReadings': DashboardReadings,
};

// Wrapper component that provides global context
function AppWrapper({ children }: { children: React.ReactNode }) {
    return (
        <QueryClientProvider client={queryClient}>
            {children}
            <Toaster />
        </QueryClientProvider>
    );
}

// Mount all React components when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // First, mount the global providers (Toaster, etc.)
    const globalRoot = document.createElement('div');
    globalRoot.id = 'react-global-root';
    document.body.appendChild(globalRoot);
    const globalReactRoot = createRoot(globalRoot);
    globalReactRoot.render(
        <QueryClientProvider client={queryClient}>
            <Toaster />
        </QueryClientProvider>
    );

    // Then mount individual components
    const mounts = document.querySelectorAll('[data-react-component]');
    mounts.forEach(mount => {
        const componentName = mount.getAttribute('data-react-component');
        const propsJson = mount.getAttribute('data-props');

        if (componentName && components[componentName]) {
            const Component = components[componentName];
            const props = propsJson ? JSON.parse(propsJson) : {};
            const root = createRoot(mount);
            root.render(
                <QueryClientProvider client={queryClient}>
                    <Component {...props} />
                </QueryClientProvider>
            );
        } else if (componentName) {
            console.warn(`React component '${componentName}' not found in registry.`);
        }
    });
});
