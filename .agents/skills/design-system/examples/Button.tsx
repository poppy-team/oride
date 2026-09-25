import React from 'react';

/**
 * Design system button variants and sizes
 */
export type ButtonVariant = 'primary' | 'secondary' | 'outline' | 'ghost' | 'destructive';
export type ButtonSize = 'sm' | 'md' | 'lg';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  isLoading?: boolean;
  asChild?: boolean;
  children?: React.ReactNode;
}

/**
 * Accessible, token-driven, polymorphic Button component
 */
export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      variant = 'primary',
      size = 'md',
      isLoading = false,
      disabled = false,
      className = '',
      children,
      onClick,
      ...props
    },
    ref
  ) => {
    const isEffectivelyDisabled = disabled || isLoading;

    const baseClasses = 'inline-flex items-center justify-center font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 select-none';

    const variantStyles: Record<ButtonVariant, string> = {
      primary: 'bg-[var(--color-action-primary-bg,#2563eb)] text-[var(--color-action-primary-fg,#ffffff)] hover:bg-[var(--color-action-primary-bg-hover,#1d4ed8)] focus-visible:ring-[var(--color-focus-ring,#3b82f6)]',
      secondary: 'bg-[var(--color-action-secondary-bg,#f1f5f9)] text-[var(--color-action-secondary-fg,#0f172a)] hover:bg-[var(--color-action-secondary-bg-hover,#e2e8f0)] focus-visible:ring-[var(--color-focus-ring,#3b82f6)]',
      outline: 'border border-[var(--color-border-default,#cbd5e1)] bg-transparent text-[var(--color-fg-default,#0f172a)] hover:bg-[var(--color-bg-subtle,#f8fafc)] focus-visible:ring-[var(--color-focus-ring,#3b82f6)]',
      ghost: 'bg-transparent text-[var(--color-fg-default,#0f172a)] hover:bg-[var(--color-bg-subtle,#f8fafc)] focus-visible:ring-[var(--color-focus-ring,#3b82f6)]',
      destructive: 'bg-[var(--color-action-destructive-bg,#dc2626)] text-[var(--color-action-destructive-fg,#ffffff)] hover:bg-[var(--color-action-destructive-bg-hover,#b91c1c)] focus-visible:ring-[var(--color-action-destructive-bg,#dc2626)]',
    };

    const sizeStyles: Record<ButtonSize, string> = {
      sm: 'h-8 px-3 text-xs rounded-md gap-1.5 min-w-[32px]',
      md: 'h-10 px-4 text-sm rounded-md gap-2 min-w-[40px]',
      lg: 'h-12 px-6 text-base rounded-lg gap-2.5 min-w-[48px]',
    };

    const combinedClassName = [
      baseClasses,
      variantStyles[variant],
      sizeStyles[size],
      className,
    ].filter(Boolean).join(' ');

    const handleClick = (e: React.MouseEvent<HTMLButtonElement>) => {
      if (isEffectivelyDisabled) {
        e.preventDefault();
        return;
      }
      onClick?.(e);
    };

    return (
      <button
        ref={ref}
        type="button"
        disabled={disabled}
        aria-disabled={isEffectivelyDisabled}
        aria-busy={isLoading}
        className={combinedClassName}
        onClick={handleClick}
        {...props}
      >
        {isLoading && (
          <svg
            className="animate-spin -ml-1 mr-2 h-4 w-4 text-current"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        )}
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';
