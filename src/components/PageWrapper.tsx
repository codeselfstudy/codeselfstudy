import { ReactNode } from "react";

interface PageWrapperProps {
  children: ReactNode;
  className?: string;
}

export function PageWrapper({ children, className = "" }: PageWrapperProps) {
  return (
    <div className={`container mx-auto px-4 pt-24 pb-12 ${className}`}>
      {children}
    </div>
  );
}
