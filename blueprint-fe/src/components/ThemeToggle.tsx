import { MoonOutlined, SunOutlined } from "@ant-design/icons";
import React from "react";
import { useTheme } from "../hooks/useTheme";

const ThemeToggle: React.FC = () => {
  const { isDark, toggleTheme } = useTheme();

  return (
    <button
      className="theme-toggle"
      onClick={toggleTheme}
      title={`Switch to ${isDark ? "light" : "dark"} mode`}
    >
      {isDark ? <SunOutlined /> : <MoonOutlined />}
    </button>
  );
};

export default ThemeToggle;
