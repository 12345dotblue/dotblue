import React, { useEffect } from 'react';
import { casdoorService } from './CasdoorService';

const Login: React.FC = () => {
  useEffect(() => {
    // Redirect to Casdoor login page automatically
    window.location.href = casdoorService.getSigninUrl();
  }, []);

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
      <h2>Redirecting to Casdoor Login...</h2>
    </div>
  );
};

export default Login;
