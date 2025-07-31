import { useEffect, useState } from 'react';
import { useAuthStore } from '../stores/useAuthStore';
import { apiClient } from '../lib/api';
import GoalsPage from '../pages/GoalsPage';

export default function Dashboard() {
  const { user, logout } = useAuthStore();
  const [healthStatus, setHealthStatus] = useState<string>('checking...');
  const [currentPage, setCurrentPage] = useState<'dashboard' | 'goals'>('dashboard');

  useEffect(() => {
    // API 서버 연결 상태 확인
    const checkHealth = async () => {
      try {
        const health = await apiClient.healthCheck();
        setHealthStatus(`${health.status} - ${health.message}`);
             } catch {
         setHealthStatus('API 서버 연결 실패');
       }
    };

    checkHealth();
  }, []);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold text-gray-900">Blueprint</h1>
              <span className="ml-4 px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-sm">
                MVP Phase 1
              </span>
              <nav className="ml-8 flex space-x-4">
                <button
                  onClick={() => setCurrentPage('dashboard')}
                  className={`px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                    currentPage === 'dashboard'
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-600 hover:text-gray-900'
                  }`}
                >
                  대시보드
                </button>
                <button
                  onClick={() => setCurrentPage('goals')}
                  className={`px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                    currentPage === 'goals'
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-600 hover:text-gray-900'
                  }`}
                >
                  📋 내 목표
                </button>
              </nav>
            </div>

            <div className="flex items-center space-x-4">
              <div className="text-sm text-gray-600">
                안녕하세요, <span className="font-medium">{user?.username}</span>님!
              </div>
              <button
                onClick={logout}
                className="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-md text-sm transition duration-200"
              >
                로그아웃
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {currentPage === 'goals' ? (
          <GoalsPage />
        ) : (
        <div className="px-4 py-6 sm:px-0">
          {/* Welcome Section */}
          <div className="bg-white overflow-hidden shadow rounded-lg mb-8">
            <div className="px-4 py-5 sm:p-6">
              <h2 className="text-lg font-medium text-gray-900 mb-4">
                🚀 Blueprint에 오신 것을 환영합니다!
              </h2>
              <p className="text-gray-600 mb-4">
                당신의 꿈을 이룬 사람들이 직접 설계해주는 인생 로드맵 플랫폼입니다.
              </p>
              <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
                <h3 className="text-sm font-medium text-blue-800 mb-2">현재 개발 단계:</h3>
                <ul className="text-sm text-blue-700 space-y-1">
                  <li>✅ 사용자 인증 시스템</li>
                  <li>✅ 기본 대시보드</li>
                  <li>🚧 목표 설정 시스템 (개발 예정)</li>
                  <li>🚧 경로 제안 기능 (개발 예정)</li>
                  <li>🚧 예측 마켓 시스템 (개발 예정)</li>
                </ul>
              </div>
            </div>
          </div>

          {/* Status Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center">
                      <span className="text-white text-sm">✓</span>
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        인증 상태
                      </dt>
                      <dd className="text-lg font-medium text-gray-900">
                        로그인됨
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center">
                      <span className="text-white text-sm">🌐</span>
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        API 연결
                      </dt>
                      <dd className="text-sm font-medium text-gray-900">
                        {healthStatus}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-purple-500 rounded-full flex items-center justify-center">
                      <span className="text-white text-sm">👤</span>
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        사용자 유형
                      </dt>
                      <dd className="text-lg font-medium text-gray-900">
                        {user?.provider === 'google' ? 'Google' : 'Local'}
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Coming Soon Features */}
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">
                곧 출시될 기능들
              </h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900 mb-2">🎯 목표 설정</h4>
                  <p className="text-sm text-gray-600">
                    커리어, 창업, 교육 등 다양한 카테고리의 목표를 설정하고 관리할 수 있습니다.
                  </p>
                </div>
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900 mb-2">🛤️ 경로 제안</h4>
                  <p className="text-sm text-gray-600">
                    전문가들이 제안하는 다양한 성공 경로를 비교하고 선택할 수 있습니다.
                  </p>
                </div>
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900 mb-2">📈 예측 마켓</h4>
                  <p className="text-sm text-gray-600">
                    전문가들의 베팅을 통해 각 경로의 성공 확률을 실시간으로 확인합니다.
                  </p>
                </div>
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900 mb-2">🎓 멘토링</h4>
                  <p className="text-sm text-gray-600">
                    선택한 경로의 전문가와 1:1 멘토링을 통해 목표를 달성할 수 있습니다.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
        )}
      </main>
    </div>
  );
}
