import React, { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { casdoorService } from './CasdoorService';
import { BACKEND_URL } from '../../config';

const PENDING_INVITE_CODE_KEY = 'pending_invite_code';

const LoginCallback: React.FC = () => {
  const navigate = useNavigate();
  const executed = useRef(false);  // Guard against StrictMode double-invoke

  useEffect(() => {
    if (executed.current) return;
    executed.current = true;

    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    const state = params.get('state');

    if (!code || !state) {
      console.error('Missing code or state in callback URL');
      navigate('/login');
      return;
    }

    // Exchange code+state on backend - keeps clientSecret server-side
    fetch(`${BACKEND_URL}/api/signin`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code, state }),
    })
      .then(res => {
        if (!res.ok) throw new Error(`Server returned ${res.status}`);
        return res.json();
      })
      .then((data: { token: string }) => {
        casdoorService.setToken(data.token);
        const pendingInviteCode = localStorage.getItem(PENDING_INVITE_CODE_KEY);
        if (pendingInviteCode) {
          localStorage.removeItem(PENDING_INVITE_CODE_KEY);
          navigate(`/invite/${pendingInviteCode}`, { replace: true });
          return;
        }
        navigate('/dashboard');
      })
      .catch(err => {
        console.error('Login failed:', err);
        navigate('/login');
      });
  }, [navigate]);

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
      <h2>Logging in...</h2>
    </div>
  );
};

export default LoginCallback;
