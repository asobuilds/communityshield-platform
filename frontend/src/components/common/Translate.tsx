import { useLanguage } from '../../context/LanguageContext';

interface TranslateProps {
  text: string;
  params?: Record<string, string>;
  children?: (translated: string) => React.ReactNode;
}

export function Translate({ text, params, children }: TranslateProps) {
  const { t } = useLanguage();
  const translated = t(text, params);

  if (children) {
    return <>{children(translated)}</>;
  }

  return <>{translated}</>;
}

// Hook for using translations in components
export function useTranslate() {
  const { t } = useLanguage();
  return t;
}