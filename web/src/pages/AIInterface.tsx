import React from 'react';
import { Outlet } from 'react-router-dom';
import AIMenu from '../components/layout/AIMenu';

/**
 * AI Interface main page component with nested routing
 */
const AIInterface: React.FC = () => {
  return (
    <div className="flex h-full w-full bg-gray-50 dark:bg-gray-900">
      {/* AI Feature Menu */}
      <AIMenu />

      {/* Main Content Area - renders child routes */}
      <Outlet />
    </div>
  );
};

export default AIInterface;
