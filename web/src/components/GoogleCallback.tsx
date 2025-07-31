import { useEffect } from 'react';
import { useAuthStore } from '../stores/useAuthStore';

export default function GoogleCallback() {
  const { handleGoogleCallback } = useAuthStore();

  useEffect(() => {
    // URL에서 authorization code 추출
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');
    const error = urlParams.get('error');

    if (error) {
      // 에러가 있으면 부모 창에 에러 메시지 전송
      window.opener?.postMessage({
        type: 'GOOGLE_AUTH_ERROR',
        error: error,
      }, window.location.origin);
      window.close();
    } else if (code) {
      // authorization code가 있으면 처리
      handleGoogleCallback(code);
    } else {
      // 코드가 없으면 에러
      window.opener?.postMessage({
        type: 'GOOGLE_AUTH_ERROR',
        error: 'No authorization code received',
      }, window.location.origin);
      window.close();
    }
  }, [handleGoogleCallback]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="text-center">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p className="text-gray-600">Processing Google login...</p>
      </div>
    </div>
  );
}
