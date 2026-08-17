import { useState } from 'react';

interface FeatureProps {
  title: string;
  description: string;
  icon: string;
  details: string;
}

function FeatureModal({ title, description, icon, details }: FeatureProps) {
  const [isOpen, setIsOpen] = useState(false);

  const openModal = () => setIsOpen(true);
  const closeModal = () => setIsOpen(false);

  return (
    <>
      <div
        onClick={openModal}
        className="bg-gray-50 dark:bg-gray-800 p-6 rounded-lg shadow-md cursor-pointer hover:shadow-lg transition w-full max-w-sm"
      >
        <div className="text-4xl mb-4">{icon}</div>
        <h3 className="text-xl font-semibold mb-2 text-gray-800 dark:text-gray-100">{title}</h3>
        <p className="text-gray-600 dark:text-gray-300">{description}</p>
        <p className="text-blue-600 dark:text-blue-400 text-sm mt-2 font-medium">Click to learn more →</p>
      </div>

      {isOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-lg w-full p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-2xl font-bold text-blue-600 dark:text-blue-400">{icon} {title}</h2>
              <button onClick={closeModal} className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 text-2xl">×</button>
            </div>
            <p className="text-gray-700 dark:text-gray-300 mb-4">{details}</p>
            <button
              onClick={closeModal}
              className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </>
  );
}

export default FeatureModal;