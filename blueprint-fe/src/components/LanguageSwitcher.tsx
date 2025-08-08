import React from 'react';
import { Select, Space } from 'antd';
import { GlobalOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';

const { Option } = Select;

interface LanguageSwitcherProps {
  size?: 'small' | 'middle' | 'large';
  showIcon?: boolean;
  style?: React.CSSProperties;
}

const LanguageSwitcher: React.FC<LanguageSwitcherProps> = ({
  size = 'middle',
  showIcon = true,
  style = {}
}) => {
  const { i18n } = useTranslation();

  const handleLanguageChange = (value: string) => {
    i18n.changeLanguage(value);
  };

  const languages = [
    { code: 'ko', name: '한국어', flag: '🇰🇷' },
    { code: 'en', name: 'English', flag: '🇺🇸' }
  ];

  return (
    <Space align="center" style={style}>
      {showIcon && <GlobalOutlined />}
      <Select
        value={i18n.language}
        onChange={handleLanguageChange}
        size={size}
        style={{ minWidth: 120 }}
        variant="borderless"
        suffixIcon={null}
      >
        {languages.map((lang) => (
          <Option key={lang.code} value={lang.code}>
            <Space>
              <span>{lang.flag}</span>
              <span>{lang.name}</span>
            </Space>
          </Option>
        ))}
      </Select>
    </Space>
  );
};

export default LanguageSwitcher;
