/**
 * 애플리케이션 전체에서 사용되는 메시지들을 중앙화하여 관리
 * 일관된 톤앤매너와 다국어 지원을 위한 기반 구조
 */

export const VALIDATION_MESSAGES = {
  // 공통 validation 메시지
  REQUIRED: "값을 입력해주세요.",
  REQUIRED_FIELD: (fieldName: string) => `${fieldName}을(를) 입력해주세요.`,

  // 길이 관련
  MIN_LENGTH: (min: number) => `최소 ${min}글자 이상 입력해주세요.`,
  MAX_LENGTH: (max: number) => `최대 ${max}글자까지 입력 가능합니다.`,

  // 중복 관련
  DUPLICATE: "이미 존재하는 값입니다.",
  DUPLICATE_OPTION: (value: string) => `"${value}" 옵션이 이미 존재합니다.`,

  // 베팅 옵션 관련
  BETTING_OPTION_REQUIRED: "옵션을 입력해주세요.",
  BETTING_OPTION_DUPLICATE: (option: string) =>
    `"${option}" 옵션이 이미 존재합니다.`,
  BETTING_OPTION_MIN_LENGTH: "옵션은 최소 2글자 이상이어야 합니다.",
  BETTING_OPTION_MAX_LENGTH: "옵션은 최대 50글자까지 가능합니다.",

  // 마일스톤 관련
  MILESTONE_TITLE_REQUIRED: "마일스톤 제목을 입력해주세요.",
  MILESTONE_TITLE_DUPLICATE: (title: string) =>
    `"${title}" 마일스톤이 이미 존재합니다.`,
  MILESTONE_TITLE_MIN_LENGTH: "마일스톤 제목은 최소 2글자 이상이어야 합니다.",
  MILESTONE_TITLE_MAX_LENGTH: "마일스톤 제목은 최대 100글자까지 가능합니다.",
} as const;

export const SUCCESS_MESSAGES = {
  // 성공 메시지
  OPTION_ADDED: "옵션이 추가되었습니다.",
  OPTION_REMOVED: "옵션이 삭제되었습니다.",
  PROJECT_CREATED: "프로젝트가 생성되었습니다.",
  PROJECT_UPDATED: "프로젝트가 업데이트되었습니다.",
} as const;

export const ERROR_MESSAGES = {
  // 에러 메시지
  UNEXPECTED_ERROR: "예상치 못한 오류가 발생했습니다.",
  NETWORK_ERROR: "네트워크 오류가 발생했습니다.",
  PROJECT_LOAD_FAILED: "프로젝트를 불러오는데 실패했습니다.",
  PROJECT_SAVE_FAILED: "프로젝트 저장에 실패했습니다.",

  // 권한 관련
  INVESTORS_EXISTS: "투자자가 있는 프로젝트는 수정할 수 없습니다.",
  BETTING_OPTIONS_LOCKED:
    "투자자가 있는 프로젝트는 투자 옵션을 변경할 수 없습니다.",
} as const;

export const WARNING_MESSAGES = {
  // 경고 메시지
  MAX_MILESTONES: "최대 5개의 마일스톤까지 추가할 수 있습니다.",
  DELETE_CONFIRMATION: "정말 삭제하시겠습니까?",
  UNSAVED_CHANGES: "저장되지 않은 변경사항이 있습니다.",
} as const;

export const INFO_MESSAGES = {
  // 정보 메시지
  LOADING: "로딩 중...",
  SAVING: "저장 중...",
  PROCESSING: "처리 중...",
} as const;

/**
 * 타입 안전한 메시지 접근을 위한 헬퍼 함수들
 */
export const MessageHelpers = {
  /**
   * 동적 메시지 생성 (타입 안전)
   */
  getDuplicateOptionMessage: (option: string): string =>
    VALIDATION_MESSAGES.BETTING_OPTION_DUPLICATE(option),

  getRequiredFieldMessage: (fieldName: string): string =>
    VALIDATION_MESSAGES.REQUIRED_FIELD(fieldName),

  getMinLengthMessage: (min: number): string =>
    VALIDATION_MESSAGES.MIN_LENGTH(min),

  getMaxLengthMessage: (max: number): string =>
    VALIDATION_MESSAGES.MAX_LENGTH(max),

  getDuplicateMilestoneMessage: (title: string): string =>
    VALIDATION_MESSAGES.MILESTONE_TITLE_DUPLICATE(title),
} as const;
