// 사용자 관련 타입
export interface User {
  id: string;
  email: string;
  username: string;
  provider: "local" | "google";
  googleId?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

// 목표 관련 타입
export interface Project {
  id: number;
  user_id: number;
  title: string;
  description: string;
  category: ProjectCategory;
  status: ProjectStatus;
  target_date: string | null;
  budget: number;
  priority: number;
  is_public: boolean;
  tags: string; // JSON string
  metrics: string;
  created_at: string;
  updated_at: string;
  milestones?: Milestone[]; // 프로젝트의 마일스톤들
}

export type ProjectCategory =
  | "career" // 💼 Career: 이직, 승진, 전직
  | "business" // 🚀 Business: 창업, 사업 확장
  | "education" // 📚 Education: 자격증, 학위, 스킬
  | "personal" // 🌱 Personal: 결혼, 건강, 취미
  | "life"; // 🏡 Life: 이민, 이사, 라이프스타일

export type ProjectStatus =
  | "draft" // 초안
  | "active" // 활성
  | "completed" // 완료
  | "cancelled" // 취소
  | "on_hold"; // 보류

// Project API 요청/응답 타입들
export interface CreateProjectRequest {
  title: string;
  description?: string;
  category: ProjectCategory;
  target_date?: string;
  budget?: number;
  priority?: number;
  is_public?: boolean;
  tags?: string[];
  metrics?: string;
}

export interface UpdateProjectRequest {
  title?: string;
  description?: string;
  category?: ProjectCategory;
  status?: ProjectStatus;
  target_date?: string;
  budget?: number;
  priority?: number;
  is_public?: boolean;
  tags?: string[];
  metrics?: string;
}

// 프로젝트와 마일스톤을 함께 생성하는 요청 ✨
export interface CreateProjectWithMilestonesRequest {
  title: string;
  description: string;
  category: ProjectCategory;
  target_date?: string;
  budget?: number;
  priority?: number;
  is_public?: boolean;
  tags?: string[];
  metrics?: string;
  milestones: CreateMilestoneRequest[];
}

// 마일스톤 생성 요청
export interface CreateMilestoneRequest {
  title: string;
  description?: string;
  order: number;
  target_date?: string;
}

// 마일스톤 업데이트 요청
export interface UpdateMilestoneRequest {
  title?: string;
  description?: string;
  status?: MilestoneStatus;
  target_date?: string;
  evidence?: string;
  notes?: string;
}

// AI 마일스톤 제안 관련 타입들 🤖
export interface AIMilestone {
  title: string;
  description: string;
  order: number;
  duration: string; // 예상 소요 기간
  difficulty: string; // 난이도 (쉬움/보통/어려움)
  category: string; // 카테고리 (준비/실행/완성)
}

export interface AIMilestoneResponse {
  milestones: AIMilestone[]; // 백엔드 호환성을 위해 milestones 유지
  tips: string[]; // 추가 팁
  warnings: string[]; // 주의사항
  usage: {
    // 사용 정보 추가
    remaining: number;
    total: number;
  };
  meta: {
    model: string;
    generated_at: string;
    user_id: number;
  };
}

export interface AIUsageInfo {
  used: number; // 사용한 횟수
  limit: number; // 최대 사용 가능 횟수
  remaining: number; // 남은 횟수
  can_use: boolean; // 사용 가능 여부
}

export interface ProjectCategoryOption {
  value: string;
  label: string;
  icon: string;
  description?: string;
}

export interface ProjectStatusOption {
  value: string;
  label: string;
  color: string;
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

// Legacy types (for backward compatibility)
export type Priority = "low" | "medium" | "high" | "urgent";

export interface Constraint {
  type: "budget" | "time" | "location" | "family" | "other";
  description: string;
  value?: string;
}

// 경로 관련 타입
export interface Path {
  id: string;
  projectId: string;
  title: string;
  description: string;
  estimatedDuration: number; // 개월 단위
  estimatedCost?: number;
  difficulty: "easy" | "medium" | "hard" | "expert";
  successProbability: number; // 0.0 - 1.0
  milestones: Milestone[];
  expertId: string;
  totalStake: number; // 베팅된 토큰 총량
  predictorCount: number; // 예측자 수
  createdAt: string;
  updatedAt: string;
}

export interface Milestone {
  id?: number; // 선택적으로 변경 (생성시에는 없음)
  project_id?: number; // 프로젝트에 직접 연결된 마일스톤
  path_id?: number; // 경로를 통한 마일스톤 (기존)
  title: string;
  description: string;
  order: number;
  target_date?: string; // 목표 날짜
  completed_at?: string;
  status?: MilestoneStatus; // 선택적으로 변경 (기본값 pending)
  is_completed?: boolean; // 선택적으로 변경 (기본값 false)
  total_support?: number; // 선택적으로 변경 (기본값 0)
  supporter_count?: number; // 선택적으로 변경 (기본값 0)
  success_probability?: number; // 선택적으로 변경 (기본값 0)
  evidence?: string; // 선택적으로 변경 (기본값 빈 JSON)
  notes?: string;
  email_sent?: boolean; // 선택적으로 변경 (기본값 false)
  reminder_sent?: boolean; // 선택적으로 변경 (기본값 false)
  created_at?: string; // 선택적으로 변경 (DB에서만 필요)
  updated_at?: string; // 선택적으로 변경 (DB에서만 필요)

  // 투자 관련 새 필드들
  betting_type?: "simple" | "custom";
  betting_options?: string[]; // 베팅 옵션 배열
}

export type MilestoneStatus = "pending" | "completed" | "failed" | "cancelled";

// 기존 Phase 타입도 호환성을 위해 유지
export type Phase = Milestone;
export type PhaseStatus = MilestoneStatus;

export interface Evidence {
  type: "image" | "document" | "link" | "text";
  content: string;
  description?: string;
  uploadedAt: string;
}

// 예측 마켓 관련 타입
export interface PathPrediction {
  id: string;
  pathId: string;
  userId: string;
  expertId: string;
  probability: number; // 0.0 - 1.0
  stakeAmount: number; // 토큰 단위
  reasoning: string;
  confidence: number; // 0.0 - 1.0
  createdAt: string;
}

export interface PredictionMarket {
  pathId: string;
  totalStake: number;
  averageProbability: number;
  predictions: PathPrediction[];
  topExperts: Expert[];
  priceHistory: PricePoint[];
}

export interface PricePoint {
  timestamp: string;
  probability: number;
  volume: number;
}

// 전문가 관련 타입
export interface Expert {
  id: string;
  userId: string;
  specialties: ProjectCategory[];
  experienceYears: number;
  successRate: number;
  totalMentees: number;
  rating: number;
  badges: Badge[];
  bio: string;
  isVerified: boolean;
  joinedAt: string;
}

export interface Badge {
  type: "expertise" | "achievement" | "contribution";
  name: string;
  description: string;
  iconUrl: string;
  earnedAt: string;
}

// 멘토링 관련 타입
export interface MentoringSession {
  id: string;
  pathId: string;
  mentorId: string;
  menteeId: string;
  milestoneId?: string;
  scheduledAt: string;
  duration: number; // 분 단위
  status: "scheduled" | "in_progress" | "completed" | "cancelled";
  notes?: string;
  rating?: number;
  feedback?: string;
  createdAt: string;
}

// API 응답 타입
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  hasNext: boolean;
  hasPrev: boolean;
}

// 인증 관련 타입들
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  username: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  user: User;
  message?: string;
}

// 로그아웃 응답 타입
export interface LogoutResponse {
  message: string;
  user_id: number;
  logout_time: string;
  instructions: string;
}

// 토큰 갱신 응답 타입
export interface RefreshTokenResponse {
  token: string;
  user: User;
  expires_in: number; // 초 단위
  refresh_time: string;
}

// 토큰 만료 확인 응답 타입
export interface TokenExpiryResponse {
  user_id: number;
  expiration_time: string;
  remaining_seconds: number;
  remaining_minutes: number;
  remaining_hours: number;
  is_expired: boolean;
  should_refresh: boolean;
  checked_at: string;
}

// 차트 및 통계 타입
export interface StatCard {
  title: string;
  value: string | number;
  change?: number;
  changeType?: "increase" | "decrease" | "neutral";
  icon?: string;
}

export interface ChartData {
  labels: string[];
  datasets: {
    label: string;
    data: number[];
    backgroundColor?: string | string[];
    borderColor?: string | string[];
    borderWidth?: number;
  }[];
}

// 폼 관련 타입
export interface ProjectFormData {
  title: string;
  description: string;
  category: ProjectCategory;
  targetDate: string;
  budget?: number;
  priority: Priority;
  constraints: Constraint[];
}

export interface PathFormData {
  title: string;
  description: string;
  estimatedDuration: number;
  estimatedCost?: number;
  difficulty: Path["difficulty"];
  milestones: Omit<Milestone, "id" | "pathId" | "createdAt" | "updatedAt">[];
}

// 상태 관리 타입
export interface AppState {
  user: User | null;
  isAuthenticated: boolean;
  currentProject: Project | null;
  selectedPath: Path | null;
  predictions: PathPrediction[];
  isLoading: boolean;
  error: string | null;
}

// 대시보드 관련 타입들
export interface ProjectTableRecord {
  id: number;
  title: string;
  category: ProjectCategory;
  status: ProjectStatus;
  progress: number;
  totalInvestment: number;
  investors: number;
  milestones: number;
  currentMilestone: number;
  createdAt: string;
  targetDate: string;
}

export interface InvestmentTableRecord {
  id: number;
  projectId: number;
  projectTitle: string;
  developer: string;
  amount: number;
  investedAt: string;
  status: "active" | "completed" | "cancelled";
  progress: number;
}

export interface ActivityRecord {
  id: number;
  type: "investment" | "milestone" | "project";
  title: string;
  description: string;
  time: string;
}

// 프로젝트 생성 관련 타입 - 이제 Milestone과 통합
export type ProjectMilestone = Milestone;

// 기존 타입도 호환성을 위해 유지
export type ProjectPhase = ProjectMilestone;
