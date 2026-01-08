import type { ReactNode } from "react";

interface PageWrapperProps {
  children: ReactNode;
  className?: string;
}

export function PageWrapper({ children, className = "" }: PageWrapperProps) {
  return (
    <div
      className={`container mx-auto px-4 pt-24 pb-12 sm:px-6 lg:px-8 ${className}`}
    >
      {children}
    </div>
  );
}
