---
id: React_Cheat_Sheet
aliases:
  - React
  - React Cheat Sheet
tags:
  - coding
  - react
  - cheatsheet
dg-publish: true
---

# React Quick Cheat Sheet

> Quick reference guide for React fundamentals, modern hooks, component lifecycle, patterns, and TypeScript integration.

### 🔗 Related Vault Notes
- [[React]]: Project setup, Vite templates, and project structure
- [[Integrating JWT]]: Handling JWT authentication tokens, login forms, and API requests in React
- [[30 Days Of React]]: Structured learning plan and React drills
- [[TypeScript Cheat Sheet]]: Typing React props, hooks, generics, and events
- [[Arrow Functions|JS Arrow Functions]]: Component syntax and event handlers
- [[Asyncronous Programming|JS Async]]: Promises, `async/await` in data fetching
- [[Observer Design Pattern]]: Design pattern behind reactivity and state listeners
- [[HTTP]], [[CORS]], [[Tokens]]: Network communication and web security

---

## 1. Component Structure & JSX

```tsx
import React from 'react';

// Functional Component with TypeScript Props
interface GreetingProps {
  name: string;
  isLoggedIn?: boolean;
}

export const Greeting: React.FC<GreetingProps> = ({ name, isLoggedIn = false }) => {
  return (
    <div className="card">
      {isLoggedIn ? <h1>Welcome back, {name}!</h1> : <h1>Please log in.</h1>}
    </div>
  );
};
```

> [!NOTE]
> **JSX Rules:**
> 1. Return a single parent element (or use Fragment `<> ... </>`).
> 2. Close all tags (e.g. `<img />`, `<br />`).
> 3. Use `camelCase` for attributes (`className`, `htmlFor`, `onClick`).

---

## 2. Props & Children

> See [[TypeScript Cheat Sheet]] for deep dive into interfaces and types.

```tsx
interface CardProps {
  title: string;
  children: React.ReactNode;         // Anything renderable (JSX, string, number, null)
  onClose?: () => void;              // Optional callback function
}

export const Card = ({ title, children, onClose }: CardProps) => {
  return (
    <div className="modal-card">
      <header>
        <h3>{title}</h3>
        {onClose && <button onClick={onClose}>&times;</button>}
      </header>
      <section className="card-body">{children}</section>
    </div>
  );
};
```

---

## 3. Essential Hooks: State Management

### `useState`
```tsx
import { useState } from 'react';

// Primitive state
const [count, setCount] = useState<number>(0);

// Functional update (use when new state depends on previous state)
setCount(prev => prev + 1);

// Object / Complex state (ALWAYS create a new copy)
interface UserState {
  name: string;
  email: string;
}

const [user, setUser] = useState<UserState>({ name: '', email: '' });

// Update specific property:
setUser(prev => ({ ...prev, name: 'Arun' }));
```

### `useReducer`
> Ideal for complex state transitions or when next state depends on multiple sub-values.

```tsx
type Action = 
  | { type: 'INCREMENT'; payload?: number }
  | { type: 'DECREMENT' }
  | { type: 'RESET' };

function counterReducer(state: { count: number }, action: Action) {
  switch (action.type) {
    case 'INCREMENT':
      return { count: state.count + (action.payload ?? 1) };
    case 'DECREMENT':
      return { count: state.count - 1 };
    case 'RESET':
      return { count: 0 };
    default:
      return state;
  }
}

const [state, dispatch] = useReducer(counterReducer, { count: 0 });
// Usage: dispatch({ type: 'INCREMENT', payload: 5 });
```

### `useRef`
> Holds a mutable reference that does **not** trigger a re-render when changed. Often used to access DOM elements or persist interval IDs.

```tsx
import { useRef, useEffect } from 'react';

const inputRef = useRef<HTMLInputElement>(null);

useEffect(() => {
  // Focus the input element on mount
  inputRef.current?.focus();
}, []);

return <input ref={inputRef} placeholder="Search..." />;
```

---

## 4. Lifecycle & Side Effects (`useEffect`)

> Related: [[Asyncronous Programming]], [[Observer Design Pattern]]

```tsx
import { useEffect, useState } from 'react';

useEffect(() => {
  // 1. Code runs here on Mount (and after render if dependencies change)
  const timer = setInterval(() => {
    console.log("Tick");
  }, 1000);

  // 2. Cleanup function (runs before re-running effect & when component Unmounts)
  return () => {
    clearInterval(timer);
  };
}, []); // Empty deps array = Run ONCE on mount
```

### Dependency Array Cheat Sheet:
| Dependencies | When it runs |
| :--- | :--- |
| `undefined` (`useEffect(() => {})`) | Runs after **every** render |
| `[]` | Runs **once** after the initial render (Mount) |
| `[prop, state]` | Runs on Mount and whenever `prop` or `state` changes |

### Data Fetching in `useEffect` (Safe Pattern)
```tsx
useEffect(() => {
  let isMounted = true;

  async function fetchData() {
    try {
      const res = await fetch('/api/data');
      const data = await res.json();
      if (isMounted) setData(data);
    } catch (err) {
      if (isMounted) setError(err);
    }
  }

  fetchData();

  return () => {
    isMounted = false; // Prevent state update on unmounted component
  };
}, []);
```

---

## 5. Performance Optimization Hooks

### `useMemo`
> Memoizes the **result** of a calculation between renders.

```tsx
// Only recalculates if 'items' or 'filterTerm' change
const filteredItems = useMemo(() => {
  return items.filter(item => item.name.includes(filterTerm));
}, [items, filterTerm]);
```

### `useCallback`
> Memoizes a **function definition** between renders. Crucial when passing callbacks to memoized child components to prevent unnecessary re-renders.

```tsx
const handleDelete = useCallback((id: string) => {
  setItems(prev => prev.filter(item => item.id !== id));
}, []); // Stable reference across renders
```

### `React.memo`
```tsx
// Component will only re-render if props change
export const UserItem = React.memo(({ user, onDelete }: { user: User; onDelete: (id: string) => void }) => {
  return <div>{user.name} <button onClick={() => onDelete(user.id)}>Delete</button></div>;
});
```

---

## 6. Context API (Global State)

> For complex state, see Zustand setup in [[Web/React/React|React Architecture Note]].

```tsx
import React, { createContext, useContext, useState } from 'react';

interface AuthContextType {
  user: string | null;
  login: (name: string) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<string | null>(null);

  const login = (name: string) => setUser(name);
  const logout = () => setUser(null);

  return (
    <AuthContext.Provider value={{ user, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

// Custom Hook to consume Context safely
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};
```

---

## 7. Event Handling & Forms (Typed)

```tsx
export const LoginForm = () => {
  const [email, setEmail] = useState('');

  // Typing Input Change Event
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEmail(e.target.value);
  };

  // Typing Form Submit Event
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    console.log('Submitted email:', email);
  };

  // Typing Button Click Event
  const handleClick = (e: React.MouseEvent<HTMLButtonElement>) => {
    console.log('Button clicked at coords:', e.clientX, e.clientY);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input type="email" value={email} onChange={handleChange} required />
      <button type="submit" onClick={handleClick}>Submit</button>
    </form>
  );
};
```

---

## 8. Conditional Rendering & Lists

```tsx
// Conditional Rendering
{isLoading && <Spinner />}
{error ? <ErrorMessage message={error} /> : <DataView data={data} />}

// Rendering Lists
<ul>
  {items.map(item => (
    // ALWAYS provide a unique, stable key (don't use index if items can be reordered/deleted)
    <li key={item.id}>
      <span>{item.title}</span>
    </li>
  ))}
</ul>
```

---

## 9. Custom Hooks

> Encapsulate and share stateful logic between components. Must start with the prefix `use`.

```tsx
import { useState, useEffect } from 'react';

export function useWindowSize() {
  const [size, setSize] = useState({
    width: window.innerWidth,
    height: window.innerHeight,
  });

  useEffect(() => {
    const handleResize = () => {
      setSize({ width: window.innerWidth, height: window.innerHeight });
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  return size;
}

// In your component:
// const { width, height } = useWindowSize();
```

---

## 10. Authentication & API Integration

> See full practical example in [[Integrating JWT]].

```tsx
import { useState } from 'react';

export const useLogin = () => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const login = async (credentials: { email: string; pass: string }) => {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(credentials),
      });

      if (!res.ok) throw new Error('Authentication failed');

      const data = await res.json();
      // Store token (see [[Tokens]], [[Integrating JWT]])
      localStorage.setItem('authToken', data.token);
      return data;
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return { login, loading, error };
};
```

---

## 11. Common Pitfalls & Rules

> [!WARNING]
> 1. **Never mutate state directly:**
>    - ❌ `state.push(newItem)`
>    - ✅ `setState(prev => [...prev, newItem])`
> 2. **Rules of Hooks:**
>    - Only call hooks at the **top level** of components/custom hooks.
>    - Never call hooks inside loops, conditions, or nested functions.
> 3. **Avoid Stale Closures:**
>    - When updating state based on previous state, always use `setVal(prev => prev + 1)` instead of `setVal(val + 1)`.
> 4. **Always Clean Up Subscriptions/Timers:**
>    - Prevent memory leaks by returning a cleanup callback from `useEffect`.
