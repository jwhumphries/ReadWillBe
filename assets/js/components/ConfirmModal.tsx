import React, { useState, useRef, useEffect, useImperativeHandle, forwardRef } from 'react';
import { getCsrfToken } from '../hooks/useCsrf';

interface ConfirmModalProps {
    id: string;
    title: string;
    message: string;
    confirmText: string;
    method: 'DELETE' | 'POST';
    url: string;
    onSuccess?: () => void;
    variant?: 'error' | 'primary' | 'warning';
}

export interface ConfirmModalRef {
    showModal: () => void;
    close: () => void;
}

export const ConfirmModal = forwardRef<ConfirmModalRef, ConfirmModalProps>(({
    id,
    title,
    message,
    confirmText,
    method,
    url,
    onSuccess,
    variant = 'error'
}, ref) => {
    const [loading, setLoading] = useState(false);
    const dialogRef = useRef<HTMLDialogElement>(null);

    useImperativeHandle(ref, () => ({
        showModal: () => dialogRef.current?.showModal(),
        close: () => dialogRef.current?.close(),
    }));

    const handleConfirm = async () => {
        setLoading(true);
        try {
            const response = await fetch(url, {
                method,
                headers: {
                    'X-CSRF-Token': getCsrfToken(),
                },
            });

            if (response.ok) {
                dialogRef.current?.close();
                if (onSuccess) {
                    onSuccess();
                } else {
                    window.location.reload();
                }
            }
        } catch (e) {
            console.error('Action failed', e);
        } finally {
            setLoading(false);
        }
    };

    // Expose showModal method via DOM element for templ integration
    useEffect(() => {
        const element = document.getElementById(id);
        if (element) {
            (element as HTMLDialogElement & { showModal: () => void }).showModal = () => dialogRef.current?.showModal();
        }
    }, [id]);

    const buttonClass = variant === 'error'
        ? 'btn btn-error'
        : variant === 'warning'
            ? 'btn btn-warning'
            : 'btn btn-primary';

    return (
        <dialog ref={dialogRef} id={id} className="modal">
            <div className="modal-box">
                <h3 className="font-bold text-lg">{title}</h3>
                <p className="py-4">{message}</p>
                <div className="modal-action">
                    <form method="dialog">
                        <button className="btn">Cancel</button>
                    </form>
                    <button
                        className={buttonClass}
                        onClick={handleConfirm}
                        disabled={loading}
                    >
                        {loading && <span className="loading loading-spinner loading-sm" />}
                        {confirmText}
                    </button>
                </div>
            </div>
            <form method="dialog" className="modal-backdrop">
                <button>close</button>
            </form>
        </dialog>
    );
});

ConfirmModal.displayName = 'ConfirmModal';
