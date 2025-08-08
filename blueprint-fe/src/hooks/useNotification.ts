import { App } from "antd";
import { useCallback } from "react";

export interface NotificationOptions {
  duration?: number;
  placement?:
    | "top"
    | "topLeft"
    | "topRight"
    | "bottom"
    | "bottomLeft"
    | "bottomRight";
}

export interface UseNotificationReturn {
  showSuccess: (message: string, options?: NotificationOptions) => void;
  showError: (message: string, options?: NotificationOptions) => void;
  showWarning: (message: string, options?: NotificationOptions) => void;
  showInfo: (message: string, options?: NotificationOptions) => void;
}

/**
 * 전역 알림 시스템을 위한 훅
 * Ant Design의 App 컨텍스트를 사용하여 일관된 알림 제공
 */
export const useNotification = (): UseNotificationReturn => {
  const { message } = App.useApp();

  const showSuccess = useCallback(
    (content: string, options?: NotificationOptions) => {
      message.success({
        content,
        duration: options?.duration || 3,
        style: { marginTop: "10vh" },
      });
    },
    [message]
  );

  const showError = useCallback(
    (content: string, options?: NotificationOptions) => {
      message.error({
        content,
        duration: options?.duration || 4,
        style: { marginTop: "10vh" },
      });
    },
    [message]
  );

  const showWarning = useCallback(
    (content: string, options?: NotificationOptions) => {
      message.warning({
        content,
        duration: options?.duration || 4,
        style: { marginTop: "10vh" },
      });
    },
    [message]
  );

  const showInfo = useCallback(
    (content: string, options?: NotificationOptions) => {
      message.info({
        content,
        duration: options?.duration || 3,
        style: { marginTop: "10vh" },
      });
    },
    [message]
  );

  return {
    showSuccess,
    showError,
    showWarning,
    showInfo,
  };
};
