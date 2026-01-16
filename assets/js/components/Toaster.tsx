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

export { toast };
