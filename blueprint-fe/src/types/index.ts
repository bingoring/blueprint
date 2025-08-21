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

  // 프로필 관련 필드
  displayName?: string; // 표시 이름
  bio?: string; // 자기소개
  avatar?: string; // 프로필 이미지 URL

  // 평판 관련 필드
  projectSuccessRate?: number; // 프로젝트 성공률 (%)
  mentoringSuccessRate?: number; // 멘토링 성공률 (%)
  totalInvestment?: number; // 총 투자액 (USDC cents)
  sbtCount?: number; // 획득한 SBT 개수

  // 설정 관련 필드
  emailNotifications?: boolean; // 이메일 알림 수신 여부
  pushNotifications?: boolean; // 푸시 알림 수신 여부
  marketingNotifications?: boolean; // 마케팅 알림 수신 여부
  profilePublic?: boolean; // 프로필 공개 여부
  investmentPublic?: boolean; // 투자 내역 공개 여부

  // 프로필 정보 (백엔드에서 populate된 경우)
  profile?: {
    display_name?: string;
    avatar?: string;
    bio?: string;
  };
}

// 계정 설정: 프로필
export interface UserProfileSettings {
  id?: number;
  user_id?: number;
  display_name?: string;
  first_name?: string;
  last_name?: string;
  avatar?: string;
  bio?: string;
  age?: number;
  location?: string;
  website?: string;
  occupation?: string;
  experience?: string;
  skills?: string;
  interests?: string;
  capital?: number;
  constraints?: string;
  github_link?: string;
  linkedin_link?: string;
  twitter_link?: string;
  email_notifications?: boolean;
  push_notifications?: boolean;
  marketing_notifications?: boolean;
  profile_public?: boolean;
  investment_public?: boolean;
  created_at?: string;
  updated_at?: string;
}

// 프로필 업데이트 요청
export interface UpdateProfileRequest {
  display_name?: string;
  avatar?: string;
  bio?: string;
}

// 신원 증명 상태
export type VerificationStatus =
  | "unverified"
  | "pending"
  | "approved"
  | "rejected";

export interface UserVerificationStatus {
  id?: number;
  user_id?: number;
  // Level 1
  email_verified?: boolean;
  email_verified_at?: string | null;
  phone_verified?: boolean;
  phone_verified_at?: string | null;
  // Level 2 - Social & Career
  linkedin_connected?: boolean;
  linkedin_profile_id?: string | null;
  linkedin_profile_url?: string | null;
  linkedin_verified_at?: string | null;
  github_connected?: boolean;
  github_profile_id?: string | null;
  github_username?: string | null;
  github_verified_at?: string | null;
  twitter_connected?: boolean;
  twitter_profile_id?: string | null;
  twitter_username?: string | null;
  twitter_verified_at?: string | null;
  work_email_verified?: boolean;
  work_email_company?: string;
  work_email_verified_at?: string | null;
  // Level 3
  professional_status?: VerificationStatus;
  professional_title?: string;
  professional_doc_path?: string;
  professional_verified_at?: string | null;
  education_status?: VerificationStatus;
  education_degree?: string;
  education_doc_path?: string;
  education_verified_at?: string | null;
  created_at?: string;
  updated_at?: string;
}

export interface SettingsAggregateResponse {
  user: {
    id: number;
    email: string;
    username: string;
  };
  profile: UserProfileSettings | null;
  verification: UserVerificationStatus | null;
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
  tags: string[]; // 단순 string 배열로 변경
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

  // 🔍 증명 및 검증 관련 필드들
  requires_proof?: boolean; // 증거 제출 필요 여부 (기본값: true)
  proof_types?: ProofType[]; // 허용되는 증거 타입들
  min_validators?: number; // 최소 검증인 수 (기본값: 3)
  min_approval_rate?: number; // 최소 승인률 (기본값: 0.6)
  verification_deadline_days?: number; // 검증 마감일 (일수, 기본값: 3)
}

export type MilestoneStatus = "pending" | "completed" | "failed" | "cancelled";

// 🔍 증거 타입 정의
export type ProofType =
  | "file" // 파일 업로드 (이미지, PDF, 문서 등)
  | "url" // 웹 링크 (GitHub, 블로그, 포트폴리오 등)
  | "api" // API 연동 데이터 (GitHub, 헬스앱 등)
  | "text" // 텍스트 설명
  | "video" // 영상 업로드/링크
  | "screenshot" // 스크린샷
  | "certificate"; // 인증서/성적표

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

// 프로필 관련 타입들
export interface ProfileStats {
  projectSuccessRate: number; // 프로젝트 성공률
  mentoringSuccessRate: number; // 멘토링 성공률
  totalInvestment: number; // 총 투자액 (USDC cents)
  sbtCount: number; // SBT 개수
}

export interface CurrentProject {
  id: number;
  title: string;
  progress: number;
  category: string;
  status: string;
}

export interface FeaturedProject {
  id: number;
  title: string;
  description: string;
  status: string;
  investment: number; // 받은 투자액
  successRate: number; // 성공률
}

export interface RecentActivity {
  id: number;
  type: string; // investment, milestone, project 등
  description: string; // 활동 설명
  timestamp: string; // "2시간 전" 형태
}

export interface ProfileResponse {
  username: string;
  displayName: string;
  bio: string;
  avatar: string;
  joinedDate: string;
  stats: ProfileStats;
  currentProjects: CurrentProject[];
  featuredProjects: FeaturedProject[];
  recentActivities: RecentActivity[];
}

export interface ActivityLogParams {
  limit?: number;
  offset?: number;
  types?: string[];
  start_date?: string;
  end_date?: string;
}

export interface ActivityLogResponse {
  activities: RecentActivity[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    pages: number;
  };
}

export interface ActivitySummaryResponse {
  activity_counts: Array<{
    activity_type: string;
    count: number;
  }>;
  recent_activities: RecentActivity[];
  summary_period: string;
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

// 💰 투자 시스템 타입들
export interface Investment {
  id: number;
  user_id: number;
  project_id: number;
  milestone_id: number;
  bet_option: string;
  amount: number;
  shares: number;
  entry_price: number;
  status: "pending" | "active" | "settled" | "cancelled";
  result: "pending" | "win" | "lose" | "tie";
  payout: number;
  net_profit: number;
  fee: number;
  fee_rate: number;
  created_at: string;
  updated_at: string;
  user?: User;
  project?: Project;
  milestone?: Milestone;
}

export interface MarketData {
  id: number;
  milestone_id: number;
  option_id: string;
  current_price: number;
  previous_price: number;
  price_change_24h: number;
  volume_24h: number;
  volume_total: number;
  trade_count_24h: number;
  total_yes_shares: number;
  total_no_shares: number;
  liquidity_pool: number;
  buy_pressure: number;
  sell_pressure: number;
  max_price_24h: number;
  min_price_24h: number;
  avg_price_24h: number;
  is_active: boolean;
  last_trade_at?: string;
  created_at: string;
  updated_at: string;
}

// 🪙 화폐 타입
export type CurrencyType = "USDC" | "BLUEPRINT";

// 💰 하이브리드 지갑
export interface UserWallet {
  id: number;
  user_id: number;

  // 🔵 USDC 잔액 (베팅/보상용)
  usdc_balance: number; // 사용 가능한 USDC (센트 단위)
  usdc_locked_balance: number; // 베팅으로 잠긴 USDC

  // 🟦 BLUEPRINT 토큰 잔액 (거버넌스/스테이킹용)
  blueprint_balance: number; // 사용 가능한 BLUEPRINT
  blueprint_locked_balance: number; // 스테이킹/분쟁으로 잠긴 BLUEPRINT

  // 📊 통계 (USDC 기준)
  total_usdc_deposit: number; // 총 USDC 입금
  total_usdc_withdraw: number; // 총 USDC 출금
  total_usdc_profit: number; // 총 USDC 수익
  total_usdc_loss: number; // 총 USDC 손실
  total_usdc_fees: number; // 총 USDC 수수료

  // 📈 통계 (BLUEPRINT 기준)
  total_blueprint_earned: number; // 총 BLUEPRINT 획득
  total_blueprint_spent: number; // 총 BLUEPRINT 사용

  // 🎯 성과
  win_rate: number; // 승률
  total_trades: number; // 총 거래 수
  created_at: string;
  updated_at: string;
}

export interface PriceHistory {
  id: number;
  milestone_id: number;
  option_id: string;
  price: number;
  volume: number;
  timestamp: string;
}

// 투자 생성 요청 (주식 스타일)
export interface CreateInvestmentRequest {
  project_id: number;
  milestone_id: number;
  bet_option: string;
  shares: number; // 주식 수 (포인트 금액이 아닌)
}

// 투자 미리보기 요청 (주식 스타일)
export interface InvestmentPreviewRequest {
  project_id: number;
  milestone_id: number;
  bet_option: string;
  shares: number; // 주식 수
}

// 거래 영향 (주식 스타일)
export interface TradeImpact {
  current_price: number; // 현재 주당 가격
  new_price: number; // 거래 후 주당 가격
  price_impact: number; // 가격 영향도 (%)
  shares: number; // 구매할 주식 수
  total_cost: number; // 총 비용 (주식수 * 가격 + 수수료)
  fee: number; // 거래 수수료
  expected_payout: number; // 예상 수익
  roi_percentage: number; // 예상 ROI (%)
}

// 투자 미리보기 응답 (주식 스타일)
export interface InvestmentPreviewResponse {
  summary: TradeImpact;
}

export interface MarketStatusResponse {
  milestone: Milestone;
  market_data: MarketData[];
  viewer_count?: number;
  total_clients?: number;
  price_history: PriceHistory[];
  total_volume: number;
  total_trades: number;
  is_active: boolean;
}

// P2P 거래 관련 타입들 (폴리마켓 스타일)

// 주문 타입
export type OrderType = "market" | "limit";
export type OrderSide = "buy" | "sell";
export type OrderStatus =
  | "pending"
  | "partial"
  | "filled"
  | "cancelled"
  | "expired";

// 주문
export interface Order {
  id: number;
  user_id: number;
  project_id: number;
  milestone_id: number;
  option_id: string;
  type: OrderType;
  side: OrderSide;
  quantity: number;
  price: number;
  filled_quantity: number;
  avg_price: number;
  status: OrderStatus;
  created_at: string;
  updated_at: string;
}

// 거래 (Trade)
export interface Trade {
  id: number;
  project_id: number;
  milestone_id: number;
  option_id: string;
  buyer_id: number;
  seller_id: number;
  buy_order_id: number;
  sell_order_id: number;
  quantity: number;
  price: number;
  total_amount: number;
  buyer_fee: number;
  seller_fee: number;
  created_at: string;
}

// 포지션
export interface Position {
  id: number;
  user_id: number;
  project_id: number;
  milestone_id: number;
  option_id: string;
  quantity: number;
  avg_buy_price: number;
  total_invested: number;
  unrealized_pnl: number;
  realized_pnl: number;
  total_bought: number;
  total_sold: number;
  trade_count: number;
  created_at: string;
  updated_at: string;
}

// 호가창
export interface OrderBookLevel {
  price: number;
  quantity: number;
  orders: number;
}

export interface OrderBook {
  milestone_id: number;
  option_id: string;
  bids: OrderBookLevel[]; // 매수 호가 (높은 가격부터)
  asks: OrderBookLevel[]; // 매도 호가 (낮은 가격부터)
  spread: number; // 스프레드
  last_price: number; // 최종 거래가
  volume_24h: number; // 24시간 거래량
  timestamp: string;
}

// API 요청/응답

// 주문 생성 요청 (USDC 기준)
export interface CreateOrderRequest {
  project_id: number;
  milestone_id: number;
  option_id: string;
  type: OrderType;
  side: OrderSide;
  quantity: number; // 주식 수량
  price: number; // 확률 (0.01-0.99)
  currency: CurrencyType; // 화폐 타입 (항상 USDC)
}

// 주문 응답
export interface OrderResponse {
  order: Order;
  trades?: Trade[]; // 즉시 체결된 거래들
  position?: Position; // 업데이트된 포지션
  user_wallet: UserWallet; // 업데이트된 지갑
}

// 호가창 응답
export interface OrderBookResponse {
  order_book: OrderBook;
  success: boolean;
  message: string;
}
