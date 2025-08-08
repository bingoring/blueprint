import { useCallback, useEffect, useState } from "react";

// Milestone 타입 정의 (validation용)
interface MilestoneForValidation {
  title?: string;
  [key: string]: unknown;
}

export interface ValidationRule<T = string> {
  validator: (value: T, context?: unknown) => boolean;
  message: string;
}

export interface UseValidationOptions<T = string> {
  rules: ValidationRule<T>[];
  context?: unknown;
}

export interface UseValidationReturn<T = string> {
  error: string | null;
  isValid: boolean;
  validate: (value: T) => boolean;
  clearError: () => void;
  setError: (message: string) => void;
}

/**
 * 재사용 가능한 validation 로직을 위한 훅
 * 다양한 validation 규칙을 조합하여 사용 가능
 */
export const useValidation = <T = string>(
  options: UseValidationOptions<T>
): UseValidationReturn<T> => {
  const [error, setErrorState] = useState<string | null>(null);

  const validate = useCallback(
    (value: T): boolean => {
      // 모든 validation 규칙을 순서대로 검사
      for (const rule of options.rules) {
        if (!rule.validator(value, options.context)) {
          setErrorState(rule.message);
          return false;
        }
      }

      setErrorState(null);
      return true;
    },
    [options.rules, options.context]
  );

  const clearError = useCallback(() => {
    setErrorState(null);
  }, []);

  const setError = useCallback((message: string) => {
    setErrorState(message);
  }, []);

  // context나 rules가 변경되면 현재 에러가 여전히 유효한지 확인
  useEffect(() => {
    if (error) {
      // context나 rules가 변경되었을 때 에러를 즉시 클리어하는 대신,
      // 다음 validation 시점에서 결과가 달라질 수 있음을 표시만 함
      // 실제 에러 클리어는 사용자의 직접적인 액션(입력, 블러 등)에서만 수행
    }
  }, [options.context, options.rules, error]);

  return {
    error,
    isValid: error === null,
    validate,
    clearError,
    setError,
  };
};

// 공통 validation 규칙들
export const ValidationRules = {
  /**
   * 빈 값 체크
   */
  required: (
    message: string = "값을 입력해주세요."
  ): ValidationRule<string> => ({
    validator: (value: string) => Boolean(value?.trim()),
    message,
  }),

  /**
   * 최소 길이 체크
   */
  minLength: (min: number, message?: string): ValidationRule<string> => ({
    validator: (value: string) => (value?.trim()?.length || 0) >= min,
    message: message || `최소 ${min}글자 이상 입력해주세요.`,
  }),

  /**
   * 최대 길이 체크
   */
  maxLength: (max: number, message?: string): ValidationRule<string> => ({
    validator: (value: string) => (value?.trim()?.length || 0) <= max,
    message: message || `최대 ${max}글자까지 입력 가능합니다.`,
  }),

  /**
   * 중복 체크 (배열 컨텍스트)
   */
  unique: (message?: string): ValidationRule<string> => ({
    validator: (value: string, context?: unknown) => {
      if (!context || !Array.isArray(context)) return true;
      const trimmedValue = value?.trim().toLowerCase();
      if (!trimmedValue) return true; // 빈 값은 unique 체크 안함

      return !context.some(
        (item: string) => item?.trim().toLowerCase() === trimmedValue
      );
    },
    message: message || "이미 존재하는 값입니다.",
  }),

  /**
   * 마일스톤 제목 중복 체크 (객체 배열 컨텍스트)
   */
  uniqueMilestoneTitle: (
    currentIndex: number,
    message?: string
  ): ValidationRule<string> => ({
    validator: (value: string, context?: unknown) => {
      if (!context || !Array.isArray(context)) return true;
      const trimmedValue = value?.trim().toLowerCase();

      // 현재 편집 중인 마일스톤 제외하고 중복 체크
      return !context.some(
        (milestone: MilestoneForValidation, index: number) =>
          index !== currentIndex &&
          milestone?.title?.trim().toLowerCase() === trimmedValue
      );
    },
    message: message || "이미 존재하는 마일스톤 제목입니다.",
  }),

  /**
   * 정규식 패턴 체크
   */
  pattern: (regex: RegExp, message: string): ValidationRule<string> => ({
    validator: (value: string) => regex.test(value?.trim() || ""),
    message,
  }),

  /**
   * 커스텀 validation 생성 헬퍼
   */
  custom: <T = string>(
    validator: (value: T, context?: unknown) => boolean,
    message: string
  ): ValidationRule<T> => ({
    validator,
    message,
  }),
};
