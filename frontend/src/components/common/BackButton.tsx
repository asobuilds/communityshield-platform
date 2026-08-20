import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../../context/LanguageContext';

export function BackButton() {
  const navigate = useNavigate();
  const { t } = useLanguage();

  return (
    <button
      onClick={() => navigate(-1)}
      className="mb-4 inline-flex items-center gap-2 text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 transition"
    >
      ← {t('common.back')}
    </button>
  );
}