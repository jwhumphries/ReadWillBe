import React from 'react';
import { createRoot } from 'react-dom/client';

// Import components
import { PlanEditor } from './components/PlanEditor';
import { NotificationBell } from './components/NotificationBell';
import { ConfirmModal } from './components/ConfirmModal';
import { ReadingList } from './components/ReadingList';

// Component registry - maps component names to React components
const components: Record<string, React.ComponentType<any>> = {
    'PlanEditor': PlanEditor,
    'NotificationBell': NotificationBell,
    'ConfirmModal': ConfirmModal,
    'ReadingList': ReadingList,
};

// Mount all React components when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    const mounts = document.querySelectorAll('[data-react-component]');
    mounts.forEach(mount => {
        const componentName = mount.getAttribute('data-react-component');
        const propsJson = mount.getAttribute('data-props');

        if (componentName && components[componentName]) {
            const Component = components[componentName];
            const props = propsJson ? JSON.parse(propsJson) : {};
            const root = createRoot(mount);
            root.render(<Component {...props} />);
        } else if (componentName) {
            console.warn(`React component '${componentName}' not found in registry.`);
        }
    });
});
