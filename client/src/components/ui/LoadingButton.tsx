import React from 'react';

interface LoadingButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  loading?: boolean;
  children: React.ReactNode;
  variant?: 'primary' | 'secondary' | 'danger';
}

export const LoadingButton: React.FC<LoadingButtonProps> = ({
  loading = false,
  children,
  variant = 'primary',
  disabled,
  className = '',
  ...props
}) => {
  const isDisabled = disabled || loading;

  const getVariantClasses = () => {
    switch (variant) {
      case 'primary':
        return 'bg-accent-primary text-background-primary hover:bg-accent-primary-hover hover:-translate-y-px hover:shadow-lg';
      case 'secondary':
        return 'bg-background-secondary border-border text-text-primary hover:bg-background-tertiary hover:border-accent-primary hover:-translate-y-px';
      case 'danger':
        return 'bg-color-danger text-white hover:bg-color-danger-hover hover:-translate-y-px hover:shadow-lg';
      default:
        return 'bg-accent-primary text-background-primary hover:bg-accent-primary-hover hover:-translate-y-px hover:shadow-lg';
    }
  };

  return (
    <button
      {...props}
      disabled={isDisabled}
      className={`relative inline-flex items-center justify-center gap-2 px-4 py-2 rounded-md font-medium text-sm transition-all duration-150 border border-transparent min-h-9 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none ${getVariantClasses()} ${className}`}
    >
      {loading && (
        <div className="w-3.5 h-3.5 border-2 border-current border-t-transparent rounded-full animate-spin" />
      )}
      <span className={loading ? 'opacity-70' : ''}>{children}</span>
    </button>
  );
};