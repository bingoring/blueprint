import {
  ClockCircleOutlined,
  ExclamationCircleOutlined,
} from "@ant-design/icons";
import { Card, Progress, Typography } from "antd";
import React, { useEffect, useState } from "react";
import type { TimeRemaining } from "../../types";

const { Title, Text } = Typography;

interface DisputeTimerProps {
  timeRemaining: TimeRemaining;
  phase: "challenge_window" | "voting_period";
  className?: string;
}

export const DisputeTimer: React.FC<DisputeTimerProps> = ({
  timeRemaining: initialTimeRemaining,
  phase,
  className = "",
}) => {
  const [timeRemaining, setTimeRemaining] = useState(initialTimeRemaining);

  useEffect(() => {
    if (timeRemaining.is_expired) return;

    const interval = setInterval(() => {
      setTimeRemaining((prev) => {
        const totalSeconds =
          prev.hours * 3600 + prev.minutes * 60 + prev.seconds - 1;

        if (totalSeconds <= 0) {
          return {
            ...prev,
            hours: 0,
            minutes: 0,
            seconds: 0,
            is_expired: true,
          };
        }

        const hours = Math.floor(totalSeconds / 3600);
        const minutes = Math.floor((totalSeconds % 3600) / 60);
        const seconds = totalSeconds % 60;

        return { ...prev, hours, minutes, seconds };
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [timeRemaining.is_expired]);

  const getPhaseTitle = () => {
    switch (phase) {
      case "challenge_window":
        return "이의 제기 대기 중";
      case "voting_period":
        return "판결 투표 진행 중";
      default:
        return "분쟁 처리 중";
    }
  };

  const getPhaseDescription = () => {
    switch (phase) {
      case "challenge_window":
        return "마일스톤 결과가 보고되었습니다. 투자자는 48시간 동안 이의를 제기할 수 있습니다.";
      case "voting_period":
        return "분쟁이 제기되었습니다. 판결단이 72시간 동안 투표를 진행합니다.";
      default:
        return "";
    }
  };

  const getProgressPercent = () => {
    const totalHours = phase === "challenge_window" ? 48 : 72;
    const remainingHours =
      timeRemaining.hours +
      timeRemaining.minutes / 60 +
      timeRemaining.seconds / 3600;
    return Math.max(
      0,
      Math.min(100, ((totalHours - remainingHours) / totalHours) * 100)
    );
  };

  const getStatusColor = () => {
    if (timeRemaining.is_expired) return "#ff4d4f";
    if (timeRemaining.hours < 12) return "#fa8c16";
    return "#52c41a";
  };

  if (timeRemaining.is_expired) {
    return (
      <Card
        className={`border border-red-500 ${className}`}
        style={{
          background: "var(--bg-card)",
          borderColor: "#ff4d4f",
        }}
      >
        <div className="text-center">
          <ExclamationCircleOutlined className="text-2xl text-red-500 mb-2" />
          <Title level={4} className="mb-2" style={{ color: "#ff4d4f" }}>
            기간 만료
          </Title>
          <Text style={{ color: "var(--text-secondary)" }}>
            {phase === "challenge_window"
              ? "이의 제기 기간이 종료되었습니다."
              : "투표 기간이 종료되었습니다."}
          </Text>
        </div>
      </Card>
    );
  }

  return (
    <Card
      className={`${className}`}
      style={{
        background: "var(--bg-card)",
        border: `2px solid ${getStatusColor()}`,
        borderRadius: "12px",
      }}
    >
      <div className="text-center space-y-4">
        {/* Phase Title */}
        <div className="flex items-center justify-center gap-2">
          <ClockCircleOutlined
            className="text-xl"
            style={{ color: getStatusColor() }}
          />
          <Title level={4} className="mb-0" style={{ color: getStatusColor() }}>
            {getPhaseTitle()}
          </Title>
        </div>

        {/* Description */}
        <Text style={{ color: "var(--text-secondary)" }}>
          {getPhaseDescription()}
        </Text>

        {/* Countdown */}
        <div
          className="bg-opacity-50 rounded-lg p-4"
          style={{ backgroundColor: getStatusColor() + "10" }}
        >
          <div className="grid grid-cols-3 gap-4 mb-3">
            <div className="text-center">
              <div
                className="text-2xl font-bold mb-1"
                style={{ color: getStatusColor() }}
              >
                {String(timeRemaining.hours).padStart(2, "0")}
              </div>
              <Text
                className="text-sm"
                style={{ color: "var(--text-secondary)" }}
              >
                시간
              </Text>
            </div>
            <div className="text-center">
              <div
                className="text-2xl font-bold mb-1"
                style={{ color: getStatusColor() }}
              >
                {String(timeRemaining.minutes).padStart(2, "0")}
              </div>
              <Text
                className="text-sm"
                style={{ color: "var(--text-secondary)" }}
              >
                분
              </Text>
            </div>
            <div className="text-center">
              <div
                className="text-2xl font-bold mb-1"
                style={{ color: getStatusColor() }}
              >
                {String(timeRemaining.seconds).padStart(2, "0")}
              </div>
              <Text
                className="text-sm"
                style={{ color: "var(--text-secondary)" }}
              >
                초
              </Text>
            </div>
          </div>

          {/* Progress Bar */}
          <Progress
            percent={getProgressPercent()}
            strokeColor={getStatusColor()}
            trailColor="var(--border-color)"
            showInfo={false}
            size="small"
          />

          <Text
            className="text-xs mt-1 block"
            style={{ color: "var(--text-secondary)" }}
          >
            {Math.round(getProgressPercent())}% 경과
          </Text>
        </div>

        {/* Warning for last 12 hours */}
        {timeRemaining.hours < 12 && (
          <div
            className="p-3 rounded-lg flex items-center gap-2"
            style={{ backgroundColor: "#fa8c16" + "10", color: "#fa8c16" }}
          >
            <ExclamationCircleOutlined />
            <Text style={{ color: "#fa8c16" }}>
              {phase === "challenge_window"
                ? "이의 제기 마감이 12시간 이내로 다가왔습니다!"
                : "투표 마감이 12시간 이내로 다가왔습니다!"}
            </Text>
          </div>
        )}
      </div>
    </Card>
  );
};

export default DisputeTimer;
