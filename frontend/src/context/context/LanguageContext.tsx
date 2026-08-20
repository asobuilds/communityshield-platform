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

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [language, setLanguage] = useState<Language>(() => {
    const saved = localStorage.getItem('language') as Language;
    return saved || 'en';
  });

  useEffect(() => {
    localStorage.setItem('language', language);
    document.documentElement.lang = language;
  }, [language]);

  const t = (key: string, params?: Record<string, string>): string => {
    // Simple translation lookup - will be replaced with actual translations
    const translations: Record<Language, Record<string, string>> = {
      en: {
        'app.name': 'CommunityShield',
        'nav.home': 'Home',
        'nav.login': 'Login',
        'nav.register': 'Register',
        'nav.logout': 'Logout',
        'welcome': 'Welcome back, {name}!'
      },
      ha: {
        'app.name': 'CommunityShield',
        'nav.home': 'Gida',
        'nav.login': 'Shiga',
        'nav.register': 'Yi rajista',
        'nav.logout': 'Fita',
        'welcome': 'Sannu da dawowa, {name}!'
      },
      yo: {
        'app.name': 'CommunityShield',
        'nav.home': 'Ile',
        'nav.login': 'Wo ile',
        'nav.register': 'Forukọsilẹ',
        'nav.logout': 'Jade',
        'welcome': 'Kaabo pada, {name}!'
      },
      ig: {
        'app.name': 'CommunityShield',
        'nav.home': 'Ụlọ',
        'nav.login': 'Banye',
        'nav.register': 'Debanye aha',
        'nav.logout': 'Pụọ',
        'welcome': 'Nnọọ ọzọ, {name}!'
      },
      tiv: {
        'app.name': 'CommunityShield',
        'nav.home': 'Ngyô',
        'nav.login': 'Bende',
        'nav.register': 'Tsem',
        'nav.logout': 'Kem',
        'welcome': 'Kuma, {name}!'
      }
    };

    let value = translations[language]?.[key] || key;

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