import { createContext, useContext, useState, useEffect, ReactNode } from 'react';

type Language = 'en' | 'ha' | 'yo' | 'ig' | 'tiv';

interface LanguageContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  t: (key: string, params?: Record<string, string>) => string;
  languages: { code: Language; name: string }[];
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

const languageNames: Record<Language, string> = {
  en: 'English',
  ha: 'Hausa',
  yo: 'Yoruba',
  ig: 'Igbo',
  tiv: 'Tiv'
};

// Translation cache
let translationsCache: Record<Language, Record<string, string>> = {} as Record<Language, Record<string, string>>;

// Load translations dynamically
async function loadTranslations(lang: Language): Promise<Record<string, string>> {
  if (translationsCache[lang]) {
    return translationsCache[lang];
  }

  try {
    const module = await import(`../i18n/locales/${lang}.json`);
    translationsCache[lang] = module.default;
    return translationsCache[lang];
  } catch (e) {
    console.error(`Failed to load ${lang} translations`);
    // Fallback to English
    if (lang !== 'en') {
      return loadTranslations('en');
    }
    return {};
  }
}

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [language, setLanguage] = useState<Language>(() => {
    const saved = localStorage.getItem('language') as Language;
    return saved || 'en';
  });
  const [translations, setTranslations] = useState<Record<string, string>>({});
  const [loaded, setLoaded] = useState(false);

  useEffect(() => {
    const load = async () => {
      const data = await loadTranslations(language);
      setTranslations(data);
      setLoaded(true);
    };
    load();
  }, [language]);

  useEffect(() => {
    localStorage.setItem('language', language);
    document.documentElement.lang = language;
  }, [language]);

  const t = (key: string, params?: Record<string, string>): string => {
    if (!loaded) return key;
    
    let value = translations[key] || key;

    if (params && value) {
      return Object.entries(params).reduce(
        (str, [paramKey, paramValue]) => str.replace(`{${paramKey}}`, paramValue),
        value
      );
    }

    return value;
  };

  const languages = Object.entries(languageNames).map(([code, name]) => ({
    code: code as Language,
    name
  }));

  if (!loaded) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto"></div>
          <p className="mt-2 text-sm text-gray-500">Loading languages...</p>
        </div>
      </div>
    );
  }

  return (
    <LanguageContext.Provider value={{ language, setLanguage, t, languages }}>
      {children}
    </LanguageContext.Provider>
  );
}

export function useLanguage() {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
}