import React, { useEffect } from 'react';
import { casdoorService } from './CasdoorService';

const Login: React.FC = () => {
  useEffect(() => {
    // Redirect to Casdoor login page automatically
    window.location.href = casdoorService.getSigninUrl();
  }, []);

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
        backgroundColor: '#f4f7fb',
        backgroundImage: "linear-gradient(rgba(255,255,255,0.82), rgba(255,255,255,0.92)), url('/brand/dotblue-login-bg.png')",
        backgroundSize: 'cover',
        backgroundPosition: 'center',
      }}
    >
      <div
        style={{
          width: '100%',
          maxWidth: 440,
          textAlign: 'center',
          padding: '40px 32px',
          borderRadius: 24,
          background: 'rgba(255, 255, 255, 0.78)',
          backdropFilter: 'blur(12px)',
          boxShadow: '0 24px 64px rgba(15, 52, 96, 0.12)',
        }}
      >
        <img
          src="/brand/dotblue-logo.png"
          alt="dotblue"
          style={{ width: 160, height: 56, objectFit: 'contain', marginBottom: 20 }}
        />
        <h2 style={{ margin: 0, color: '#0f172a', fontSize: 28 }}>Redirecting to Casdoor</h2>
        <p style={{ margin: '12px 0 0', color: '#475569', lineHeight: 1.7 }}>
          Preparing your secure login experience...
        </p>
      </div>
    </div>
  );
};

export default Login;
