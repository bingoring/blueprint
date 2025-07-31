import { useState } from 'react';
import { useAuthStore } from '../stores/useAuthStore';
import AuthPage from './AuthPage';
import Dashboard from './Dashboard';
import CreateDreamModal from './CreateDreamModal';

// 임시 모의 데이터
const mockGoals = [
  {
    id: 1,
    title: "3년 내 카페 사장되기",
    description: "서울 강남구에서 독립 카페 창업",
    category: "Business",
    probability: 72,
    totalStake: 1250000,
    timeLeft: "2년 3개월",
    creator: "김창업",
    participants: 23,
    trending: true
  },
  {
    id: 2,
    title: "2년 내 개발자 전직",
    description: "마케팅에서 풀스택 개발자로 커리어 전환",
    category: "Career",
    probability: 68,
    totalStake: 980000,
    timeLeft: "1년 8개월",
    creator: "이코딩",
    participants: 18,
    trending: false
  },
  {
    id: 3,
    title: "1년 내 토익 900점 달성",
    description: "현재 650점에서 900점으로 향상",
    category: "Education",
    probability: 85,
    totalStake: 450000,
    timeLeft: "11개월",
    creator: "박영어",
    participants: 32,
    trending: true
  },
  {
    id: 4,
    title: "5년 내 미국 이민",
    description: "소프트웨어 엔지니어로 실리콘밸리 이주",
    category: "Life",
    probability: 45,
    totalStake: 2100000,
    timeLeft: "4년 6개월",
    creator: "최이민",
    participants: 15,
    trending: false
  },
  {
    id: 5,
    title: "6개월 내 10kg 감량",
    description: "건강한 식단과 운동으로 체중 관리",
    category: "Health",
    probability: 78,
    totalStake: 320000,
    timeLeft: "5개월",
    creator: "정건강",
    participants: 41,
    trending: true
  },
  {
    id: 6,
    title: "2년 내 유튜브 구독자 10만명",
    description: "요리 컨텐츠로 인플루언서 되기",
    category: "Personal",
    probability: 35,
    totalStake: 750000,
    timeLeft: "1년 9개월",
    creator: "김요리",
    participants: 27,
    trending: false
  }
];

const categories = ["전체", "Career", "Business", "Education", "Life", "Health", "Personal"];

export default function HomePage() {
  const { isAuthenticated, user } = useAuthStore();
  const [showAuthModal, setShowAuthModal] = useState(false);
  const [showCreateDreamModal, setShowCreateDreamModal] = useState(false);
  const [currentView, setCurrentView] = useState<'home' | 'dashboard'>('home');
  const [selectedCategory, setSelectedCategory] = useState("전체");
  const [sortBy, setSortBy] = useState<"trending" | "probability" | "stake">("trending");

  // 대시보드 페이지를 보여주는 경우
  if (currentView === 'dashboard' && isAuthenticated) {
    return <Dashboard onNavigateHome={() => setCurrentView('home')} />;
  }

  const filteredGoals = mockGoals
    .filter(goal => selectedCategory === "전체" || goal.category === selectedCategory)
    .sort((a, b) => {
      switch (sortBy) {
        case "trending":
          return Number(b.trending) - Number(a.trending);
        case "probability":
          return b.probability - a.probability;
        case "stake":
          return b.totalStake - a.totalStake;
        default:
          return 0;
      }
    });

  const getCategoryColor = (category: string) => {
    const colors = {
      Career: "bg-blue-100 text-blue-800",
      Business: "bg-green-100 text-green-800",
      Education: "bg-purple-100 text-purple-800",
      Life: "bg-orange-100 text-orange-800",
      Health: "bg-red-100 text-red-800",
      Personal: "bg-pink-100 text-pink-800"
    };
    return colors[category as keyof typeof colors] || "bg-gray-100 text-gray-800";
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow-sm border-b sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <button
                onClick={() => setCurrentView('home')}
                className="text-2xl font-bold bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent hover:from-blue-700 hover:to-purple-700 transition-all duration-200"
              >
                LifePathDAO
              </button>
              <span className="ml-3 px-2 py-1 bg-gradient-to-r from-blue-100 to-purple-100 text-blue-800 rounded-full text-xs font-medium">
                Beta
              </span>
            </div>

            <div className="flex items-center space-x-4">
              {isAuthenticated ? (
                <>
                  <span className="text-sm text-gray-600">
                    안녕하세요, <span className="font-medium">{user?.username}</span>님!
                  </span>
                  <button
                    onClick={() => setShowCreateDreamModal(true)}
                    className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition duration-200"
                  >
                    ✨ 꿈 등록하기
                  </button>
                  <button
                    onClick={() => setCurrentView('dashboard')}
                    className="text-gray-600 hover:text-gray-900 px-3 py-2 text-sm font-medium border border-gray-300 rounded-lg hover:border-gray-400 transition duration-200"
                  >
                    📊 대시보드
                  </button>
                </>
              ) : (
                <>
                  <button
                    onClick={() => setShowAuthModal(true)}
                    className="text-gray-600 hover:text-gray-900 px-3 py-2 text-sm font-medium"
                  >
                    로그인
                  </button>
                                    <button
                    onClick={() => setShowCreateDreamModal(true)}
                    className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition duration-200"
                  >
                    ✨ 꿈 등록하기
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <div className="bg-gradient-to-br from-blue-50 via-purple-50 to-pink-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
          <div className="text-center">
            <h2 className="text-4xl font-bold text-gray-900 mb-4">
              당신의 꿈을 이룬 사람들이 설계하는
              <span className="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent"> 인생 로드맵</span>
            </h2>
            <p className="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">
              전문가들의 예측과 멘토링으로 목표 달성 확률을 높이고, 성공 시 보상을 받는 분산형 라이프 코칭 플랫폼
            </p>
            <div className="flex justify-center space-x-6 text-sm text-gray-500">
              <span>📊 실시간 성공 확률</span>
              <span>🎯 전문가 멘토링</span>
              <span>💰 성과 기반 보상</span>
              <span>🤝 커뮤니티 지원</span>
            </div>
          </div>
        </div>
      </div>

      {/* Stats Bar */}
      <div className="bg-white border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex justify-center space-x-8 text-sm">
            <div className="text-center">
              <div className="font-bold text-lg text-gray-900">1,247</div>
              <div className="text-gray-500">활성 목표</div>
            </div>
            <div className="text-center">
              <div className="font-bold text-lg text-gray-900">₩523M</div>
              <div className="text-gray-500">총 베팅금</div>
            </div>
            <div className="text-center">
              <div className="font-bold text-lg text-gray-900">73%</div>
              <div className="text-gray-500">평균 성공률</div>
            </div>
            <div className="text-center">
              <div className="font-bold text-lg text-gray-900">2,891</div>
              <div className="text-gray-500">등록 사용자</div>
            </div>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between bg-white rounded-lg shadow-sm p-4 mb-6">
          <div className="flex flex-wrap gap-2 mb-4 sm:mb-0">
            {categories.map((category) => (
              <button
                key={category}
                onClick={() => setSelectedCategory(category)}
                className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
                  selectedCategory === category
                    ? "bg-blue-600 text-white"
                    : "bg-gray-100 text-gray-600 hover:bg-gray-200"
                }`}
              >
                {category}
              </button>
            ))}
          </div>

          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-500">정렬:</span>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as "trending" | "probability" | "stake")}
              className="text-sm border border-gray-200 rounded-md px-3 py-1.5 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="trending">🔥 인기순</option>
              <option value="probability">📈 성공률순</option>
              <option value="stake">💰 베팅금순</option>
            </select>
          </div>
        </div>

        {/* Goal Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredGoals.map((goal) => (
            <div
              key={goal.id}
              className="bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow duration-200 p-6 cursor-pointer border border-gray-100"
            >
              <div className="flex items-start justify-between mb-3">
                <span className={`px-2 py-1 rounded-full text-xs font-medium ${getCategoryColor(goal.category)}`}>
                  {goal.category}
                </span>
                {goal.trending && (
                  <span className="bg-orange-100 text-orange-600 px-2 py-1 rounded-full text-xs font-medium flex items-center">
                    🔥 HOT
                  </span>
                )}
              </div>

              <h3 className="text-lg font-semibold text-gray-900 mb-2 line-clamp-2">
                {goal.title}
              </h3>
              <p className="text-sm text-gray-600 mb-4 line-clamp-2">
                {goal.description}
              </p>

              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-500">성공 확률</span>
                  <span className={`text-lg font-bold ${
                    goal.probability >= 70 ? "text-green-600" :
                    goal.probability >= 50 ? "text-yellow-600" : "text-red-600"
                  }`}>
                    {goal.probability}%
                  </span>
                </div>

                <div className="w-full bg-gray-200 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full ${
                      goal.probability >= 70 ? "bg-green-500" :
                      goal.probability >= 50 ? "bg-yellow-500" : "bg-red-500"
                    }`}
                    style={{ width: `${goal.probability}%` }}
                  ></div>
                </div>

                <div className="flex items-center justify-between text-sm text-gray-500">
                  <span>💰 {(goal.totalStake / 10000).toFixed(0)}만원</span>
                  <span>👥 {goal.participants}명 참여</span>
                </div>

                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-500">by {goal.creator}</span>
                  <span className="text-blue-600 font-medium">⏰ {goal.timeLeft}</span>
                </div>
              </div>

              <button className="w-full mt-4 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white py-2 rounded-lg text-sm font-medium transition duration-200">
                자세히 보기
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Auth Modal */}
      {showAuthModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg max-w-md w-full max-h-screen overflow-y-auto">
            <div className="p-6">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold">로그인 / 회원가입</h3>
                <button
                  onClick={() => setShowAuthModal(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  ✕
                </button>
              </div>
              <AuthPage />
            </div>
          </div>
        </div>
      )}

      {/* Create Dream Modal */}
      <CreateDreamModal
        isOpen={showCreateDreamModal}
        onClose={() => setShowCreateDreamModal(false)}
        onSuccess={(dream) => {
          console.log('꿈 등록 성공:', dream);
          // TODO: 성공 시 홈페이지 새로고침 또는 목록 업데이트
        }}
        onLoginRequired={() => {
          setShowCreateDreamModal(false);
          setShowAuthModal(true);
        }}
      />
    </div>
  );
}
