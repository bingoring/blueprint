import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import LanguageDetector from 'i18next-browser-languagedetector';

// 언어 리소스 import
import enTranslations from './locales/en.json';
import koTranslations from './locales/ko.json';

const resources = {
  en: {
    translation: enTranslations
  },
  ko: {
    translation: koTranslations
  }
};

i18n
  // 브라우저 언어 감지
  .use(LanguageDetector)
  // react-i18next 연결
  .use(initReactI18next)
  // 초기화
  .init({
    resources,

    // 기본 언어
    fallbackLng: 'ko',

    // 언어 감지 설정
    detection: {
      order: ['localStorage', 'navigator', 'htmlTag'],
      lookupLocalStorage: 'i18nextLng',
      caches: ['localStorage'],
    },

    interpolation: {
      escapeValue: false, // React가 XSS를 방지하므로 불필요
    },

    // 개발 모드에서 키 표시
    debug: import.meta.env.DEV,
  });

export default i18n;
