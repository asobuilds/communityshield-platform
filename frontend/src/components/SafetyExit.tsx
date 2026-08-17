import { useState } from 'react';

function SafetyExit() {
  const [showConfirm, setShowConfirm] = useState(false);
  const path = window.location.pathname;
  const isAuthPage = path === '/login' || path === '/register';

  if (isAuthPage) return null;

  const handleExit = () => {
    window.location.href = 'https://www.google.com';
  };

  return (
    <div className="fixed bottom-4 right-4 z-50">
      <button
        onClick={() => setShowConfirm(true)}
        className="bg-gray-800 hover:bg-gray-700 text-white p-3 rounded-full shadow-lg transition-all duration-200 flex items-center justify-center"
        title="Safety Exit – hide this page quickly"
      >
        <span className="text-xl">🛡️</span>
      </button>

      {showConfirm && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl p-6 max-w-sm mx-4">
            <h3 className="text-lg font-semibold text-gray-800 mb-2">Safety Exit</h3>
            <p className="text-gray-600 mb-4">
              This will immediately hide this page and redirect you to a safe website.
              <br />
              <span className="text-xs text-gray-500">Your session will remain active.</span>
            </p>
            <div className="flex space-x-3">
              <button
                onClick={handleExit}
                className="flex-1 bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 font-medium"
              >
                🚪 Exit Now
              </button>
              <button
                onClick={() => setShowConfirm(false)}
                className="flex-1 bg-gray-300 text-gray-800 px-4 py-2 rounded hover:bg-gray-400 font-medium"
              >
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default SafetyExit;