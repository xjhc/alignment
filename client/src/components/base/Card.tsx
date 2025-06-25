import React from 'react';

interface CardProps {
  children: React.ReactNode;
  variant?: 'default' | 'outlined' | 'elevated';
  padding?: 'none' | 'sm' | 'md' | 'lg';
  className?: string;
  onClick?: () => void;
  hoverable?: boolean;
}

interface CardHeaderProps {
  children: React.ReactNode;
  className?: string;
}

interface CardBodyProps {
  children: React.ReactNode;
  className?: string;
}

interface CardFooterProps {
  children: React.ReactNode;
  className?: string;
}

export const Card: React.FC<CardProps> & {
  Header: React.FC<CardHeaderProps>;
  Body: React.FC<CardBodyProps>;
  Footer: React.FC<CardFooterProps>;
} = ({
  children,
  variant = 'default',
  padding = 'md',
  className = '',
  onClick,
  hoverable = false,
  ...props
}) => {
  const baseClasses = 'bg-background-primary rounded-lg transition-all duration-150 w-full text-left focus-visible:outline-2 focus-visible:outline-primary focus-visible:outline-offset-2';
  
  const variantClasses = {
    default: 'border border-border',
    outlined: 'border border-border bg-transparent',
    elevated: 'border-none shadow-md',
  };

  const paddingClasses = {
    none: 'p-0',
    sm: 'p-3',
    md: 'p-4',
    lg: 'p-6',
  };

  const interactiveClasses = [
    hoverable || onClick ? 'hover:border-primary hover:shadow-lg' : '',
    onClick ? 'cursor-pointer border-none font-inherit text-inherit hover:-translate-y-0.5 hover:shadow-xl active:-translate-y-px active:shadow-lg' : '',
  ].filter(Boolean).join(' ');

  const classes = [
    baseClasses,
    variantClasses[variant],
    paddingClasses[padding],
    interactiveClasses,
    className,
  ].filter(Boolean).join(' ');

  const CardComponent = onClick ? 'button' : 'div';

  return (
    <CardComponent
      {...props}
      className={classes}
      onClick={onClick}
    >
      {children}
    </CardComponent>
  );
};

Card.Header = ({ children, className = '' }) => (
  <div className={`pb-3 border-b border-border mb-4 ${className}`}>
    {children}
  </div>
);

Card.Body = ({ children, className = '' }) => (
  <div className={`flex-1 ${className}`}>
    {children}
  </div>
);

Card.Footer = ({ children, className = '' }) => (
  <div className={`pt-3 border-t border-border mt-4 ${className}`}>
    {children}
  </div>
);