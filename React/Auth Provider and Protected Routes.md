---
id: Auth_Provider_and_Protected_Routes
aliases:
  - Auth Provider
  - Protected Routes
  - Role-Based Access
tags:
  - coding
  - react
  - typescript
  - authentication
dg-publish: true
---

# Auth Provider & Route Protection Pattern

> A comprehensive architectural guide to building a centralized `AuthProvider` in React + TypeScript, managing guest vs authenticated users, and implementing clean route protection.

### 🔗 Related Vault Notes
- [[React Cheat Sheet]]: Context API, Hooks, and Component Lifecycle
- [[TypeScript Cheat Sheet]]: Typing interfaces, context, and generic props
- [[Integrating JWT]]: Low-level JWT storage, decoding, and login/signup handlers
- [[Tokens]]: Access tokens vs refresh tokens
- [[Session Cookie Authenticaion]]: Cookie-based vs token-based session management

---

## 1. The Core Concepts

When designing web applications with authentication, pages and features fall into three categories:

| Route Type | Target Audience | Example Paths | Behavior |
| :--- | :--- | :--- | :--- |
| **Public Route** | Anyone (Guest + User) | `/`, `/about`, `/docs`, `/pricing` | Always accessible. May show dynamic navbar (Login vs Profile). |
| **Guest-Only Route** | Non-logged-in users | `/login`, `/register`, `/forgot-password` | If user is already logged in, redirect to `/dashboard`. |
| **Protected Route** | Authenticated users | `/dashboard`, `/profile`, `/settings` | If user is not logged in, redirect to `/login` (saving intended destination). |

---

## 2. Step 1: Building `AuthContext` & `AuthProvider`

Create a centralized `AuthContext.tsx` that manages authentication state, verifies existing tokens on initial load, and provides `login` and `logout` actions.

```tsx
// src/context/AuthContext.tsx
import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';

// 1. Define User and Context Interfaces
export interface User {
  id: string;
  name: string;
  email: string;
  role?: 'user' | 'admin';
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean; // Crucial: prevents flash of redirect before token check finishes
  login: (token: string, userData: User) => void;
  logout: () => void;
}

// 2. Create Context with undefined default
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// 3. Provider Component
export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Initialize Auth State from storage on app startup
  useEffect(() => {
    const initializeAuth = async () => {
      try {
        const storedToken = localStorage.getItem('token');
        const storedUser = localStorage.getItem('currentUser');

        if (storedToken && storedUser) {
          // Optional: Verify token expiration or call /api/me here
          setToken(storedToken);
          setUser(JSON.parse(storedUser));
        }
      } catch (error) {
        console.error('Failed to restore auth session:', error);
        localStorage.removeItem('token');
        localStorage.removeItem('currentUser');
      } finally {
        setIsLoading(false); // Done loading token
      }
    };

    initializeAuth();
  }, []);

  const login = (newToken: string, newUser: User) => {
    setToken(newToken);
    setUser(newUser);
    localStorage.setItem('token', newToken);
    localStorage.setItem('currentUser', JSON.stringify(newUser));
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    localStorage.removeItem('token');
    localStorage.removeItem('currentUser');
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isAuthenticated: !!user,
        isLoading,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

// 4. Custom Hook for consuming AuthContext
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
```

> [!IMPORTANT]
> **Why `isLoading` is vital:**
> When the user refreshes `/dashboard`, React mounts with `user = null` before the `useEffect` reads `localStorage`. Without `isLoading`, your protected route would falsely assume the user isn't logged in and redirect them to `/login`.

---

## 3. Step 2: Route Guards (React Router v6)

### A. Protected Route Guard (For Logged-in Users Only)
```tsx
// src/components/routes/ProtectedRoute.tsx
import React from 'react';
import { Navigate, useLocation, Outlet } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';

export const ProtectedRoute = () => {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return <div className="loading-spinner">Loading session...</div>;
  }

  if (!isAuthenticated) {
    // Redirect to /login and save current location so user can return after logging in
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  // Renders the child route component
  return <Outlet />;
};
```

### B. Guest-Only Route Guard (For Login / Signup Pages)
```tsx
// src/components/routes/GuestRoute.tsx
import React from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';

export const GuestRoute = () => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <div className="loading-spinner">Loading...</div>;
  }

  // Already logged in? Redirect to dashboard
  if (isAuthenticated) {
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
};
```

---

## 4. Step 3: Wiring It Together in App Router

```tsx
// src/App.tsx
import React from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { ProtectedRoute } from './components/routes/ProtectedRoute';
import { GuestRoute } from './components/routes/GuestRoute';

// Pages
import HomePage from './pages/HomePage';
import AboutPage from './pages/AboutPage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import DashboardPage from './pages/DashboardPage';
import ProfilePage from './pages/ProfilePage';
import Navbar from './components/Navbar';

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Navbar />
        <main>
          <Routes>
            {/* 1. PUBLIC ROUTES: Accessible to everyone */}
            <Route path="/" element={<HomePage />} />
            <Route path="/about" element={<AboutPage />} />

            {/* 2. GUEST-ONLY ROUTES: Logged-in users are redirected out */}
            <Route element={<GuestRoute />}>
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
            </Route>

            {/* 3. PROTECTED ROUTES: Guests are redirected to /login */}
            <Route element={<ProtectedRoute />}>
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/profile" element={<ProfilePage />} />
            </Route>
          </Routes>
        </main>
      </BrowserRouter>
    </AuthProvider>
  );
}
```

---

## 5. Allowing Non-Logged-In Users: UI Patterns

### Pattern 1: Dynamic Navigation Bar
Allow visitors to browse freely while showing conditional actions:

```tsx
// src/components/Navbar.tsx
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Navbar() {
  const { user, isAuthenticated, logout } = useAuth();

  return (
    <nav className="nav-container">
      <Link to="/" className="brand">MyApp</Link>
      
      {/* Public Links */}
      <div className="nav-links">
        <Link to="/">Home</Link>
        <Link to="/about">About</Link>

        {isAuthenticated ? (
          <>
            <Link to="/dashboard">Dashboard</Link>
            <span>Welcome, {user?.name}</span>
            <button onClick={logout}>Logout</button>
          </>
        ) : (
          <>
            <Link to="/login">Sign In</Link>
            <Link to="/register" className="btn-cta">Get Started</Link>
          </>
        )}
      </div>
    </nav>
  );
}
```

### Pattern 2: "Soft Wall" / Gatekeeper Action
Non-logged-in users can browse products or read articles, but clicking "Add to Wishlist" or "Comment" prompts a sign-in modal or redirects them:

```tsx
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export const ArticlePage = () => {
  const { isAuthenticated } = useAuth();
  const navigate = useNavigate();

  const handlePostComment = () => {
    if (!isAuthenticated) {
      // Option A: Redirect to login with return URL
      navigate('/login', { state: { from: window.location.pathname } });
      return;
    }

    // Option B: Open auth modal or submit comment
    submitComment();
  };

  return (
    <div>
      <h1>Public Article Content</h1>
      <p>Anyone can read this article without logging in...</p>

      <button onClick={handlePostComment}>
        {isAuthenticated ? "Post Comment" : "Sign in to Comment"}
      </button>
    </div>
  );
};
```

---

## 6. Returning Users to Where They Left Off

When a guest attempts to view `/dashboard/reports` and is redirected to `/login`, redirect them back after a successful login using `location.state`:

```tsx
// src/pages/LoginPage.tsx
import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  // Retrieve origin URL or default to dashboard
  const from = (location.state as any)?.from?.pathname || '/dashboard';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      });
      const data = await response.json();

      login(data.token, data.user);
      // Redirect user to the original protected page they wanted to visit
      navigate(from, { replace: true });
    } catch (err) {
      console.error(err);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input value={email} onChange={e => setEmail(e.target.value)} />
      <input type="password" value={password} onChange={e => setPassword(e.target.value)} />
      <button type="submit">Sign In</button>
    </form>
  );
}
```
