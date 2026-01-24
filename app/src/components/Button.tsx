import type { ButtonHTMLAttributes, ReactNode } from 'react'

type ButtonVariant = 'primary' | 'regular' | 'danger'

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: ButtonVariant
  children: ReactNode
}

const baseStyles =
  'inline-flex items-center justify-center rounded-md font-medium shadow-sm transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed'

const variantStyles: Record<ButtonVariant, string> = {
  primary:
    'text-teal-700 bg-white border border-teal-300 hover:bg-teal-50 focus:ring-teal-500',
  regular:
    'text-gray-700 bg-white border border-gray-300 hover:bg-gray-50 focus:ring-teal-500',
  danger:
    'text-red-700 bg-white border border-red-300 hover:bg-red-50 focus:ring-red-500',
}

export function Button({
  variant = 'regular',
  className,
  children,
  ...props
}: ButtonProps) {
  const classes = [baseStyles, variantStyles[variant], className]
    .filter(Boolean)
    .join(' ')

  return (
    <button className={classes} {...props}>
      {children}
    </button>
  )
}
