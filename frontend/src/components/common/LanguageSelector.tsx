import { useLanguage } from '../../context/LanguageContext';

export function LanguageSelector() {
  const { language, setLanguage, languages } = useLanguage();

  return (
    <select
      value={language}
      onChange={(e) => setLanguage(e.target.value as any)}
      className="bg-gray-100 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
    >
      {languages.map((lang) => (
        <option key={lang.code} value={lang.code}>
          {lang.name}
        </option>
      ))}
    </select>
  );
}