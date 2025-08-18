import React from "react";

interface IconProps {
  size?: number;
  color?: string;
  className?: string;
  progress?: number;
}

// 나침반 아이콘 - 프로젝트 탐색
export const CompassIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <circle cx="12" cy="12" r="10" stroke={color} strokeWidth="2" />
    <polygon
      points="16.24,7.76 14.12,14.12 7.76,16.24 9.88,9.88"
      fill={color}
    />
    <circle cx="12" cy="12" r="2" fill={color} />
  </svg>
);

// 로켓 아이콘 - 새 프로젝트 시작
export const RocketIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <path
      d="M4.5 16.5c-1.5 1.5-1.5 3.5-1.5 3.5s2 0 3.5-1.5L8.5 16.5"
      stroke={color}
      strokeWidth="2"
    />
    <path
      d="M12 15l-3-3a22 22 0 0 1 2-3.95A12.88 12.88 0 0 1 22 2c0 2.72-.78 7.5-6 11a22.35 22.35 0 0 1-4 2z"
      stroke={color}
      strokeWidth="2"
      fill="none"
    />
    <path d="M9 12h4v4" stroke={color} strokeWidth="2" />
  </svg>
);

// 타겟 아이콘 - 마일스톤 (점)
export const MilestoneIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <circle cx="12" cy="12" r="10" stroke={color} strokeWidth="2" />
    <circle cx="12" cy="12" r="6" stroke={color} strokeWidth="2" />
    <circle cx="12" cy="12" r="2" fill={color} />
  </svg>
);

// 경로 아이콘 - 프로젝트 타임라인 (선)
export const PathIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <circle cx="5" cy="6" r="3" fill={color} />
    <circle cx="19" cy="18" r="3" fill={color} />
    <path
      d="M8 6h10l-5 6h5"
      stroke={color}
      strokeWidth="2"
      strokeLinecap="round"
    />
    <circle cx="12" cy="12" r="2" fill={color} />
  </svg>
);

// 대시보드 아이콘 - 홈
export const DashboardIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <rect
      x="3"
      y="3"
      width="7"
      height="7"
      stroke={color}
      strokeWidth="2"
      rx="1"
    />
    <rect
      x="14"
      y="3"
      width="7"
      height="7"
      stroke={color}
      strokeWidth="2"
      rx="1"
    />
    <rect
      x="14"
      y="14"
      width="7"
      height="7"
      stroke={color}
      strokeWidth="2"
      rx="1"
    />
    <rect
      x="3"
      y="14"
      width="7"
      height="7"
      stroke={color}
      strokeWidth="2"
      rx="1"
    />
  </svg>
);

// 포트폴리오 아이콘 - 내 활동
export const PortfolioIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <rect
      x="3"
      y="4"
      width="18"
      height="16"
      stroke={color}
      strokeWidth="2"
      rx="2"
    />
    <path d="M7 4v16" stroke={color} strokeWidth="2" />
    <path
      d="M17 8l-5 5-3-3"
      stroke={color}
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);

// 멘토링 아이콘 - 사람들 연결
export const MentoringIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <path
      d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"
      stroke={color}
      strokeWidth="2"
    />
    <circle cx="9" cy="7" r="4" stroke={color} strokeWidth="2" />
    <path d="M23 21v-2a4 4 0 0 0-3-3.87" stroke={color} strokeWidth="2" />
    <path d="M16 3.13a4 4 0 0 1 0 7.75" stroke={color} strokeWidth="2" />
  </svg>
);

// 트로피 아이콘 - 명예의 전당
export const TrophyIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <path d="M6 9H4.5a2.5 2.5 0 0 1 0-5H6" stroke={color} strokeWidth="2" />
    <path d="M18 9h1.5a2.5 2.5 0 0 0 0-5H18" stroke={color} strokeWidth="2" />
    <path d="M4 22h16" stroke={color} strokeWidth="2" />
    <path
      d="M10 14.66V17c0 .55.45 1 1 1h2c.55 0 1-.45 1-1v-2.34"
      stroke={color}
      strokeWidth="2"
    />
    <path d="M18 2H6v7a6 6 0 0 0 12 0V2Z" stroke={color} strokeWidth="2" />
  </svg>
);

// 연결선 아이콘 - Blueprint 핵심 아이덴티티
export const ConnectionIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <circle cx="6" cy="6" r="3" fill={color} />
    <circle cx="18" cy="6" r="3" fill={color} />
    <circle cx="6" cy="18" r="3" fill={color} />
    <circle cx="18" cy="18" r="3" fill={color} />
    <path d="M9 6h6M9 18h6M6 9v6M18 9v6" stroke={color} strokeWidth="2" />
  </svg>
);

// 투자 아이콘 - 차트 상승
export const InvestmentIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <path d="M3 3v18h18" stroke={color} strokeWidth="2" />
    <path
      d="M7 12l4-4 4 4 4-4"
      stroke={color}
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <circle cx="7" cy="12" r="2" fill={color} />
    <circle cx="11" cy="8" r="2" fill={color} />
    <circle cx="15" cy="12" r="2" fill={color} />
    <circle cx="19" cy="8" r="2" fill={color} />
  </svg>
);

// 프로그레스 아이콘 - 진행 상황
export const ProgressIcon: React.FC<IconProps> = ({
  size = 24,
  color = "currentColor",
  className = "",
  progress = 0.6,
}) => (
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="none"
    className={className}
  >
    <circle cx="12" cy="12" r="10" stroke="#e5e7eb" strokeWidth="2" />
    <circle
      cx="12"
      cy="12"
      r="10"
      stroke={color}
      strokeWidth="2"
      strokeDasharray={`${2 * Math.PI * 10 * progress} ${2 * Math.PI * 10}`}
      strokeDashoffset="0"
      transform="rotate(-90 12 12)"
      strokeLinecap="round"
    />
    <circle cx="12" cy="12" r="6" fill={color} fillOpacity="0.1" />
    <text x="12" y="16" textAnchor="middle" fontSize="8" fill={color}>
      {Math.round(progress * 100)}%
    </text>
  </svg>
);
