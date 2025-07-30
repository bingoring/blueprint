// 사용자 관련 타입
export interface User {
  id: string;
  email: string;
  username: string;
  provider: 'local' | 'google';
  googleId?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

// 목표 관련 타입
export interface Goal {
  id: string;
  userId: string;
  title: string;
  description: string;
  category: GoalCategory;
  targetDate: string;
  budget?: number;
  status: GoalStatus;
  priority: Priority;
  constraints: Constraint[];
  createdAt: string;
  updatedAt: string;
}

export type GoalCategory =
  | 'career'      // 💼 Career: 이직, 승진, 전직
  | 'business'    // 💰 Business: 창업, 사업 확장
  | 'education'   // 🎓 Education: 자격증, 학위, 스킬
  | 'personal'    // 💑 Personal: 결혼, 건강, 취미
  | 'lifestyle';  // 🏠 Life: 이민, 이사, 라이프스타일

export type GoalStatus =
  | 'draft'       // 초안
  | 'active'      // 진행중
  | 'completed'   // 완료
  | 'paused'      // 일시정지
  | 'failed';     // 실패

export type Priority = 'low' | 'medium' | 'high' | 'urgent';

export interface Constraint {
  type: 'budget' | 'time' | 'location' | 'family' | 'other';
  description: string;
  value?: string;
}

// 경로 관련 타입
export interface Path {
  id: string;
  goalId: string;
  title: string;
  description: string;
  estimatedDuration: number; // 개월 단위
  estimatedCost?: number;
  difficulty: 'easy' | 'medium' | 'hard' | 'expert';
  successProbability: number; // 0.0 - 1.0
  milestones: Milestone[];
  expertId: string;
  totalStake: number; // 베팅된 토큰 총량
  predictorCount: number; // 예측자 수
  createdAt: string;
  updatedAt: string;
}

export interface Milestone {
  id: string;
  pathId: string;
  title: string;
  description: string;
  order: number;
  dueDate: string;
  isCompleted: boolean;
  completedAt?: string;
  evidence: Evidence[];
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Evidence {
  type: 'image' | 'document' | 'link' | 'text';
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
  specialties: GoalCategory[];
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
  type: 'expertise' | 'achievement' | 'contribution';
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
  status: 'scheduled' | 'in_progress' | 'completed' | 'cancelled';
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

// 인증 관련 타입
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
  user: User;
  token: string;
  refreshToken: string;
}

// 차트 및 통계 타입
export interface StatCard {
  title: string;
  value: string | number;
  change?: number;
  changeType?: 'increase' | 'decrease' | 'neutral';
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
export interface GoalFormData {
  title: string;
  description: string;
  category: GoalCategory;
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
  difficulty: Path['difficulty'];
  milestones: Omit<Milestone, 'id' | 'pathId' | 'createdAt' | 'updatedAt'>[];
}

// 상태 관리 타입
export interface AppState {
  user: User | null;
  isAuthenticated: boolean;
  currentGoal: Goal | null;
  selectedPath: Path | null;
  predictions: PathPrediction[];
  isLoading: boolean;
  error: string | null;
}
