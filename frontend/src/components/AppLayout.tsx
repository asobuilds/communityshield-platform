import { ReactNode } from 'react';
import Sidebar from './Sidebar';
import SafetyExit from './SafetyExit';

interface AppLayoutProps {
  children: ReactNode;
}

function AppLayout({ children }: AppLayoutProps) {
  return (
    <div className="flex min-h-screen bg-gray-100 dark:bg-gray-900 transition-colors duration-200">
      <Sidebar />
      <main className="flex-1 p-4 md:p-8 overflow-y-auto">
        {children}
        <SafetyExit />
      </main>
    </div>
  );
}

export default AppLayout;