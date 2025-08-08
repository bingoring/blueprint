import { ConfigProvider, theme } from "antd";
import React, { useEffect, useState } from "react";
import { ThemeContext, type Theme } from "./theme";

interface ThemeProviderProps {
  children: React.ReactNode;
}

export const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const [currentTheme, setCurrentTheme] = useState<Theme>(() => {
    const saved = localStorage.getItem("theme");
    return (saved as Theme) || "dark"; // 기본값을 dark로 설정
  });

  const toggleTheme = () => {
    const newTheme = currentTheme === "light" ? "dark" : "light";
    setCurrentTheme(newTheme);
    localStorage.setItem("theme", newTheme);
  };

  const isDark = currentTheme === "dark";

  // 시스템 테마에 따른 자동 감지 (옵션)
  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = (e: MediaQueryListEvent) => {
      if (!localStorage.getItem("theme")) {
        setCurrentTheme(e.matches ? "dark" : "light");
      }
    };

    mediaQuery.addEventListener("change", handleChange);
    return () => mediaQuery.removeEventListener("change", handleChange);
  }, []);

  // CSS 변수 설정
  useEffect(() => {
    const root = document.documentElement;

    if (isDark) {
      root.style.setProperty("--bg-primary", "#0a0e13");
      root.style.setProperty("--bg-secondary", "#141922");
      root.style.setProperty("--bg-tertiary", "#1f2329");
      root.style.setProperty("--text-primary", "#ffffff");
      root.style.setProperty("--text-secondary", "#8a8a8a");
      root.style.setProperty("--border-color", "#2a2a2a");
      root.style.setProperty("--green", "#00d395");
      root.style.setProperty("--red", "#ff4d6d");
      root.style.setProperty("--blue", "#1890ff");
      root.style.setProperty("--yellow", "#faad14");
    } else {
      root.style.setProperty("--bg-primary", "#ffffff");
      root.style.setProperty("--bg-secondary", "#fafafa");
      root.style.setProperty("--bg-tertiary", "#f5f5f5");
      root.style.setProperty("--text-primary", "#000000");
      root.style.setProperty("--text-secondary", "#666666");
      root.style.setProperty("--border-color", "#d9d9d9");
      root.style.setProperty("--green", "#52c41a");
      root.style.setProperty("--red", "#ff4d4f");
      root.style.setProperty("--blue", "#1890ff");
      root.style.setProperty("--yellow", "#faad14");
    }
  }, [isDark]);

  const antdTheme = {
    algorithm: isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
    token: {
      colorPrimary: "#1890ff",
      colorSuccess: isDark ? "#00d395" : "#52c41a",
      colorError: isDark ? "#ff4d6d" : "#ff4d4f",
      colorWarning: "#faad14",
      colorBgBase: isDark ? "#0a0e13" : "#ffffff",
      colorBgContainer: isDark ? "#141922" : "#ffffff",
      colorBgElevated: isDark ? "#1f2329" : "#ffffff",
      colorText: isDark ? "#ffffff" : "#000000",
      colorTextSecondary: isDark ? "#8a8a8a" : "#666666",
      colorBorder: isDark ? "#2a2a2a" : "#d9d9d9",
      borderRadius: 8,
    },
  };

  return (
    <ThemeContext.Provider value={{ theme: currentTheme, toggleTheme, isDark }}>
      <ConfigProvider theme={antdTheme}>{children}</ConfigProvider>
    </ThemeContext.Provider>
  );
};
