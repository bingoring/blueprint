import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ConfigProvider } from 'antd'
import koKR from 'antd/locale/ko_KR'
import enUS from 'antd/locale/en_US'
import './index.css'
import './i18n' // i18n 설정 import
import App from './App.tsx'

// Ant Design 언어 설정 (기본값: 한국어)
const getAntdLocale = () => {
  const language = localStorage.getItem('i18nextLng') || 'ko';
  return language === 'en' ? enUS : koKR;
};

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ConfigProvider
      locale={getAntdLocale()}
      theme={{
        token: {
          colorPrimary: '#1890ff',
          borderRadius: 6,
        },
      }}
    >
      <App />
    </ConfigProvider>
  </StrictMode>,
)
