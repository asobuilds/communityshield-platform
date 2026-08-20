import { Suspense, lazy, ComponentType } from 'react';

// Loading fallback component
export function LoadingFallback() {
  return (
    <div className="flex items-center justify-center min-h-[200px]">
      <div className="text-center">
        <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto"></div>
        <p className="mt-2 text-sm text-gray-500">Loading...</p>
      </div>
    </div>
  );
}

// Lazy load wrapper
export function lazyLoad<T extends ComponentType<any>>(
  importFn: () => Promise<{ default: T }>
) {
  const LazyComponent = lazy(importFn);
  return (props: any) => (
    <Suspense fallback={<LoadingFallback />}>
      <LazyComponent {...props} />
    </Suspense>
  );
}